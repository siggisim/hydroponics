package s3

import (
	"context"
	"fmt"
	"io"
	"regexp"
	"sync"
	"time"

	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/aws/aws-sdk-go/service/s3/s3manager"
	"github.com/pkg/errors"
	"github.com/zenreach/hatchet"
	"github.com/zenreach/hydroponics/internal/cache"
	"github.com/zenreach/hydroponics/internal/pipes"
)

// Cache implements a cache backed by AWS S3.
type Cache struct {
	client     *s3.S3
	uploader   *s3manager.Uploader
	downloader *s3manager.Downloader
	clean      *regexp.Regexp
	bucket     string
	prefix     string
	logger     hatchet.Logger
	shutdown   chan struct{}
	wg         sync.WaitGroup
}

// New returns a new S3 cache which stores objects in the bucket with the given
// key prefix. A trailing slash is appended if one does not exist.
//
// The Get and Put methods sanitize the key value. Only alpahnumerics,
// underscores, and dashes are allowed. All other characters are replaced by
// underscores. Ensure keys match this pattern in order to avoid collisions due
// to the sanitization.
func New(bucket, prefix string, logger hatchet.Logger) (*Cache, error) {
	sesh, err := session.NewSession()
	if err != nil {
		return nil, errors.Wrap(err, "aws client")
	}

	l := len(prefix)
	if l > 0 && prefix[l-1:] != "/" {
		prefix = fmt.Sprintf("%s/", prefix)
	}

	client := s3.New(sesh)
	return &Cache{
		client:     client,
		uploader:   s3manager.NewUploaderWithClient(client),
		downloader: s3manager.NewDownloaderWithClient(client),
		clean:      regexp.MustCompile(`[^a-zA-Z0-9_-]`),
		bucket:     bucket,
		prefix:     prefix,
		logger:     logger,
		shutdown:   make(chan struct{}),
	}, nil
}

func (c *Cache) realKey(key string) string {
	key = c.clean.ReplaceAllString(key, "_")
	if c.prefix != "" {
		key = fmt.Sprintf("%s%s", c.prefix, key)
	}
	return key
}

func (c *Cache) Get(ctx context.Context, key string) (io.ReadCloser, error) {
	realKey := c.realKey(key)

	// check if the object exists
	_, err := c.client.HeadObjectWithContext(ctx, &s3.HeadObjectInput{
		Bucket: sp(c.bucket),
		Key:    sp(realKey),
	})
	if isErrCode(err, s3.ErrCodeNoSuchKey) {
		return nil, cache.ErrCacheMiss
	} else if err != nil {
		if err == ctx.Err() {
			return nil, err
		}
		return nil, errors.Wrap(err, "aws client")
	}

	c.wg.Add(2)

	// create a new context for the downloader that will be cancelled on shutdown
	var downloadCtx context.Context
	var downloadCancel context.CancelFunc
	if deadline, ok := ctx.Deadline(); ok {
		downloadCtx, downloadCancel = context.WithDeadline(context.Background(), deadline)
	} else {
		downloadCtx, downloadCancel = context.WithCancel(context.Background())
	}

	go func() {
		select {
		case <-c.shutdown:
			downloadCancel()
		case <-downloadCtx.Done():
		}
		c.wg.Done()
	}()

	// download the object concurrently
	pipe := pipes.NewBlocks()
	go func() {
		defer downloadCancel()
		_, err := c.downloader.DownloadWithContext(downloadCtx, pipe, &s3.GetObjectInput{
			Bucket: sp(c.bucket),
			Key:    sp(realKey),
		})
		if err == nil {
			pipe.Close()
		} else {
			pipe.CloseWithError(err)
		}
		c.wg.Done()
	}()
	c.touch(key)
	return &nopCloser{pipe}, nil
}

func (c *Cache) Put(ctx context.Context, key string, data io.Reader) error {
	_, err := c.uploader.UploadWithContext(ctx, &s3manager.UploadInput{
		Bucket: sp(c.bucket),
		Key:    sp(c.realKey(key)),
		Body:   data,
	})
	if err == ctx.Err() {
		return err
	}
	return errors.Wrap(err, "aws client")
}

func (c *Cache) Shutdown(ctx context.Context) error {
	close(c.shutdown)
	ch := make(chan struct{})
	go func() {
		c.wg.Wait()
		close(ch)
	}()

	select {
	case <-ch:
	case <-ctx.Done():
		return ctx.Err()
	}
	return nil
}

func (c *Cache) touch(key string) {
	c.wg.Add(1)
	go func() {
		defer c.wg.Done()
		realKey := c.realKey(key)
		source := fmt.Sprintf("/%s/%s", c.bucket, realKey)
		_, err := c.client.CopyObject(&s3.CopyObjectInput{
			Bucket:     sp(c.bucket),
			Key:        sp(realKey),
			CopySource: sp(source),
			Metadata: map[string]*string{
				"refreshed": sp(fmt.Sprintf("%d", time.Now().UTC().Unix())),
			},
			MetadataDirective: sp("REPLACE"),
		})
		if err == nil {
			c.logDebug(realKey, source, "refresh key")
		} else {
			c.logError(err, realKey, source, "key refresh error")
		}
	}()
}

func (c *Cache) logError(err error, key, source, msg string) {
	c.logger.Log(hatchet.L{
		"message": msg,
		"bucket":  c.bucket,
		"key":     key,
		"source":  source,
		"level":   "error",
		"error":   err,
	})
}

func (c *Cache) logDebug(key, source, msg string) {
	c.logger.Log(hatchet.L{
		"message": msg,
		"bucket":  c.bucket,
		"key":     key,
		"source":  source,
		"level":   "debug",
	})
}

func sp(s string) *string {
	return &s
}

func isErrCode(err error, code string) bool {
	awsErr, ok := err.(awserr.Error)
	if !ok {
		return false
	}
	return awsErr.Code() == code
}

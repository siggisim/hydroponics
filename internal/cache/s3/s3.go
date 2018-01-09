package s3

import (
	"context"
	"fmt"
	"io"
	"regexp"

	"github.com/aws/aws-sdk-go/aws/awserr"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/aws/aws-sdk-go/service/s3/s3manager"
	"github.com/pkg/errors"
	"github.com/zenreach/hydroponics/internal/cache"
)

// Cache implements a cache backed by AWS S3.
type Cache struct {
	client   *s3.S3
	uploader *s3manager.Uploader
	clean    *regexp.Regexp
	bucket   string
	prefix   string
}

// New returns a new S3 cache which stores objects in the bucket with the given
// key prefix. A trailing slash is appended if one does not exist.
//
// The Get and Put methods sanitize the key value. Only alpahnumerics,
// underscores, and dashes are allowed. All other characters are replaced by
// underscores. Ensure keys match this pattern in order to avoid collisions due
// to the sanitization.
func New(bucket, prefix string) (*Cache, error) {
	sesh, err := session.NewSession()
	if err != nil {
		return nil, errors.Wrap(err, "aws client")
	}

	l := len(prefix)
	if l > 0 && prefix[l-1:] != "/" {
		prefix = fmt.Sprintf("%s/", prefix)
	}
	return &Cache{
		client:   s3.New(sesh),
		uploader: s3manager.NewUploader(sesh),
		clean:    regexp.MustCompile(`[^a-zA-Z0-9_-]`),
		bucket:   bucket,
		prefix:   prefix,
	}, nil
}

func (c *Cache) realKey(key string) string {
	key = c.clean.ReplaceAllString(key, "_")
	if c.prefix != "" {
		key = fmt.Sprintf("%s/%s", c.prefix, key)
	}
	return key
}

func (c *Cache) Get(ctx context.Context, key string) (io.ReadCloser, error) {
	res, err := c.client.GetObjectWithContext(ctx, &s3.GetObjectInput{
		Bucket: sp(c.bucket),
		Key:    sp(c.realKey(key)),
	})
	if isErrCode(err, s3.ErrCodeNoSuchKey) {
		return nil, cache.ErrCacheMiss
	} else if err != nil {
		if err == ctx.Err() {
			return nil, err
		}
		return nil, errors.Wrap(err, "aws client")
	}
	return res.Body, nil
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

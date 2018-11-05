package httphandler

import (
	"bytes"
	"compress/gzip"
	"context"
	"io"
	"io/ioutil"
	"net/http"
	"path"
	"time"

	"github.com/zenreach/hatchet"
	"github.com/zenreach/hydroponics/internal/cache"
)

func New(cas cache.Cache, ac cache.Cache, timeout time.Duration, logger hatchet.Logger) http.Handler {
	mux := http.NewServeMux()
	mux.Handle("/cas/", &cacheHandler{
		Cache:   cas,
		Timeout: timeout,
		Logger:  logger,
	})
	mux.Handle("/ac/", &cacheHandler{
		Cache:   ac,
		Timeout: timeout,
		Logger:  logger,
	})
	return mux
}

type cacheHandler struct {
	Cache   cache.Cache
	Timeout time.Duration
	Logger  hatchet.Logger
}

func (h *cacheHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	_, key := path.Split(r.URL.Path)
	if key == "" {
		httpError(w, http.StatusNotFound)
		return
	}

	ctx := context.Background()
	if h.Timeout > 0 {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, h.Timeout)
		defer cancel()
	}

	if r.Body != nil {
		defer r.Body.Close()
	}

	switch r.Method {
	case http.MethodGet:
		rdr, err := h.Cache.Get(ctx, key)
		if err == cache.ErrCacheMiss {
			h.logDebug(key, "cache miss")
			httpError(w, http.StatusNotFound)
			return
		} else if err != nil {
			h.logError(err, key, "cache error")
			httpError(w, http.StatusInternalServerError)
			return
		}
		defer rdr.Close()

		var b1 bytes.Buffer
		_, err = io.Copy(&b1, rdr)
		if err != nil {
			h.logError(err, key, "i/o error")
			return
		}

		gzRdr, err := gzip.NewReader(&b1)
		if err != nil {
			h.logError(err, key, "gzip error")
			return
		}

		b2, err := ioutil.ReadAll(gzRdr)
		if err != nil {
			h.logError(err, key, "gzip i/o error")
			return
		}
		err = gzRdr.Close()
		if err != nil {
			h.logError(err, key, "gzip error")
			return
		}

		w.WriteHeader(http.StatusOK)
		_, err = io.Copy(w, bytes.NewBuffer(b2))
		if err != nil {
			h.logError(err, key, "write error")
			return
		}
		h.logDebug(key, "cache hit")
	case http.MethodPut:
		var b bytes.Buffer
		gzWrt, err := gzip.NewWriterLevel(&b, gzip.BestCompression)
		if err != nil {
			h.logError(err, key, "gzip error")
			return
		}

		_, err = io.Copy(gzWrt, r.Body)
		if err != nil {
			h.logError(err, key, "write error")
			return
		}
		gzWrt.Close()

		err = h.Cache.Put(ctx, key, &b)
		if err != nil {
			h.logError(err, key, "cache error")
			httpError(w, http.StatusInternalServerError)
		}
		h.logDebug(key, "cache put")
	default:
		httpError(w, http.StatusMethodNotAllowed)
	}
}

func (h *cacheHandler) logDebug(key, msg string) {
	h.Logger.Log(hatchet.L{
		"message": msg,
		"key":     key,
		"level":   "debug",
	})
}

func (h *cacheHandler) logError(err error, key, msg string) {
	h.Logger.Log(hatchet.L{
		"message": msg,
		"key":     key,
		"level":   "error",
		"error":   err,
	})
}

func httpError(w http.ResponseWriter, code int) {
	http.Error(w, http.StatusText(code), code)
}

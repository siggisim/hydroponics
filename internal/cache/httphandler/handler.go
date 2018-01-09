package httphandler

import(
	"context"
	"io"
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
			h.logInfo(key, "cache miss")
			httpError(w, http.StatusNotFound)
			return
		} else if err != nil {
			h.logError(err, key, "cache error")
			httpError(w, http.StatusInternalServerError)
			return
		}
		defer rdr.Close()

		w.WriteHeader(http.StatusOK)
		_, err = io.Copy(w, rdr)
		if err != nil {
			h.logError(err, key, "write error")
			return
		}
		h.logInfo(key, "cache hit")
	case http.MethodPut:
		err := h.Cache.Put(ctx, key, r.Body)
		if err != nil {
			h.logError(err, key, "cache error")
			httpError(w, http.StatusInternalServerError)
		}
		h.logInfo(key, "cache put")
	default:
		httpError(w, http.StatusMethodNotAllowed)
	}
}

func (h *cacheHandler) logInfo(key, msg string) {
	h.Logger.Log(hatchet.L{
		"message": msg,
		"key":     key,
		"level":   "info",
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

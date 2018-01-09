package cache

import (
	"context"
	"errors"
	"io"
)

var (
	ErrCacheMiss = errors.New("cache miss")
)

type Cache interface {
	// Get returns a reader providing access to the named cache object. An
	// ErrCacheMiss is returned if the object does not exist. Other
	// implementation specific errors may be returned. The caller must close
	// the reader when finished. If the context expires during the get
	// operation then the operation is cancelled and ctx.Err() is returned. 
	Get(context.Context, string) (io.ReadCloser, error)

	// Put caches the contents of a reader with the given name. An error is
	// returned on failure. If the context expires during the get operation
	// then the operation is cancelled and ctx.Err() is returned. 
	Put(context.Context, string, io.Reader) error
}

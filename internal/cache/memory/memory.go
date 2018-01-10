package memory

import (
	"bytes"
	"context"
	"fmt"
	"io"

	"github.com/golang/groupcache/lru"
	"github.com/zenreach/hydroponics/internal/cache"
)

// lruCache implements an in-memory LRU cache.
type lruCache struct {
	lru *lru.Cache
}

// New returns a new in-memory LRU cache which will keep up to size items.
func New(size int) cache.Cache {
	return &lruCache{
		lru: lru.New(size),
	}
}

func (c *lruCache) Get(_ context.Context, key string) (io.ReadCloser, error) {
	data, err := c.getBytes(key)
	if err != nil {
		return nil, err
	}
	return &nopCloser{bytes.NewBuffer(data)}, nil
}

func (c *lruCache) Put(_ context.Context, key string, rdr io.Reader) error {
	buf := &bytes.Buffer{}
	_, err := io.Copy(buf, rdr)
	if err != nil {
		return err
	}
	return c.putBytes(key, buf.Bytes())
}

func (c *lruCache) getBytes(key string) ([]byte, error) {
	if c.lru == nil {
		// defensive sanity check; cache was not created with New
		panic("cache lru not initialized")
	}
	iface, ok := c.lru.Get(key)
	if !ok {
		return nil, cache.ErrCacheMiss
	}
	data, ok := iface.([]byte)
	if !ok {
		// defensive sanity check; should not happen if all is implemented properly
		panic(fmt.Sprintf("lru key %s contains invalid type", key))
	}
	return data, nil
}

func (c *lruCache) putBytes(key string, data []byte) error {
	if c.lru == nil {
		// defensive sanity check; cache was not created with New
		panic("cache lru not initialized")
	}
	c.lru.Add(key, data)
	return nil
}

type nopCloser struct {
	io.Reader
}

func (nopCloser) Close() error {
	return nil
}

package memory_test

import (
	"testing"

	"github.com/zenreach/hydroponics/internal/cache"
	"github.com/zenreach/hydroponics/internal/cache/cachetest"
	"github.com/zenreach/hydroponics/internal/cache/memory"
)

func TestCommon(t *testing.T) {
	cachetest.Test(t, func() cache.Cache {
		return memory.New(100)
	})
}

func TestExpire(t *testing.T) {
	c := memory.New(2)

	// add values
	cachetest.AssertPut(t, c, "key1", []byte("value1"))
	cachetest.AssertPut(t, c, "key2", []byte("value2"))

	// access a key to ensure it stays in the cache
	cachetest.AssertGet(t, c, "key1", []byte("value1"))

	// add a value to evict a key
	cachetest.AssertPut(t, c, "key3", []byte("value3"))

	// verify cache contents
	cachetest.AssertGet(t, c, "key1", []byte("value1"))
	cachetest.AssertMiss(t, c, "key2")
	cachetest.AssertGet(t, c, "key3", []byte("value3"))
}

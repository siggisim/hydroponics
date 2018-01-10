package cachetest

import (
	"bytes"
	"context"
	"io"
	"io/ioutil"
	"reflect"
	"testing"
	"time"

	"github.com/zenreach/hydroponics/internal/cache"
)

// Test a cache implementation.
func Test(t *testing.T, factory func() cache.Cache) {
	t.Parallel()
	tests := map[string]func(*testing.T, cache.Cache){
		"get hit":      testGetHit,
		"get miss":     testGetMiss,
		"put existing": testPutExisting,
	}

	for name := range tests {
		test := tests[name]
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			test(t, factory())
		})
	}
}

func testGetHit(t *testing.T, c cache.Cache) {
	key := "hit"
	data := []byte("example cache value")
	AssertPut(t, c, key, data)
	AssertGet(t, c, key, data)
}

func testGetMiss(t *testing.T, c cache.Cache) {
	AssertMiss(t, c, "missing")
}

func testPutExisting(t *testing.T, c cache.Cache) {
	key := "exists"
	data1 := []byte("replace this value")
	data2 := []byte("a replacement value")
	AssertPut(t, c, key, data1)
	AssertGet(t, c, key, data1)
	AssertPut(t, c, key, data2)
	AssertGet(t, c, key, data2)
}

func AssertGet(t *testing.T, c cache.Cache, key string, want []byte) {
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	rdr, err := c.Get(ctx, key)
	if err != nil {
		t.Fatalf("failed to get value: %s")
	}
	have := ReadAll(t, rdr)
	if !reflect.DeepEqual(have, want) {
		t.Errorf("expected value \"%s\", got \"%s\"", have, want)
	}
}

func AssertPut(t *testing.T, c cache.Cache, key string, have []byte) {
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	err := c.Put(ctx, key, NewReader(have))
	if err != nil {
		t.Fatalf("failed to put value: %s")
	}
}

func AssertMiss(t *testing.T, c cache.Cache, key string) {
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	rdr, err := c.Get(ctx, key)
	if err != cache.ErrCacheMiss {
		t.Errorf("expected \"%s\", got \"%s\"", cache.ErrCacheMiss, err)
	}
	if rdr != nil {
		t.Error("expected nil reader")
	}
}

func NewReader(data []byte) io.ReadCloser {
	return ioutil.NopCloser(bytes.NewBuffer(data))
}

func ReadAll(t *testing.T, reader io.ReadCloser) []byte {
	data, err := ioutil.ReadAll(reader)
	if err != nil {
		t.Errorf("failed to read data: %s", err)
	}
	err = reader.Close()
	if err != nil {
		t.Errorf("failed to close reader: %s", err)
	}
	return data
}

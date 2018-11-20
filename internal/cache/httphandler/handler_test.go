package httphandler_test

import (
	"bytes"
	"compress/gzip"
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"reflect"
	"testing"
	"time"

	"github.com/zenreach/hatchet"
	"github.com/zenreach/hydroponics/internal/cache"
	"github.com/zenreach/hydroponics/internal/cache/cachetest"
	"github.com/zenreach/hydroponics/internal/cache/httphandler"
	"github.com/zenreach/hydroponics/internal/cache/memory"
)

type service struct {
	Name  string
	Cache cache.Cache
}

type testEnv struct {
	*testing.T
	CAS    *service
	AC     *service
	Client *http.Client
	Server *httptest.Server
}

func Setup(t *testing.T) *testEnv {
	t.Parallel()
	te := &testEnv{
		T: t,
		CAS: &service{
			Name:  "cas",
			Cache: memory.New(10),
		},
		AC: &service{
			Name:  "ac",
			Cache: memory.New(10),
		},
		Client: &http.Client{},
	}
	handler := httphandler.New(te.CAS.Cache, te.AC.Cache, 15*time.Second, hatchet.Test(t))
	te.Server = httptest.NewServer(handler)
	return te
}

func (te *testEnv) Teardown() {
	te.Server.Close()
}

func (te *testEnv) Services() []*service {
	return []*service{te.CAS, te.AC}
}

func (te *testEnv) URL(svc *service, key string) string {
	return fmt.Sprintf("%s/%s/%s", te.Server.URL, svc.Name, key)
}

func (te *testEnv) Get(svc *service, key string) *http.Response {
	res, err := te.Client.Get(te.URL(svc, key))
	if err != nil {
		te.Fatalf("client error: %s", err)
	}
	return res
}

func (te *testEnv) GetValue(svc *service, key string) []byte {
	res := te.Get(svc, key)
	if res.Body != nil {
		defer res.Body.Close()
	}
	if res.StatusCode != http.StatusOK {
		te.Errorf("expected 404, got %d", http.StatusNotFound)
		return nil
	}
	if res.Body != nil {
		return cachetest.ReadAll(te.T, res.Body)
	}
	return nil
}

func (te *testEnv) Put(svc *service, key string, value []byte) *http.Response {
	uri, err := url.Parse(te.URL(svc, key))
	if err != nil {
		te.Fatalf("uri error: %s", err)
	}
	res, err := te.Client.Do(&http.Request{
		Method: http.MethodPut,
		URL:    uri,
		Body:   cachetest.NewReader(value),
	})
	if err != nil {
		te.Fatalf("client error: %s", err)
	}
	return res
}

func (te *testEnv) TestEach(test func(*testEnv, *service)) {
	services := te.Services()
	for i := range services {
		svc := services[i]
		te.Run(svc.Name, func(t *testing.T) {
			test(te, svc)
		})
	}
}

func TestGetHit(t *testing.T) {
	te := Setup(t)
	te.TestEach(testGetHit)
}

func testGetHit(t *testEnv, svc *service) {
	key := "exists"
	want := []byte("existing value")
	wantCmp := compress(want)

	// load value into cache
	cachetest.AssertPut(t.T, svc.Cache, key, wantCmp)

	// retrieve it via the handler
	have := t.GetValue(svc, key)
	if !reflect.DeepEqual(have, want) {
		t.Errorf("expected value \"%s\", got \"%s\"", want, have)
	}
}

func TestGetMiss(t *testing.T) {
	te := Setup(t)
	te.TestEach(testGetMiss)
}

func testGetMiss(t *testEnv, svc *service) {
	key := "missing"
	res := t.Get(svc, key)
	if res.StatusCode != http.StatusNotFound {
		t.Errorf("expected status code %d, got %d", http.StatusNotFound, res.StatusCode)
	}
}

func TestPutNew(t *testing.T) {
	te := Setup(t)
	te.TestEach(testPutNew)
}

func testPutNew(t *testEnv, svc *service) {
	key := "new"
	value := []byte("new value")
	valueCmp := compress(value)

	// put value via the handler
	res := t.Put(svc, key, value)
	if res.StatusCode != http.StatusOK {
		t.Fatalf("expected status code %d, got %d", http.StatusOK, res.StatusCode)
	}

	// verify in cache
	cachetest.AssertGet(t.T, svc.Cache, key, valueCmp)
}

func TestPutExisting(t *testing.T) {
	te := Setup(t)
	te.TestEach(testPutExisting)
}

func testPutExisting(t *testEnv, svc *service) {
	key := "new"
	oldvalue := []byte("existing value")
	oldvalueCmp := compress(oldvalue)

	// load value into cache
	cachetest.AssertPut(t.T, svc.Cache, key, oldvalueCmp)

	// put value via the handler
	newvalue := []byte("new value")
	newvalueCmp := compress(newvalue)
	res := t.Put(svc, key, newvalue)
	if res.StatusCode != http.StatusOK {
		t.Fatalf("expected status code %d, got %d", http.StatusOK, res.StatusCode)
	}

	// verify in cache
	cachetest.AssertGet(t.T, svc.Cache, key, newvalueCmp)
}

func compress(value []byte) []byte {
	var buf bytes.Buffer
	gzipper, _ := gzip.NewWriterLevel(&buf, gzip.BestCompression)
	gzipper.Write(value)
	gzipper.Close()
	return buf.Bytes()
}

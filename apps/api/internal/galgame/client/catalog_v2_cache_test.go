package client

import (
	"bytes"
	"context"
	"net/http"
	"net/http/httptest"
	"net/url"
	"sync"
	"testing"
	"time"

	"github.com/alicebob/miniredis/v2"
	"github.com/redis/go-redis/v9"
)

type v2Origin struct {
	mu     sync.Mutex
	hits   int
	status int
	body   []byte
}

func (o *v2Origin) ServeHTTP(w http.ResponseWriter, _ *http.Request) {
	o.mu.Lock()
	o.hits++
	status, body := o.status, o.body
	o.mu.Unlock()
	w.WriteHeader(status)
	_, _ = w.Write(body)
}

func (o *v2Origin) n() int {
	o.mu.Lock()
	defer o.mu.Unlock()
	return o.hits
}

func newV2CacheFixture(t *testing.T, redisOn bool) (*GalgameClient, *miniredis.Miniredis, *v2Origin) {
	t.Helper()
	o := &v2Origin{status: http.StatusOK, body: []byte(`{"ok":true}`)}
	srv := httptest.NewServer(o)
	t.Cleanup(srv.Close)
	c := New(srv.URL, "nmk_test", "")
	var mr *miniredis.Miniredis
	if redisOn {
		mr = miniredis.RunT(t)
		rdb := redis.NewClient(&redis.Options{Addr: mr.Addr()})
		t.Cleanup(func() { _ = rdb.Close() })
		c.WithRedis(rdb)
	}
	return c, mr, o
}

func doV2CacheKey(path string, query url.Values) string {
	return v2CacheKey(v2CatalogPath(path), v2CatalogQuery(path, query))
}

func TestV2CacheMissStoresAndHitServes(t *testing.T) {
	c, mr, o := newV2CacheFixture(t, true)
	ctx := context.Background()
	path := "/catalog/works"
	key := doV2CacheKey(path, nil)

	_, body1, appErr := c.doV2(ctx, http.MethodGet, path, nil, nil)
	if appErr != nil {
		t.Fatalf("doV2 miss: %v", appErr)
	}
	if o.n() != 1 {
		t.Fatalf("origin hits after miss = %d, want 1", o.n())
	}
	stored, err := mr.Get(key)
	if err != nil {
		t.Fatalf("miss did not store %s: %v", key, err)
	}
	if stored != `{"ok":true}` {
		t.Fatalf("stored %q, want origin bytes", stored)
	}

	_, body2, appErr := c.doV2(ctx, http.MethodGet, path, nil, nil)
	if appErr != nil {
		t.Fatalf("doV2 hit: %v", appErr)
	}
	if o.n() != 1 {
		t.Fatalf("origin hits after cache hit = %d, want 1", o.n())
	}
	if !bytes.Equal(body1, body2) {
		t.Fatalf("hit body %q != miss body %q", body2, body1)
	}
}

func TestV2RawCacheMissStoresAndHitServes(t *testing.T) {
	c, mr, o := newV2CacheFixture(t, true)
	ctx := context.Background()
	v2Path := "/v2/catalog/works/123/covers"
	key := v2CacheKey(v2Path, nil)

	body1, appErr := c.getV2Raw(ctx, v2Path, nil)
	if appErr != nil {
		t.Fatalf("getV2Raw miss: %v", appErr)
	}
	if o.n() != 1 {
		t.Fatalf("origin hits after miss = %d, want 1", o.n())
	}
	stored, err := mr.Get(key)
	if err != nil {
		t.Fatalf("miss did not store %s: %v", key, err)
	}
	if stored != `{"ok":true}` {
		t.Fatalf("stored %q, want origin bytes", stored)
	}

	body2, appErr := c.getV2Raw(ctx, v2Path, nil)
	if appErr != nil {
		t.Fatalf("getV2Raw hit: %v", appErr)
	}
	if o.n() != 1 {
		t.Fatalf("origin hits after cache hit = %d, want 1", o.n())
	}
	if !bytes.Equal(body1, body2) {
		t.Fatalf("hit body %q != miss body %q", body2, body1)
	}
}

func TestV2CacheSharedByDoV2AndGetV2Raw(t *testing.T) {
	c, _, o := newV2CacheFixture(t, true)
	ctx := context.Background()
	_, _, appErr := c.doV2(ctx, http.MethodGet, "/catalog/works", nil, nil)
	if appErr != nil {
		t.Fatalf("doV2: %v", appErr)
	}
	raw, appErr := c.getV2Raw(ctx, "/v2/catalog/works", v2CatalogQuery("/catalog/works", nil))
	if appErr != nil {
		t.Fatalf("getV2Raw: %v", appErr)
	}
	if o.n() != 1 {
		t.Fatalf("shared cache missed, origin hits = %d, want 1", o.n())
	}
	if string(raw) != `{"ok":true}` {
		t.Fatalf("raw = %q", raw)
	}
}

func TestV2CacheSkipsNon200(t *testing.T) {
	c, mr, o := newV2CacheFixture(t, true)
	o.status = http.StatusNotFound
	o.body = []byte(`{"err":true}`)
	ctx := context.Background()
	path := "/catalog/works"
	key := doV2CacheKey(path, nil)

	status, _, appErr := c.doV2(ctx, http.MethodGet, path, nil, nil)
	if appErr != nil {
		t.Fatalf("doV2 404: %v", appErr)
	}
	if status != http.StatusNotFound {
		t.Fatalf("status = %d, want 404", status)
	}
	if _, err := mr.Get(key); err == nil {
		t.Fatalf("non-200 was cached: %s", key)
	}
	if _, _, appErr = c.doV2(ctx, http.MethodGet, path, nil, nil); appErr != nil {
		t.Fatalf("doV2 404 second: %v", appErr)
	}
	if o.n() != 2 {
		t.Fatalf("doV2 origin hits after two 404s = %d, want 2", o.n())
	}

	_, appErr = c.getV2Raw(ctx, "/v2/catalog/works/1", nil)
	if appErr == nil {
		t.Fatal("getV2Raw 404: want error")
	}
	if _, err := mr.Get(v2CacheKey("/v2/catalog/works/1", nil)); err == nil {
		t.Fatal("getV2Raw non-200 was cached")
	}
	if o.n() != 3 {
		t.Fatalf("origin hits = %d, want 3", o.n())
	}
}

func TestV2CacheSkipsLargeBody(t *testing.T) {
	c, mr, o := newV2CacheFixture(t, true)
	o.body = bytes.Repeat([]byte("a"), v2CacheMaxBody+1)
	ctx := context.Background()
	path := "/catalog/works"
	key := doV2CacheKey(path, nil)

	_, _, appErr := c.doV2(ctx, http.MethodGet, path, nil, nil)
	if appErr != nil {
		t.Fatalf("doV2 large: %v", appErr)
	}
	if _, err := mr.Get(key); err == nil {
		t.Fatalf("body >512KiB was cached: %s", key)
	}

	_, _, appErr = c.doV2(ctx, http.MethodGet, path, nil, nil)
	if appErr != nil {
		t.Fatalf("doV2 large second: %v", appErr)
	}
	if o.n() != 2 {
		t.Fatalf("origin hits = %d, want 2", o.n())
	}
}

func TestV2CacheTTLClasses(t *testing.T) {
	c, mr, _ := newV2CacheFixture(t, true)
	ctx := context.Background()

	if _, _, appErr := c.doV2(ctx, http.MethodGet, "/catalog/works", nil, nil); appErr != nil {
		t.Fatalf("list: %v", appErr)
	}
	if _, _, appErr := c.doV2(ctx, http.MethodGet, "/catalog/works/9", nil, nil); appErr != nil {
		t.Fatalf("detail: %v", appErr)
	}
	if _, appErr := c.getV2Raw(ctx, "/v2/catalog/works/9/covers", nil); appErr != nil {
		t.Fatalf("subresource: %v", appErr)
	}

	listKey := doV2CacheKey("/catalog/works", nil)
	detailKey := doV2CacheKey("/catalog/works/9", nil)
	subKey := v2CacheKey("/v2/catalog/works/9/covers", nil)
	if got := mr.TTL(listKey); got != v2CacheListTTL {
		t.Fatalf("list TTL = %v, want %v", got, v2CacheListTTL)
	}
	if got := mr.TTL(detailKey); got != v2CacheDetailTTL {
		t.Fatalf("detail TTL = %v, want %v", got, v2CacheDetailTTL)
	}
	if got := mr.TTL(subKey); got != v2CacheDetailTTL {
		t.Fatalf("subresource TTL = %v, want %v", got, v2CacheDetailTTL)
	}
}

func TestV2CacheNilRedisDisabled(t *testing.T) {
	c, _, o := newV2CacheFixture(t, false)
	ctx := context.Background()
	if _, _, appErr := c.doV2(ctx, http.MethodGet, "/catalog/works", nil, nil); appErr != nil {
		t.Fatalf("first: %v", appErr)
	}
	if _, _, appErr := c.doV2(ctx, http.MethodGet, "/catalog/works", nil, nil); appErr != nil {
		t.Fatalf("second: %v", appErr)
	}
	if o.n() != 2 {
		t.Fatalf("nil rdb origin hits = %d, want 2", o.n())
	}
}

func TestV2CacheRedisDownFallthrough(t *testing.T) {
	c, mr, o := newV2CacheFixture(t, true)
	mr.Close()
	ctx := context.Background()
	if _, _, appErr := c.doV2(ctx, http.MethodGet, "/catalog/works", nil, nil); appErr != nil {
		t.Fatalf("first with redis down: %v", appErr)
	}
	if _, _, appErr := c.doV2(ctx, http.MethodGet, "/catalog/works", nil, nil); appErr != nil {
		t.Fatalf("second with redis down: %v", appErr)
	}
	if o.n() != 2 {
		t.Fatalf("redis-down origin hits = %d, want 2", o.n())
	}
}

func TestV2CacheTTL(t *testing.T) {
	cases := []struct {
		path string
		ttl  time.Duration
	}{
		{"/v2/catalog/works", v2CacheListTTL},
		{"/v2/catalog/works/123", v2CacheDetailTTL},
		{"/v2/catalog/works/123/covers", v2CacheDetailTTL},
		{"/v2/catalog/calendar", v2CacheListTTL},
		{"/v2/catalog/search", v2CacheListTTL},
		{"/v2/catalog/companies", v2CacheListTTL},
		{"/v2/catalog/companies/5/graph", v2CacheDetailTTL},
	}
	for _, c := range cases {
		if got := v2CacheTTL(c.path); got != c.ttl {
			t.Errorf("v2CacheTTL(%q) = %v, want %v", c.path, got, c.ttl)
		}
	}
}

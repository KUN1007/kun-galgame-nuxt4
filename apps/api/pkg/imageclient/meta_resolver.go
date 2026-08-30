package imageclient

import (
	"context"
	"sync"
	"time"
)

// ungradedTTL bounds how long a cached entry may keep saying "not graded yet".
// Width, height and thumbhash are content-addressed and never change, so they
// are cached for the life of the process; the sexual grade appears the night
// after the upload, and without an expiry the process would keep serving an
// image as ungraded until it restarted.
const ungradedTTL = time.Hour

type cachedMeta struct {
	meta    ImageMeta
	expires time.Time
}

type MetaResolver struct {
	client   *Client
	timeout  time.Duration
	gradeTTL time.Duration
	mu       sync.RWMutex
	cache    map[string]cachedMeta
}

func (c *Client) NewMetaResolver(timeout time.Duration) *MetaResolver {
	if timeout <= 0 {
		timeout = 3 * time.Second
	}
	return &MetaResolver{
		client: c, timeout: timeout, gradeTTL: ungradedTTL,
		cache: map[string]cachedMeta{},
	}
}

func (r *MetaResolver) Resolve(hashes []string) map[string]ImageMeta {
	out := make(map[string]ImageMeta, len(hashes))
	var miss []string

	now := time.Now()
	r.mu.RLock()
	for _, h := range hashes {
		if c, ok := r.cache[h]; ok && (c.expires.IsZero() || c.expires.After(now)) {
			out[h] = c.meta
		} else {
			miss = append(miss, h)
		}
	}
	r.mu.RUnlock()

	if len(miss) == 0 || !r.client.Configured() {
		return out
	}

	ctx, cancel := context.WithTimeout(context.Background(), r.timeout)
	defer cancel()
	fetched, err := r.client.MetaBatch(ctx, dedupHashes(miss))
	if err != nil {
		return out
	}

	r.mu.Lock()
	for h, m := range fetched {
		out[h] = m
		if m.Thumbhash == "" {
			continue
		}
		entry := cachedMeta{meta: m}
		if m.Sexual == nil {
			entry.expires = now.Add(r.gradeTTL)
		}
		r.cache[h] = entry
	}
	r.mu.Unlock()
	return out
}

func dedupHashes(in []string) []string {
	if len(in) < 2 {
		return in
	}
	seen := make(map[string]struct{}, len(in))
	out := make([]string, 0, len(in))
	for _, s := range in {
		if _, ok := seen[s]; ok {
			continue
		}
		seen[s] = struct{}{}
		out = append(out, s)
	}
	return out
}

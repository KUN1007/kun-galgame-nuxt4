package client

import (
	"net/url"
	"testing"
)

func TestV2CatalogQuery_CalendarBuckets(t *testing.T) {
	cases := []struct {
		path   string
		want   map[string]string
		absent []string
	}{
		{
			path:   "/catalog/calendar",
			want:   map[string]string{"include_total": "true"},
			absent: []string{"precision", "status"},
		},
		{
			// status=unknown lands on the SAME undated bucket as tba upstream;
			// the year-known bucket is only reachable through precision.
			path:   "/catalog/calendar/pending",
			want:   map[string]string{"precision": "year", "include_total": "true"},
			absent: []string{"status"},
		},
		{
			path:   "/catalog/calendar/tba",
			want:   map[string]string{"status": "unknown", "include_total": "true"},
			absent: []string{"precision"},
		},
	}
	for _, c := range cases {
		t.Run(c.path, func(t *testing.T) {
			q := v2CatalogQuery(c.path, url.Values{"limit": {"100"}})
			for k, v := range c.want {
				if q.Get(k) != v {
					t.Errorf("%s = %q, want %q (%v)", k, q.Get(k), v, q)
				}
			}
			for _, k := range c.absent {
				if q.Get(k) != "" {
					t.Errorf("%s = %q, want it unset (%v)", k, q.Get(k), q)
				}
			}
		})
	}
}

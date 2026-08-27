package client

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"slices"
	"strings"
	"testing"
)

func TestClaimSiteAcceptedOnBothSpellings(t *testing.T) {
	cases := []struct {
		site string
		want int
	}{
		{"kungal", 4321},
		{"galgame_wiki", 4321},
		{"moyu", 0},
		{"", 0},
	}
	for _, tc := range cases {
		t.Run(tc.site, func(t *testing.T) {
			it := &CatalogWorkListItem{ClaimedBy: &catClaimedBy{Site: tc.site, WorkID: 4321}}
			if got := it.gid(); got != tc.want {
				t.Errorf("gid() on site %q = %d, want %d", tc.site, got, tc.want)
			}
		})
	}
	if got := (&CatalogWorkListItem{ID: 14}).gid(); got != 14 {
		t.Errorf("unclaimed gid() = %d, want the catalog work id", got)
	}
	if (&CatalogWorkListItem{}).gid() != 0 {
		t.Error("a row with no catalog id has no gid")
	}
}

func TestAnchorLookupAsksForEverySourceKey(t *testing.T) {
	var asked []string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		var items []string
		for _, token := range strings.Split(req.URL.Query().Get("refs"), ",") {
			source, ext := token, ""
			if i := strings.LastIndex(token, ":"); i >= 0 {
				source, ext = token[:i], token[i+1:]
			}
			asked = append(asked, source)
			if source == "galgame_wiki" && ext == "7" {
				items = append(items, `{"id":"9001","refs":[{"source":"galgame_wiki","external_id":"7"}]}`)
			}
		}
		_, _ = w.Write([]byte(`{"object":"list","items":[` + strings.Join(items, ",") + `],"missing":[]}`))
	}))
	t.Cleanup(srv.Close)

	c := New(srv.URL, "nm_test_key", "")
	ids, appErr := c.catalogIDsForGIDs(t.Context(), []int{7})
	if appErr != nil {
		t.Fatalf("catalogIDsForGIDs: %v", appErr)
	}
	if ids[7] != 9001 {
		t.Errorf("gid 7 resolved to %d, want 9001 via the pre-rename source key", ids[7])
	}
	for _, key := range anchorSourceKeys {
		if !slices.Contains(asked, key) {
			t.Errorf("source key %q was never asked for; the lookup only resolves "+
				"on the side of the rename it happens to name", key)
		}
	}
}

func TestAnchorLookupResolvesAfterTheRename(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		var items []string
		for _, token := range strings.Split(req.URL.Query().Get("refs"), ",") {
			source, ext := token, ""
			if i := strings.LastIndex(token, ":"); i >= 0 {
				source, ext = token[:i], token[i+1:]
			}
			if source == "curated" && ext == "8" {
				items = append(items, `{"id":"9002","refs":[{"source":"curated","external_id":"8"}]}`)
			}
		}
		_, _ = w.Write([]byte(`{"object":"list","items":[` + strings.Join(items, ",") + `],"missing":[]}`))
	}))
	t.Cleanup(srv.Close)

	c := New(srv.URL, "nm_test_key", "")
	ids, appErr := c.catalogIDsForGIDs(t.Context(), []int{8})
	if appErr != nil {
		t.Fatalf("catalogIDsForGIDs: %v", appErr)
	}
	if ids[8] != 9002 {
		t.Errorf("gid 8 resolved to %d, want 9002 via the post-rename source key", ids[8])
	}
}

func adoptedStub(t *testing.T, rows map[int64]string) *GalgameClient {
	t.Helper()
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		switch {
		case strings.HasSuffix(req.URL.Path, "/catalog/lookup/batch"):
			var body struct {
				Items []struct {
					ExternalID string `json:"external_id"`
				} `json:"items"`
			}
			_ = json.NewDecoder(req.Body).Decode(&body)
			out := make([]string, 0, len(body.Items))
			for _, it := range body.Items {
				out = append(out, `{"external_id":"`+it.ExternalID+`","work":null}`)
			}
			_, _ = w.Write([]byte(`{"object":"list","items":[` +
				strings.Join(out, ",") + `]}`))
		case strings.HasSuffix(req.URL.Path, "/catalog/works"):
			var items []string
			for _, raw := range strings.Split(req.URL.Query().Get("ids"), ",") {
				var id int64
				for _, r := range strings.TrimSpace(raw) {
					if r >= '0' && r <= '9' {
						id = id*10 + int64(r-'0')
					}
				}
				if frag, ok := rows[id]; ok {
					items = append(items, frag)
				}
			}
			_, _ = w.Write([]byte(`{"object":"list","items":[` +
				strings.Join(items, ",") + `],"next_cursor":null}`))
		default:
			w.WriteHeader(http.StatusNotFound)
		}
	}))
	t.Cleanup(srv.Close)
	return New(srv.URL, "nm_test_key", "")
}

func TestAdoptedIDResolvesWithoutAnAnchor(t *testing.T) {
	c := adoptedStub(t, map[int64]string{
		90210: `{"id":90210,"claimed_by":{"site":"kungal","work_id":90210,"state":"pending"}}`,
	})
	ids, appErr := c.catalogIDsForGIDs(t.Context(), []int{90210})
	if appErr != nil {
		t.Fatalf("catalogIDsForGIDs: %v", appErr)
	}
	if ids[90210] != 90210 {
		t.Errorf("adopted id resolved to %d, want 90210 — a submission with no "+
			"anchor must still reach its own page", ids[90210])
	}
}

func TestIdentityRouteRefusesAWorkThatNamesSomethingElse(t *testing.T) {
	c := adoptedStub(t, map[int64]string{
		42: `{"id":42,"claimed_by":{"site":"kungal","work_id":7,"state":"live"}}`,
	})
	ids, appErr := c.catalogIDsForGIDs(t.Context(), []int{42})
	if appErr != nil {
		t.Fatalf("catalogIDsForGIDs: %v", appErr)
	}
	if _, found := ids[42]; found {
		t.Errorf("gid 42 resolved to work %d — the round-trip check must reject a "+
			"work whose claim names another id", ids[42])
	}
}

func TestIdentityRouteResolvesAForeignClaimByCatalogID(t *testing.T) {
	c := adoptedStub(t, map[int64]string{
		500: `{"id":500,"claimed_by":{"site":"moyu","work_id":500,"state":"live"}}`,
	})
	ids, appErr := c.catalogIDsForGIDs(t.Context(), []int{500})
	if appErr != nil {
		t.Fatalf("catalogIDsForGIDs: %v", appErr)
	}
	if ids[500] != 500 {
		t.Errorf("foreign-claimed work 500 resolved to %d, want catalog id 500", ids[500])
	}
}

func TestIdentityRouteResolvesAnUnclaimedWorkByCatalogID(t *testing.T) {
	c := adoptedStub(t, map[int64]string{
		600: `{"id":600,"claimed_by":null}`,
	})
	ids, appErr := c.catalogIDsForGIDs(t.Context(), []int{600})
	if appErr != nil {
		t.Fatalf("catalogIDsForGIDs: %v", appErr)
	}
	if ids[600] != 600 {
		t.Errorf("unclaimed work 600 resolved to %d, want catalog id 600", ids[600])
	}
}

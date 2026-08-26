package client

import (
	"context"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"

	"kun-galgame-api/internal/galgame/dto"
)

func labelCatalog(t *testing.T) (*GalgameClient, func(string) int) {
	t.Helper()
	var mu sync.Mutex
	hits := map[string]int{}
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		mu.Lock()
		hits[r.URL.Path]++
		mu.Unlock()
		w.Header().Set("Content-Type", "application/json")
		switch r.URL.Path {
		case "/v1/catalog/labels/107":
			_, _ = w.Write([]byte(`{"code":0,"message":"成功","data":{"id":107,` +
				`"display_name":"Purple SOFTWARE","links":[` +
				`{"source":"twitter","url":"https://x.com/purplesoftware"},` +
				`{"source":"official_site","url":"https://www.purplesoftware.jp"}]}}`))
		case "/v1/catalog/labels/208":
			_, _ = w.Write([]byte(`{"code":0,"message":"成功","data":{"id":208,` +
				`"display_name":"同人サークル","links":[` +
				`{"source":"twitter","url":"https://x.com/doujin_circle"},` +
				`{"source":"cien","url":"https://ci-en.dlsite.com/creator/12345"}]}}`))
		case "/v1/catalog/labels/309":
			_, _ = w.Write([]byte(`{"code":0,"message":"成功","data":{"id":309,"display_name":"无官网社","links":[]}}`))
		default:
			w.WriteHeader(http.StatusNotFound)
			_, _ = w.Write([]byte(`{"code":4,"message":"资源不存在"}`))
		}
	}))
	t.Cleanup(srv.Close)
	return New(srv.URL, "test-key", ""), func(path string) int {
		mu.Lock()
		defer mu.Unlock()
		return hits[path]
	}
}

func detailWithLabels(ids ...int) dto.NextMoeGalgameDetailFull {
	var g dto.NextMoeGalgameDetailFull
	for _, id := range ids {
		g.Official = append(g.Official, dto.NextMoeOfficialRel{
			Official: dto.NextMoeOfficial{ID: id},
		})
	}
	return g
}

func TestHydrateOfficialLinksPrefersTheOfficialSite(t *testing.T) {
	c, _ := labelCatalog(t)
	g := detailWithLabels(107, 208, 309, 999)

	c.HydrateOfficialLinks(context.Background(), &g)

	want := []string{
		"https://www.purplesoftware.jp",
		"",
		"",
		"",
	}
	for i, w := range want {
		if got := g.Official[i].Official.Link; got != w {
			t.Fatalf("official[%d].link = %q, want %q", i, got, w)
		}
	}
}

func TestHydrateOfficialLinksMemoizesPerLabel(t *testing.T) {
	c, hits := labelCatalog(t)
	for range 2 {
		g := detailWithLabels(107, 309)
		c.HydrateOfficialLinks(context.Background(), &g)
		if g.Official[0].Official.Link != "https://www.purplesoftware.jp" {
			t.Fatalf("second view lost the link: %q", g.Official[0].Official.Link)
		}
	}
	if n := hits("/v1/catalog/labels/107"); n != 1 {
		t.Fatalf("label 107 fetched %d times, want 1", n)
	}
	if n := hits("/v1/catalog/labels/309"); n != 1 {
		t.Fatalf("linkless label 309 fetched %d times, want 1 (negative not cached)", n)
	}
}

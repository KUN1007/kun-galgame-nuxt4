package client

import (
	"context"
	"encoding/json"
	"net/url"
	"os"
	"strings"
	"testing"
)

func TestRewriteV2JSON_CoverSlotsDoNotRecurse(t *testing.T) {
	raw := []byte(`{"object":"work","id":"4242","display_name":"Kun","cover":{"url":"https://cdn.example/p.webp","hash":"aa","width":600,"height":800},"banner":{"url":"https://cdn.example/b.webp","width":1280,"height":720,"thumbhash":"B"},"covers":[{"url":"https://cdn.example/p.webp","kind":"main"}],"cover_slots":{"portrait":{"url":"https://cdn.example/p.webp","width":600,"height":800},"banner":{"url":"https://cdn.example/b.webp","width":1280,"height":720,"thumbhash":"B"}}}`)
	out := rewriteV2JSON(raw, "")
	var got map[string]any
	if err := json.Unmarshal(out, &got); err != nil {
		t.Fatalf("rewrite produced invalid json: %v\n%s", err, out)
	}
	slots, _ := got["cover_slots"].(map[string]any)
	if slots == nil {
		t.Fatalf("cover_slots missing: %s", out)
	}
	if _, nested := slots["cover_slots"]; nested {
		t.Fatalf("cover_slots was rewritten as a work: %s", out)
	}
	covers, ok := got["covers"].([]any)
	if !ok || len(covers) != 1 {
		t.Fatalf("covers array was overwritten: %s", out)
	}
}

func TestRewriteV2JSON_CoverAndBannerBecomeSlots(t *testing.T) {
	raw := []byte(`{"object":"work","id":"1","display_name":"Kun","cover":{"url":"https://cdn.example/p.webp","width":600,"height":800,"thumbhash":"P"},"banner":{"url":"https://cdn.example/b.webp","width":1280,"height":720,"thumbhash":"B"}}`)
	out := rewriteV2JSON(raw, "")
	var got map[string]any
	if err := json.Unmarshal(out, &got); err != nil {
		t.Fatalf("rewrite produced invalid json: %v", err)
	}
	if _, ok := got["cover"].(string); !ok {
		t.Fatalf("cover Image was not flattened to a URL: %s", out)
	}
	slots, _ := got["cover_slots"].(map[string]any)
	banner, _ := slots["banner"].(map[string]any)
	if banner["url"] != "https://cdn.example/b.webp" || banner["thumbhash"] != "B" {
		t.Fatalf("banner slot = %v", banner)
	}
}

func TestV2CatalogQuery_EmptySearchGetsFacets(t *testing.T) {
	q := v2CatalogQuery("/catalog/works/search", url.Values{"limit": {"24"}})
	if q.Get("facets") != "olang,tag_id" {
		t.Fatalf("facets = %q, want olang,tag_id so empty browse uses Meili total", q.Get("facets"))
	}
	if q.Get("include_total") != "true" {
		t.Fatalf("include_total = %q", q.Get("include_total"))
	}
}

func TestV2CatalogQuery_PageBecomesCursor(t *testing.T) {
	q := v2CatalogQuery("/catalog/works/search", url.Values{"q": {"kun"}, "page": {"2"}})
	if q.Get("page") != "" {
		t.Fatalf("page leaked: %q", q.Get("page"))
	}
	if !strings.HasPrefix(q.Get("cursor"), "cur_") {
		t.Fatalf("cursor = %q, want cur_…", q.Get("cursor"))
	}
}

func TestLiveV2Adapter(t *testing.T) {
	base, key := os.Getenv("KUN_NEXTMOE_API_BASE"), os.Getenv("KUN_NEXTMOE_API_KEY")
	if base == "" || key == "" || os.Getenv("SMOKE_CATALOG_V2") == "" {
		t.Skip("set SMOKE_CATALOG_V2=1 with KUN_NEXTMOE_API_BASE and KUN_NEXTMOE_API_KEY")
	}
	c := New(base, key, os.Getenv("KUN_IMAGE_PUBLIC_BASE_URL"))
	q := url.Values{"limit": {"2"}, "include": {"titles,covers,refs"}}
	OpenPopulation(q)
	page, appErr := c.CatalogWorksSearch(context.Background(), q)
	if appErr != nil {
		t.Fatalf("CatalogWorksSearch: %v", appErr)
	}
	if len(page.Items) == 0 {
		t.Fatal("empty works page")
	}
	b := CatalogItemToBrief(context.Background(), &page.Items[0])
	if b.ID <= 0 || b.EffectiveBannerURL == "" {
		t.Fatalf("brief not rewritten: %+v", b)
	}
	d, found, appErr := c.CatalogWorkDetail(context.Background(), b.ID)
	if appErr != nil {
		t.Fatalf("CatalogWorkDetail: %v", appErr)
	}
	if !found || d == nil || d.DisplayName == "" {
		t.Fatalf("detail missing: found=%v d=%+v", found, d)
	}
}

func TestV2CatalogPath(t *testing.T) {
	cases := map[string]string{
		"/catalog/works/search":            "/v2/catalog/works",
		"/catalog/works/12":                "/v2/catalog/works/12",
		"/catalog/labels/3":                "/v2/catalog/companies/3",
		"/catalog/labels/3/relation-graph": "/v2/catalog/companies/3/graph",
		"/catalog/names/9":                 "/v2/catalog/credit-names/9",
		"/catalog/calendar/pending":        "/v2/catalog/calendar",
	}
	for in, want := range cases {
		if got := v2CatalogPath(in); got != want {
			t.Errorf("v2CatalogPath(%q) = %q, want %q", in, got, want)
		}
	}
}

package service

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

	"kun-galgame-api/internal/galgame/client"
)

func TestSearchCatalogEntities_MapsPublicHits(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		if req.URL.Path != "/v1/catalog/search" {
			t.Errorf("path = %q, want /v1/catalog/search", req.URL.Path)
		}
		if got := req.URL.Query().Get("type"); got != "names" {
			t.Errorf("type = %q, want names", got)
		}
		if got := req.URL.Query().Get("q"); got != "丸戸" {
			t.Errorf("q = %q, want 丸戸", got)
		}
		_ = json.NewEncoder(w).Encode(map[string]any{
			"code": 0, "message": "ok",
			"data": map[string]any{
				"items": []map[string]any{
					{"id": 42, "display_name": "丸戸史明", "entity_type": "name",
						"localized": map[string]any{
							"zh-Hans": map[string]any{"value": "丸户史明", "kind": "translation"},
						}},
				},
			},
		})
	}))
	t.Cleanup(srv.Close)

	items, appErr := searchCatalogEntities(
		t.Context(),
		client.New(srv.URL, "nm_test_key", ""),
		"names",
		url.Values{"q": {"丸戸"}},
	)
	if appErr != nil {
		t.Fatalf("searchCatalogEntities: %v", appErr)
	}
	if len(items) != 1 || items[0].ID != 42 || items[0].Name != "丸户史明" {
		t.Errorf("items = %+v, want the hit rendered in the reader's locale", items)
	}
}

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
		if req.URL.Path != "/v2/catalog/search" {
			t.Errorf("path = %q, want /v2/catalog/search", req.URL.Path)
		}
		// v1's type=names is object=credit_name on v2; huma drops the old name
		// silently and answers an unfiltered search across every entity kind.
		if got := req.URL.Query().Get("object"); got != "credit_name" {
			t.Errorf("object = %q, want credit_name", got)
		}
		if req.URL.Query().Has("type") {
			t.Error("v1's type= leaked to the wire")
		}
		if got := req.URL.Query().Get("q"); got != "丸戸" {
			t.Errorf("q = %q, want 丸戸", got)
		}
		_ = json.NewEncoder(w).Encode(map[string]any{
			"object": "list",
			"items": []map[string]any{
				{"object": "search_result", "target_object": "credit_name",
					"id": "42", "display_name": "丸戸史明",
					"localized": map[string]any{
						"zh-Hans": map[string]any{"value": "丸户史明", "is_machine": false},
					}},
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

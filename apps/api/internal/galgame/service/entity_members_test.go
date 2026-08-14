package service

import (
	"context"
	"net/url"
	"testing"
)

func TestEngineDetail_ListsCatalogWorksNotJustForumClaims(t *testing.T) {
	rec := &worksQueryRecorder{}
	svc := NewEngineService(rec.client(t), &GalgameEnricher{}, nil)

	if _, appErr := svc.GetDetail(context.Background(), "11", url.Values{}, true); appErr != nil {
		t.Fatalf("GetDetail: %v", appErr)
	}
	if rec.path != "/v1/catalog/works/search" {
		t.Errorf("path = %q, want /v1/catalog/works/search — the detail must list the catalog, not the forum's own galgame table", rec.path)
	}
	if got := rec.get("engine_id"); got != "11" {
		t.Errorf("engine_id = %q, want 11", got)
	}
	if got := rec.get("claim_state"); got != "live" {
		t.Errorf("claim_state = %q, want live — foreign-claimed works are live too and must be listed", got)
	}
}

func TestCatalogWorksSort(t *testing.T) {
	cases := map[[2]string]string{
		{"release_date", "asc"}:   "released_asc",
		{"release_date", "desc"}:  "released_desc",
		{"release_date", ""}:      "released_desc",
		{"time", "desc"}:          "updated",
		{"view", "desc"}:          "popularity",
		{"rating", "desc"}:        "popularity",
		{"view_30d", "desc"}:      "popularity",
		{"created", "desc"}:       "released_desc",
		{"", "desc"}:              "released_desc",
	}
	for in, want := range cases {
		if got := catalogWorksSort(in[0], in[1]); got != want {
			t.Errorf("catalogWorksSort(%q, %q) = %q, want %q", in[0], in[1], got, want)
		}
	}
}

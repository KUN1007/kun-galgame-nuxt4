package client

import (
	"context"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"
)

func movedCatalog(t *testing.T, survivor string) (*GalgameClient, *[]string) {
	t.Helper()
	var mu sync.Mutex
	var seen []string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		mu.Lock()
		seen = append(seen, r.URL.Path)
		mu.Unlock()
		w.Header().Set("Content-Type", "application/json")
		switch r.URL.Path {
		case "/v2/catalog/companies/" + survivor:
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte(`{"id":"6935","display_name":"生存ブランド","company_kind":"game_brand","work_count":12}`))
		case "/v2/catalog/companies/13323", "/v2/catalog/companies/13324":
			// Captured from company 15845 on the running catalog: v2 announces a
			// merge as a 404 whose code is ENTITY_MERGED, where v1 sent 301 +
			// Location. Reading the status alone turns every merge into "absent".
			w.WriteHeader(http.StatusNotFound)
			_, _ = w.Write([]byte(`{"type":"https://developer.nextmoe.dev/problems/catalog/entity-merged",` +
				`"title":"Entity merged","status":404,"code":"ENTITY_MERGED",` +
				`"object":"company","current_id":"` + survivor + `"}`))
		default:
			w.WriteHeader(http.StatusNotFound)
			_, _ = w.Write([]byte(`{"title":"Not found","status":404,"code":"NOT_FOUND"}`))
		}
	}))
	t.Cleanup(srv.Close)
	return New(srv.URL, "test-key", ""), &seen
}

func TestCatalogLabelMergedIDReportsSurvivor(t *testing.T) {
	c, seen := movedCatalog(t, "6935")

	rec, found, movedTo, appErr := c.CatalogLabel(context.Background(), "13323")
	if appErr != nil {
		t.Fatalf("unexpected error: %v", appErr)
	}
	if found {
		t.Fatal("a merged id must not be reported as found — the caller would render the ghost page")
	}
	if movedTo != 6935 {
		t.Fatalf("movedTo = %d, want 6935", movedTo)
	}
	if rec.DisplayName != "" {
		t.Fatalf("the survivor's content must never travel under the dead id, got %q", rec.DisplayName)
	}
	if len(*seen) != 1 || (*seen)[0] != "/v2/catalog/companies/13323" {
		t.Fatalf("client followed the redirect: %v", *seen)
	}
}

func TestCatalogLabelLiveAndAbsentUnaffected(t *testing.T) {
	c, _ := movedCatalog(t, "6935")

	rec, found, movedTo, appErr := c.CatalogLabel(context.Background(), "6935")
	if appErr != nil {
		t.Fatalf("unexpected error: %v", appErr)
	}
	if !found || movedTo != 0 || rec.DisplayName != "生存ブランド" {
		t.Fatalf("live label mangled: found=%v movedTo=%d name=%q", found, movedTo, rec.DisplayName)
	}

	_, found, movedTo, appErr = c.CatalogLabel(context.Background(), "999999")
	if appErr != nil {
		t.Fatalf("an absent id is a miss, not an error: %v", appErr)
	}
	if found || movedTo != 0 {
		t.Fatalf("absent id: found=%v movedTo=%d, want false/0", found, movedTo)
	}
}

func TestCatalogLabelChainIsResolvedUpstream(t *testing.T) {
	c, seen := movedCatalog(t, "6935")

	_, _, movedTo, appErr := c.CatalogLabel(context.Background(), "13324")
	if appErr != nil {
		t.Fatalf("unexpected error: %v", appErr)
	}
	if movedTo != 6935 {
		t.Fatalf("movedTo = %d, want the final survivor 6935", movedTo)
	}
	if len(*seen) != 1 {
		t.Fatalf("expected exactly one upstream call, got %v", *seen)
	}
}

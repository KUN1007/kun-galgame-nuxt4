package service

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"kun-galgame-api/internal/galgame/client"
)

func tagLaneStub(t *testing.T) *httptest.Server {
	t.Helper()
	row := func(id int, name, tier string, sexual bool, works int) map[string]any {
		return map[string]any{
			"id": id, "name": name, "tier": tier, "kind": "content",
			"sexual": sexual, "work_count": works,
		}
	}
	page1 := []map[string]any{
		row(1, "青梅竹马", "core", false, 300),
		row(2, "游戏", "hidden", false, 9000),
		row(3, "陵辱", "core", true, 500),
	}
	page2 := []map[string]any{
		row(4, "校园", "core", false, 400),
		row(5, "PC", "hidden", false, 8000),
		row(6, "触手", "core", true, 100),
	}
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		if !strings.HasSuffix(req.URL.Path, "/catalog/tags") {
			http.NotFound(w, req)
			return
		}
		items, next := page1, "p2"
		if req.URL.Query().Get("cursor") == "p2" {
			items, next = page2, ""
		}
		_ = json.NewEncoder(w).Encode(map[string]any{
			"object": "list",
			"items":  items, "next_cursor": next,
			"total": len(page1) + len(page2),
		})
	}))
	t.Cleanup(srv.Close)
	return srv
}

func tagPage(t *testing.T, svc *TagService, isSFW bool, q url.Values) ([]string, int64) {
	t.Helper()
	page, appErr := svc.GetList(context.Background(), q, isSFW)
	if appErr != nil {
		t.Fatalf("GetList: %v", appErr)
	}
	names := make([]string, 0, len(page.Tags))
	for _, tag := range page.Tags {
		names = append(names, tag.Name)
	}
	return names, page.Total
}

func TestTagList_FiltersBeforePagingAndTotalsWhatItServes(t *testing.T) {
	svc := NewTagService(client.New(tagLaneStub(t).URL, "nm_test_key", ""), nil, nil)

	open, total := tagPage(t, svc, false, url.Values{})
	want := []string{"陵辱", "校园", "青梅竹马", "触手"}
	if strings.Join(open, ",") != strings.Join(want, ",") {
		t.Errorf("open list = %v, want %v", open, want)
	}
	if total != 4 {
		t.Errorf("open total = %d, want 4 — never upstream's 6", total)
	}

	sfw, sfwTotal := tagPage(t, svc, true, url.Values{})
	if strings.Join(sfw, ",") != "校园,青梅竹马" {
		t.Errorf("SFW list = %v, want the two non-adult terms", sfw)
	}
	if sfwTotal != 2 {
		t.Errorf("SFW total = %d, want 2 — the total moves with the gate, or the pager lies", sfwTotal)
	}
}

func TestTagList_PagesAreFull(t *testing.T) {
	svc := NewTagService(client.New(tagLaneStub(t).URL, "nm_test_key", ""), nil, nil)

	first, total := tagPage(t, svc, false, url.Values{"page": {"1"}, "limit": {"3"}})
	if len(first) != 3 {
		t.Errorf("page 1 of 3 = %v, want three rows", first)
	}
	second, _ := tagPage(t, svc, false, url.Values{"page": {"2"}, "limit": {"3"}})
	if len(second) != 1 {
		t.Errorf("page 2 of 3 = %v, want the one remaining row", second)
	}
	if last := (total + 2) / 3; last != 2 {
		t.Errorf("total %d implies %d pages, want 2", total, last)
	}
	beyond, _ := tagPage(t, svc, false, url.Values{"page": {"3"}, "limit": {"3"}})
	if len(beyond) != 0 {
		t.Errorf("page 3 = %v, want nothing", beyond)
	}
}

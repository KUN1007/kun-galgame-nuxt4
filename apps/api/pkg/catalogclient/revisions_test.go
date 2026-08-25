package catalogclient

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestEditRevisionsSince_RequestAndParse(t *testing.T) {
	var gotPath, gotSort, gotLimit, gotObject, gotCursor string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path
		gotSort = r.URL.Query().Get("sort")
		gotLimit = r.URL.Query().Get("limit")
		gotObject = r.URL.Query().Get("object")
		gotCursor = r.URL.Query().Get("cursor")
		_, _ = w.Write([]byte(`{"object":"list","items":[
			{"id":"41","target_object":"work","entity_id":"1207",
			 "seq":8,"action":"merged","changed_fields":["catalog.work.name_ja_jp"],
			 "actor_uid":"9","amender_uid":null,"proposal_id":"77","site":"kungal",
			 "created_at":"2026-07-30T10:00:00Z"}]}`))
	}))
	defer srv.Close()

	c := New(Config{BaseURL: srv.URL, AppKey: "nmk_test"})
	page, err := c.EditRevisionsSince(context.Background(), 40, 1000, EntityTypeWork)
	if err != nil {
		t.Fatalf("EditRevisionsSince: %v", err)
	}

	if gotPath != "/v2/catalog/revisions" {
		t.Errorf("path = %q", gotPath)
	}
	if gotSort != "recorded_asc" || gotLimit != "100" || gotObject != "work" {
		t.Errorf("query = sort:%q limit:%q object:%q", gotSort, gotLimit, gotObject)
	}
	if gotCursor != encodeWatermark(40) {
		t.Errorf("cursor = %q, want watermark of 40", gotCursor)
	}
	if len(page.Items) != 1 {
		t.Fatalf("items = %d, want 1", len(page.Items))
	}
	got := page.Items[0]
	if got.ID != 41 || got.EntityID != 1207 || got.Seq != 8 || got.ActorUID != 9 {
		t.Errorf("item = %+v", got)
	}
	if got.Action != EditActionMerged {
		t.Errorf("action = %d, want merged", got.Action)
	}
	if page.NextSince != 41 {
		t.Errorf("next_since = %d, want 41", page.NextSince)
	}
}

func TestEditRevisionsSince_EmptyPage(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write([]byte(`{"object":"list","items":[]}`))
	}))
	defer srv.Close()

	c := New(Config{BaseURL: srv.URL, AppKey: "nmk_test"})
	page, err := c.EditRevisionsSince(context.Background(), 12681, 100, EntityTypeWork)
	if err != nil {
		t.Fatalf("EditRevisionsSince: %v", err)
	}
	if len(page.Items) != 0 || page.NextSince != 12681 {
		t.Errorf("page = %+v", page)
	}
}

func TestEditRevisionsSince_NotConfigured(t *testing.T) {
	if _, err := New(Config{}).EditRevisionsSince(context.Background(), 0, 10, ""); err == nil {
		t.Fatal("want an error from an unconfigured client")
	}
}

func TestListEditRevisionsWalksPages(t *testing.T) {
	var limits []string
	pages := 0
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		pages++
		limits = append(limits, r.URL.Query().Get("limit"))
		if r.URL.Query().Get("cursor") == "" {
			_, _ = w.Write([]byte(`{"object":"list","next_cursor":"` + encodeWatermark(10) + `","items":[` +
				`{"id":"10","seq":10,"action":"merged","actor_uid":"1","entity_id":"1000","target_object":"work"}]}`))
			return
		}
		_, _ = w.Write([]byte(`{"object":"list","items":[` +
			`{"id":"3","seq":3,"action":"created","actor_uid":"1","entity_id":"1000","target_object":"work"}]}`))
	}))
	defer srv.Close()

	c := New(Config{BaseURL: srv.URL, AppKey: "nmk_test"})
	items, err := c.ListEditRevisions(context.Background(), EntityTypeWork, 1000, 150)
	if err != nil {
		t.Fatalf("ListEditRevisions: %v", err)
	}
	if pages != 2 {
		t.Fatalf("pages = %d, want 2 so a limit above 100 still fills", pages)
	}
	if len(items) != 2 || items[0].Seq != 10 || items[1].Seq != 3 {
		t.Fatalf("items = %+v", items)
	}
	if limits[0] != "100" {
		t.Fatalf("first page limit = %s, want 100", limits[0])
	}

	id, err := c.RevisionIDBySeq(context.Background(), EntityTypeWork, 1000, 3)
	if err != nil {
		t.Fatalf("RevisionIDBySeq: %v", err)
	}
	if id != 3 {
		t.Fatalf("id = %d, want 3", id)
	}
}

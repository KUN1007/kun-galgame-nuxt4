package newsclient

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestSourcesDecodesDirectory(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/v2/news/sources" {
			t.Errorf("unexpected path %q", r.URL.Path)
		}
		if r.Header.Get("Authorization") != "Bearer k" {
			t.Errorf("missing bearer, got %q", r.Header.Get("Authorization"))
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"code":0,"data":{"sources":[
			{"key":"ymgal","display_name":"月幕 Galgame","homepage_url":"https://www.ymgal.games",
			 "attribution":"本条情报转载自月幕 Galgame","column_url":"","publisher_uid":114748},
			{"key":"galgame_hihyou","display_name":"Galgame 批评","homepage_url":"https://space.bilibili.com/2072586344",
			 "attribution":"本条情报转载自 Galgame 批评","column_url":"https://space.bilibili.com/2072586344/article","publisher_uid":115235}
		]}}`))
	}))
	defer srv.Close()

	got, err := New(Config{BaseURL: srv.URL, APIKey: "k"}).Sources(context.Background())
	if err != nil {
		t.Fatalf("Sources: %v", err)
	}
	if len(got) != 2 {
		t.Fatalf("got %d sources, want 2", len(got))
	}
	if got[0].Key != "ymgal" || got[0].DisplayName != "月幕 Galgame" || got[0].PublisherUID != 114748 {
		t.Errorf("first source decoded wrong: %+v", got[0])
	}
	if got[1].ColumnURL == "" || got[1].Attribution == "" {
		t.Errorf("attribution and column url must survive: %+v", got[1])
	}
}

func TestSourcesUnconfigured(t *testing.T) {
	if _, err := New(Config{}).Sources(context.Background()); !errors.Is(err, ErrNotConfigured) {
		t.Fatalf("got %v, want ErrNotConfigured", err)
	}
}

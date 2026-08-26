package service

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"kun-galgame-api/internal/galgame/client"
)

func graphService(t *testing.T) *OfficialService {
	t.Helper()
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		if r.URL.Path != "/v1/catalog/labels/24/relation-graph" {
			w.WriteHeader(http.StatusNotFound)
			_, _ = w.Write([]byte(`{"code":4,"message":"资源不存在"}`))
			return
		}
		_, _ = w.Write([]byte(`{"code":0,"message":"成功","data":{"nodes":[` +
			`{"id":24,"display_name":"ねこねこソフト","localized":{"zh-Hans":{"value":"猫猫社","kind":"translation"}},"logo_hash":"aabbccddeeff","work_count":33},` +
			`{"id":993,"display_name":"VisualArt's","logo_hash":"","work_count":120}],` +
			`"edges":[{"from":24,"to":993,"relation":"parent"}]}}`))
	}))
	t.Cleanup(srv.Close)
	return NewOfficialService(client.New(srv.URL, "nm_test_key", "https://cdn.test/image"), nil)
}

func TestOfficialRelationGraphResolvesLogosToCDNURLs(t *testing.T) {
	svc := graphService(t)

	graph, appErr := svc.GetRelationGraph(context.Background(), "24")
	if appErr != nil {
		t.Fatalf("unexpected error: %v", appErr)
	}
	if len(graph.Nodes) != 2 || len(graph.Edges) != 1 {
		t.Fatalf("graph = %d nodes / %d edges, want 2/1", len(graph.Nodes), len(graph.Edges))
	}
	want := "https://cdn.test/image/aa/bb/aabbccddeeff.webp"
	if got := graph.Nodes[0].Logo; got != want {
		t.Errorf("node logo = %q, want %q", got, want)
	}
	if got := graph.Nodes[1].Logo; got != "" {
		t.Errorf("logoless node = %q, want empty", got)
	}
	if graph.Nodes[0].Name != "猫猫社" || graph.Nodes[1].Name != "VisualArt's" {
		t.Errorf("node names = %q / %q, want 猫猫社 / VisualArt's",
			graph.Nodes[0].Name, graph.Nodes[1].Name)
	}
	if graph.Nodes[1].WorkCount != 120 {
		t.Errorf("work_count = %d, want 120", graph.Nodes[1].WorkCount)
	}
	if e := graph.Edges[0]; e.From != 24 || e.To != 993 || e.Relation != "parent" {
		t.Errorf("edge = %+v, want 24→993 parent", e)
	}
}

func TestOfficialRelationGraphUnknownLabelIs404(t *testing.T) {
	svc := graphService(t)

	graph, appErr := svc.GetRelationGraph(context.Background(), "999999")
	if graph != nil {
		t.Fatalf("graph = %v, want nil on an unknown label", graph)
	}
	if appErr == nil || appErr.StatusCode != 404 {
		t.Fatalf("appErr = %v, want a 404 — an unknown 会社 must not read as a 200 empty graph", appErr)
	}
}

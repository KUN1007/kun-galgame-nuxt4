package service

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"reflect"
	"strconv"
	"testing"

	"kun-galgame-api/pkg/communityclient"
)

func TestPaginateGalgameCommentEntriesUsesCreatedAtAndID(t *testing.T) {
	posts := []communityclient.AuthorPostView{
		authorPost(400, "2026-08-20T08:00:00Z"),
		authorPost(200, "2026-08-20T10:00:00Z"),
		authorPost(300, "2026-08-20T10:00:00Z"),
		authorPost(100, "2026-08-20T09:00:00Z"),
	}

	first, cursor := paginateGalgameCommentEntries(authoredGalgameCommentEntries(posts), "", 2)
	if got, want := entryIDs(first), []int64{300, 200}; !reflect.DeepEqual(got, want) {
		t.Fatalf("first page ids = %v, want %v", got, want)
	}
	if cursor == "" {
		t.Fatal("first page must return a composite cursor")
	}
	decoded, ok := decodeGalgameCommentCursor(cursor)
	if !ok || decoded.ID != 200 || decoded.Created != parseGalgameCommentTime("2026-08-20T10:00:00Z") {
		t.Fatalf("decoded cursor = %+v, ok=%v", decoded, ok)
	}

	second, next := paginateGalgameCommentEntries(authoredGalgameCommentEntries(posts), cursor, 2)
	if got, want := entryIDs(second), []int64{100, 400}; !reflect.DeepEqual(got, want) {
		t.Fatalf("second page ids = %v, want %v", got, want)
	}
	if next != "" {
		t.Fatalf("last page cursor = %q, want empty", next)
	}
}

func TestPaginateGalgameCommentEntriesSplitsEqualTimestampsByID(t *testing.T) {
	const created = "2026-08-20T10:00:00Z"
	posts := []communityclient.AuthorPostView{
		authorPost(10, created),
		authorPost(30, created),
		authorPost(20, created),
	}

	first, cursor := paginateGalgameCommentEntries(authoredGalgameCommentEntries(posts), "", 2)
	if got, want := entryIDs(first), []int64{30, 20}; !reflect.DeepEqual(got, want) {
		t.Fatalf("first page ids = %v, want %v", got, want)
	}

	second, next := paginateGalgameCommentEntries(authoredGalgameCommentEntries(posts), cursor, 2)
	if got, want := entryIDs(second), []int64{10}; !reflect.DeepEqual(got, want) {
		t.Fatalf("second page ids = %v, want %v", got, want)
	}
	if next != "" {
		t.Fatalf("last page cursor = %q, want empty", next)
	}
}

func TestLikedGalgameCommentEntriesKeepMissingPostsStable(t *testing.T) {
	postIDs := []int64{20, 900, 30}
	posts := []communityclient.AuthorPostView{
		authorPost(20, "2026-08-20T09:00:00Z"),
		authorPost(30, "2026-08-20T10:00:00Z"),
	}

	first, cursor := paginateGalgameCommentEntries(likedGalgameCommentEntries(postIDs, posts), "", 2)
	if got, want := entryIDs(first), []int64{30, 20}; !reflect.DeepEqual(got, want) {
		t.Fatalf("first liked page ids = %v, want %v", got, want)
	}
	if cursor == "" {
		t.Fatal("first liked page must return a composite cursor")
	}

	second, next := paginateGalgameCommentEntries(likedGalgameCommentEntries(postIDs, posts), cursor, 2)
	if got, want := entryIDs(second), []int64{900}; !reflect.DeepEqual(got, want) {
		t.Fatalf("second liked page ids = %v, want %v", got, want)
	}
	if second[0].Post != nil {
		t.Fatal("an unresolved liked post must remain a deleted placeholder")
	}
	if next != "" {
		t.Fatalf("last page cursor = %q, want empty", next)
	}
}

func TestCollectAuthorGalgamePostsExhaustsUpstreamBeforeSorting(t *testing.T) {
	var requests []string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Query().Get("limit") != strconv.Itoa(communityPostBatchSize) ||
			r.URL.Query().Get("anchor_kind") != strconv.Itoa(communityclient.AnchorSiteGame) {
			t.Errorf("unexpected query: %s", r.URL.RawQuery)
		}
		after := r.URL.Query().Get("after")
		requests = append(requests, after)
		posts := []communityclient.AuthorPostView{
			authorPost(300, "2026-08-20T08:00:00Z"),
			authorPost(200, "2026-08-20T10:00:00Z"),
		}
		next := "200"
		if after == "200" {
			posts = []communityclient.AuthorPostView{authorPost(100, "2026-08-20T09:00:00Z")}
			next = ""
		}
		_ = json.NewEncoder(w).Encode(map[string]any{
			"code": 0,
			"data": map[string]any{"posts": posts, "next_cursor": next},
		})
	}))
	defer srv.Close()

	s := &UserContentService{community: communityclient.New(communityclient.Config{
		BaseURL: srv.URL, ClientID: "test", ClientSecret: "test",
	})}
	posts, err := s.collectAuthorGalgamePosts(context.Background(), 7)
	if err != nil {
		t.Fatalf("collectAuthorGalgamePosts: %v", err)
	}
	page, _ := paginateGalgameCommentEntries(authoredGalgameCommentEntries(posts), "", 3)
	if got, want := entryIDs(page), []int64{200, 100, 300}; !reflect.DeepEqual(got, want) {
		t.Fatalf("globally sorted ids = %v, want %v", got, want)
	}
	if got, want := requests, []string{"", "200"}; !reflect.DeepEqual(got, want) {
		t.Fatalf("upstream cursors = %v, want %v", got, want)
	}
}

func TestResolveGalgameCommentPostsUsesContractBatchSize(t *testing.T) {
	var batchSizes []int
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var req communityclient.PostsResolveRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			t.Errorf("decode request: %v", err)
			http.Error(w, "bad request", http.StatusBadRequest)
			return
		}
		batchSizes = append(batchSizes, len(req.IDs))
		posts := make([]communityclient.AuthorPostView, len(req.IDs))
		for i, id := range req.IDs {
			posts[i] = authorPost(id, "2026-08-20T10:00:00Z")
		}
		_ = json.NewEncoder(w).Encode(map[string]any{
			"code": 0,
			"data": map[string]any{"posts": posts},
		})
	}))
	defer srv.Close()

	s := &UserContentService{community: communityclient.New(communityclient.Config{
		BaseURL: srv.URL, ClientID: "test", ClientSecret: "test",
	})}
	ids := make([]int64, 201)
	for i := range ids {
		ids[i] = int64(i + 1)
	}
	posts, err := s.resolveGalgameCommentPosts(context.Background(), ids)
	if err != nil {
		t.Fatalf("resolveGalgameCommentPosts: %v", err)
	}
	if len(posts) != len(ids) {
		t.Fatalf("resolved posts = %d, want %d", len(posts), len(ids))
	}
	if got, want := batchSizes, []int{100, 100, 1}; !reflect.DeepEqual(got, want) {
		t.Fatalf("batch sizes = %v, want %v", got, want)
	}
}

func authorPost(id int64, created string) communityclient.AuthorPostView {
	return communityclient.AuthorPostView{
		Post: communityclient.PostView{ID: id, CreatedAt: created},
		Thread: communityclient.PostThreadContext{
			AnchorKind: communityclient.AnchorSiteGame,
			AnchorID:   "1",
		},
	}
}

func entryIDs(entries []galgameCommentEntry) []int64 {
	ids := make([]int64, len(entries))
	for i, entry := range entries {
		ids[i] = entry.ID
	}
	return ids
}

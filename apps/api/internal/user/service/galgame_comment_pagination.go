package service

import (
	"cmp"
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"log/slog"
	"slices"
	"time"

	"kun-galgame-api/pkg/communityclient"
)

// The community author feed is ordered by post id and its cursor IS a post id,
// but the 2026-07 import gave galgame posts ids that have nothing to do with
// their age: 9,610 of the 10,469 galgame posts sit at a different position
// under id order than under created_at order (the busiest author: 433 of 434,
// max shift 430). Paging on that cursor is what made the profile tab look
// unsorted. Until /authors/{id}/posts can order by created_at, the only way to
// page by time is to read the author's whole feed and sort it here — this is a
// stopgap for a missing upstream sort, not the shape this wants to keep.
//
// 100 is measured, not chosen: the feed silently clamps limit to 100 (asking
// for 200 returns 100), and /posts/resolve answers 422 LIMIT_TOO_LARGE at 101
// ids.
const communityPostBatchSize = 100

// A drain with no ceiling is a request handler someone can make issue unbounded
// upstream calls, and this one is reachable unauthenticated on any profile.
// The busiest author needs 5 pages today.
const maxAuthorPostPages = 40

type galgameCommentCursor struct {
	Created time.Time `json:"c"`
	ID      int64     `json:"i"`
}

type galgameCommentEntry struct {
	ID      int64
	Created time.Time
	Post    *communityclient.AuthorPostView
}

func (s *UserContentService) collectAuthorGalgamePosts(ctx context.Context, userID int) ([]communityclient.AuthorPostView, error) {
	var posts []communityclient.AuthorPostView
	after := ""
	seen := map[string]struct{}{}
	for page := 0; page < maxAuthorPostPages; page++ {
		res, err := s.community.AuthorPosts(ctx, int64(userID), after, communityPostBatchSize, communityclient.AnchorSiteGame)
		if err != nil {
			return nil, err
		}
		posts = append(posts, res.Posts...)
		if res.NextCursor == "" {
			return posts, nil
		}
		if _, ok := seen[res.NextCursor]; ok {
			return nil, fmt.Errorf("community author posts repeated cursor %q", res.NextCursor)
		}
		seen[res.NextCursor] = struct{}{}
		after = res.NextCursor
	}
	slog.Warn("author galgame posts hit the page cap; the oldest comments are ordered wrong",
		"user_id", userID, "pages", maxAuthorPostPages, "posts", len(posts))
	return posts, nil
}

func (s *UserContentService) resolveGalgameCommentPosts(ctx context.Context, postIDs []int64) ([]communityclient.AuthorPostView, error) {
	posts := make([]communityclient.AuthorPostView, 0, len(postIDs))
	for chunk := range slices.Chunk(postIDs, communityPostBatchSize) {
		page, err := s.community.ResolvePosts(ctx, chunk)
		if err != nil {
			return nil, err
		}
		posts = append(posts, page.Posts...)
	}
	return posts, nil
}

func authoredGalgameCommentEntries(posts []communityclient.AuthorPostView) []galgameCommentEntry {
	entries := make([]galgameCommentEntry, len(posts))
	for i := range posts {
		entries[i] = galgameCommentEntry{
			ID:      posts[i].Post.ID,
			Created: parseGalgameCommentTime(posts[i].Post.CreatedAt),
			Post:    &posts[i],
		}
	}
	return entries
}

func likedGalgameCommentEntries(postIDs []int64, posts []communityclient.AuthorPostView) []galgameCommentEntry {
	byID := make(map[int64]*communityclient.AuthorPostView, len(posts))
	for i := range posts {
		byID[posts[i].Post.ID] = &posts[i]
	}
	entries := make([]galgameCommentEntry, len(postIDs))
	for i, postID := range postIDs {
		post := byID[postID]
		entries[i] = galgameCommentEntry{ID: postID, Post: post}
		if post != nil {
			entries[i].Created = parseGalgameCommentTime(post.Post.CreatedAt)
		}
	}
	return entries
}

func paginateGalgameCommentEntries(entries []galgameCommentEntry, after string, limit int) ([]galgameCommentEntry, string) {
	if limit <= 0 {
		return []galgameCommentEntry{}, ""
	}
	slices.SortFunc(entries, func(a, b galgameCommentEntry) int {
		if c := b.Created.Compare(a.Created); c != 0 {
			return c
		}
		return cmp.Compare(b.ID, a.ID)
	})

	if cursor, ok := decodeGalgameCommentCursor(after); ok {
		// Sorted, so everything past the cursor is one contiguous tail.
		start := slices.IndexFunc(entries, func(e galgameCommentEntry) bool {
			return e.Created.Before(cursor.Created) ||
				(e.Created.Equal(cursor.Created) && e.ID < cursor.ID)
		})
		if start < 0 {
			return []galgameCommentEntry{}, ""
		}
		entries = entries[start:]
	}
	if len(entries) <= limit {
		return entries, ""
	}
	page := entries[:limit]
	last := page[len(page)-1]
	return page, encodeGalgameCommentCursor(last.Created, last.ID)
}

func encodeGalgameCommentCursor(created time.Time, id int64) string {
	raw, err := json.Marshal(galgameCommentCursor{Created: created, ID: id})
	if err != nil {
		return ""
	}
	return base64.RawURLEncoding.EncodeToString(raw)
}

func decodeGalgameCommentCursor(cursor string) (galgameCommentCursor, bool) {
	if cursor == "" {
		return galgameCommentCursor{}, false
	}
	raw, err := base64.RawURLEncoding.DecodeString(cursor)
	if err != nil {
		return galgameCommentCursor{}, false
	}
	var payload galgameCommentCursor
	if err := json.Unmarshal(raw, &payload); err != nil || payload.ID <= 0 {
		return galgameCommentCursor{}, false
	}
	return payload, true
}

func parseGalgameCommentTime(value string) time.Time {
	created, err := time.Parse(time.RFC3339Nano, value)
	if err != nil {
		// Every entry at the zero time sorts by id, which is exactly the bug
		// this file exists to undo — silently. Say so instead.
		slog.Warn("community post created_at is unparseable; the tab falls back to id order",
			"value", value, "error", err)
	}
	return created
}

package service

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"sort"
	"time"

	"kun-galgame-api/pkg/communityclient"
)

const communityPostBatchSize = 100

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
	for {
		page, err := s.community.AuthorPosts(ctx, int64(userID), after, communityPostBatchSize, communityclient.AnchorSiteGame)
		if err != nil {
			return nil, err
		}
		posts = append(posts, page.Posts...)
		if page.NextCursor == "" {
			return posts, nil
		}
		if _, ok := seen[page.NextCursor]; ok {
			return nil, fmt.Errorf("community author posts repeated cursor %q", page.NextCursor)
		}
		seen[page.NextCursor] = struct{}{}
		after = page.NextCursor
	}
}

func (s *UserContentService) resolveGalgameCommentPosts(ctx context.Context, postIDs []int64) ([]communityclient.AuthorPostView, error) {
	posts := make([]communityclient.AuthorPostView, 0, len(postIDs))
	for start := 0; start < len(postIDs); start += communityPostBatchSize {
		end := min(start+communityPostBatchSize, len(postIDs))
		page, err := s.community.ResolvePosts(ctx, postIDs[start:end])
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
	sort.Slice(entries, func(i, j int) bool {
		if !entries[i].Created.Equal(entries[j].Created) {
			return entries[i].Created.After(entries[j].Created)
		}
		return entries[i].ID > entries[j].ID
	})

	if cursor, ok := decodeGalgameCommentCursor(after); ok {
		filtered := entries[:0]
		for _, entry := range entries {
			if entry.Created.Before(cursor.Created) ||
				(entry.Created.Equal(cursor.Created) && entry.ID < cursor.ID) {
				filtered = append(filtered, entry)
			}
		}
		entries = filtered
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
	created, _ := time.Parse(time.RFC3339Nano, value)
	return created
}

package service

import (
	"net/url"
	"testing"

	"kun-galgame-api/internal/galgame/dto"
)

func TestBuildEntityFilter(t *testing.T) {
	q := url.Values{}
	q.Set("type", "patch")
	q.Set("language", "zh-cn")
	q.Set("platform", "windows")
	q.Set("gameType", "moe")
	q.Set("sortField", "view_7d")
	q.Set("sortOrder", "asc")
	q.Set("page", "3")
	q.Set("limit", "24")
	q.Set("showNoResource", "true")

	f := buildEntityFilter(q, []int{5, 9, 12})

	if f.Type != "patch" || f.Language != "zh-cn" || f.Platform != "windows" {
		t.Errorf("resource filters not read: %+v", f)
	}
	if f.GameType != "moe" {
		t.Errorf("GameType = %q, want moe (chip must not be dead)", f.GameType)
	}
	if f.SortField != "view_7d" || f.SortOrder != "asc" {
		t.Errorf("sort = %q/%q, want view_7d/asc", f.SortField, f.SortOrder)
	}
	if f.Page != 3 || f.Limit != 24 {
		t.Errorf("page/limit = %d/%d, want 3/24", f.Page, f.Limit)
	}
	if !f.ShowNoResource {
		t.Error("ShowNoResource should be true")
	}
	if len(f.RestrictIDs) != 3 || f.RestrictIDs[0] != 5 {
		t.Errorf("RestrictIDs = %v, want [5 9 12]", f.RestrictIDs)
	}
}

func TestBuildEntityFilterDefaults(t *testing.T) {
	f := buildEntityFilter(url.Values{}, []int{})

	if f.SortOrder != "desc" {
		t.Errorf("SortOrder default = %q, want desc", f.SortOrder)
	}
	if f.Page != 1 || f.Limit != 24 {
		t.Errorf("page/limit default = %d/%d, want 1/24", f.Page, f.Limit)
	}
	if f.ShowNoResource {
		t.Error("ShowNoResource default should be false")
	}
	if f.RestrictIDs == nil {
		t.Error("RestrictIDs must stay non-nil so an empty entity restricts to nothing")
	}
}

// Entity pages list every catalog member since 方案③, so this may not assert
// true: a member with no forum resource drew a "0 浏览 0 点赞" strip and an
// empty byline, the same card defect galgame_enricher.go records.
func TestListCardsToEntityCardsIsOnForum(t *testing.T) {
	out := listCardsToEntityCards([]dto.GalgameListCard{
		{ID: 1, View: 10, IsOnForum: true},
		{ID: 2},
	})
	if len(out) != 2 || out[0].ID != 1 || out[0].View != 10 {
		t.Fatalf("field copy wrong: %+v", out)
	}
	if !out[0].IsOnForum {
		t.Error("IsOnForum should survive the copy for a card that has a resource")
	}
	if out[1].IsOnForum {
		t.Error("a catalog member with no forum resource must not claim IsOnForum")
	}
}

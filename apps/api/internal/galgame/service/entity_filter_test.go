package service

import (
	"net/url"
	"testing"

	"kun-galgame-api/internal/galgame/dto"
	"kun-galgame-api/internal/galgame/model"
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

	f := buildEntityFilter(q)

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
}

func TestBuildEntityFilterDefaults(t *testing.T) {
	f := buildEntityFilter(url.Values{})

	if f.SortOrder != "desc" {
		t.Errorf("SortOrder default = %q, want desc", f.SortOrder)
	}
	if f.Page != 1 || f.Limit != 24 {
		t.Errorf("page/limit default = %d/%d, want 1/24", f.Page, f.Limit)
	}
	if f.ShowNoResource {
		t.Error("ShowNoResource default should be false")
	}
}

// The sort chips on an entity page are the forum's, and only the release date is
// a column catalog also holds. Answering the rest out of catalog's vocabulary
// would silently rank the page by something the reader did not ask for.
func TestCatalogMemberSortOnlyClaimsTheReleaseDate(t *testing.T) {
	for _, tc := range []struct{ field, order, want string }{
		{"release_date", "desc", "released_desc"},
		{"release_date", "asc", "released_asc"},
		{"view", "desc", ""},
		{"rating", "desc", ""},
		{"time", "desc", ""},
		{"", "desc", ""},
	} {
		f := model.GalgameListFilter{SortField: tc.field, SortOrder: tc.order}
		if got := catalogMemberSort(f); got != tc.want {
			t.Errorf("catalogMemberSort(%q/%q) = %q, want %q", tc.field, tc.order, got, tc.want)
		}
	}
}

func TestEntityMemberQueryCarriesTheWalkOrder(t *testing.T) {
	q := entityMemberQuery("tag_id", "1337", model.GalgameListFilter{
		SortField: "release_date", SortOrder: "asc",
	})
	if q.Get("tag_id") != "1337" || q.Get("sort") != "released_asc" {
		t.Errorf("query = %v, want tag_id=1337 and sort=released_asc", q)
	}
	q = entityMemberQuery("tag_id", "1337", model.GalgameListFilter{SortField: "view"})
	if q.Has("sort") {
		t.Errorf("sort = %q reached the walk; catalog cannot rank the forum's own counters",
			q.Get("sort"))
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

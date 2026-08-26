package repository

import (
	"testing"

	"kun-galgame-api/internal/galgame/model"
	"kun-galgame-api/internal/testdb"

	"gorm.io/gorm"
)

// NULL means the content-limit sync has not reached the row yet, and it MUST
// stay listed: the column ships empty, so a filter that reads NULL as "not sfw"
// empties every list between the migration and the first sync run — and keeps
// the rows catalog has no work for hidden forever. Catalog's own gate at
// hydrate time is what actually hides a work; this predicate only decides which
// ids get that far, so erring open costs a card, erring closed costs the page.
func TestListIDsSFWFilter(t *testing.T) {
	db := testdb.Open(t)

	const base = 2_000_100_000
	unsynced, safe, adult := base, base+1, base+2
	all := []int{unsynced, safe, adult}

	cleanup := func() {
		db.Exec("DELETE FROM galgame_resource WHERE galgame_id = ANY(?::int[])", intArrayLit(all))
		db.Exec("DELETE FROM galgame WHERE id = ANY(?::int[])", intArrayLit(all))
	}
	cleanup()
	defer cleanup()

	seed(t, db, unsynced, nil)
	seed(t, db, safe, ptr("sfw"))
	seed(t, db, adult, ptr("nsfw"))

	repo := NewGalgameListRepository(db)
	for name, tc := range map[string]struct {
		filter model.GalgameListFilter
		want   []int
	}{
		"sfw reader keeps unsynced and sfw": {
			model.GalgameListFilter{SFWOnly: true},
			[]int{unsynced, safe},
		},
		"nsfw reader keeps everything": {
			model.GalgameListFilter{},
			all,
		},
		"the resource-filter lane gates too": {
			model.GalgameListFilter{SFWOnly: true, Platform: "windows"},
			[]int{unsynced, safe},
		},
	} {
		t.Run(name, func(t *testing.T) {
			f := tc.filter
			f.RestrictIDs, f.Page, f.Limit, f.SortOrder = all, 1, 10, "desc"
			ids, total := repo.ListIDs(f)
			if total != int64(len(tc.want)) {
				t.Errorf("total = %d, want %d — the pager counts what the reader can reach", total, len(tc.want))
			}
			if !sameSet(ids, tc.want) {
				t.Errorf("ids = %v, want %v", ids, tc.want)
			}
		})
	}
}

func seed(t *testing.T, db *gorm.DB, id int, contentLimit *string) {
	t.Helper()
	if err := db.Create(&model.GalgameLocal{
		ID: id, Published: true, ContentLimit: contentLimit,
	}).Error; err != nil {
		t.Fatalf("seed galgame %d: %v", id, err)
	}
	if err := db.Create(&model.GalgameResource{
		GalgameID: id, UserID: 1, Platform: "windows", Language: "ja", Type: "game",
	}).Error; err != nil {
		t.Fatalf("seed resource for %d: %v", id, err)
	}
}

func ptr(s string) *string { return &s }

func sameSet(got, want []int) bool {
	if len(got) != len(want) {
		return false
	}
	seen := make(map[int]bool, len(got))
	for _, id := range got {
		seen[id] = true
	}
	for _, id := range want {
		if !seen[id] {
			return false
		}
	}
	return true
}

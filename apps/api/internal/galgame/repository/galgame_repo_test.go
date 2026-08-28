package repository

import (
	"testing"

	"kun-galgame-api/internal/galgame/model"
	"kun-galgame-api/internal/testdb"
)

// galgame_resource is ON DELETE CASCADE, so an unguarded delete of a local row
// takes the resources with it and says nothing. A draft claim carrying a
// published resource is reachable — publishing a resource sets `published`
// without moving the claim state — and catalog will happily delete that draft.
func TestDeleteLocalDraftRefusesARowThatCarriesAResource(t *testing.T) {
	db := testdb.Open(t)
	repo := NewGalgameRepository(db)

	const base = 2_000_200_000
	bare, withResource := base, base+1
	all := []int{bare, withResource}

	cleanup := func() {
		db.Exec("DELETE FROM galgame_resource WHERE galgame_id = ANY(?::int[])", intArrayLit(all))
		db.Exec("DELETE FROM galgame WHERE id = ANY(?::int[])", intArrayLit(all))
	}
	cleanup()
	defer cleanup()

	for _, id := range all {
		if err := db.Create(&model.GalgameLocal{ID: id}).Error; err != nil {
			t.Fatalf("seed galgame %d: %v", id, err)
		}
	}
	if err := db.Create(&model.GalgameResource{
		GalgameID: withResource, UserID: 1, Platform: "windows", Language: "ja", Type: "game",
	}).Error; err != nil {
		t.Fatalf("seed resource: %v", err)
	}

	if err := repo.DeleteLocalDraft(bare); err != nil {
		t.Fatalf("DeleteLocalDraft(bare): %v", err)
	}
	if err := repo.DeleteLocalDraft(withResource); err != nil {
		t.Fatalf("DeleteLocalDraft(withResource): %v", err)
	}

	var rows int64
	db.Table("galgame").Where("id = ?", bare).Count(&rows)
	if rows != 0 {
		t.Error("an empty draft row must be gone; nothing else will ever clean it up")
	}
	db.Table("galgame").Where("id = ?", withResource).Count(&rows)
	if rows != 1 {
		t.Error("a row carrying a resource must survive rather than cascade it away")
	}
	db.Table("galgame_resource").Where("galgame_id = ?", withResource).Count(&rows)
	if rows != 1 {
		t.Error("the resource was cascade-deleted")
	}
}

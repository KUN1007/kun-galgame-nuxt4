package repository

import (
	"time"

	"kun-galgame-api/internal/galgame/model"
	"kun-galgame-api/internal/infrastructure/viewstats"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type GalgameRepository struct {
	db *gorm.DB
}

func NewGalgameRepository(db *gorm.DB) *GalgameRepository {
	return &GalgameRepository{db: db}
}

func (r *GalgameRepository) DB() *gorm.DB {
	return r.db
}

type GalgameLocalRow struct {
	ID                 int       `gorm:"column:id"`
	LikeCount          int       `gorm:"column:like_count"`
	FavoriteCount      int       `gorm:"column:favorite_count"`
	View               int       `gorm:"column:view"`
	ResourceUpdateTime time.Time `gorm:"column:resource_update_time"`
	CreatorUserID      *int      `gorm:"column:creator_user_id"`
	Published          bool      `gorm:"column:published"`
}

func (r *GalgameRepository) FindLocal(id int) model.GalgameLocal {
	var row model.GalgameLocal
	r.db.Where("id = ?", id).First(&row)
	return row
}

func (r *GalgameRepository) FindLocalBatch(ids []int) map[int]GalgameLocalRow {
	if len(ids) == 0 {
		return map[int]GalgameLocalRow{}
	}
	var rows []GalgameLocalRow
	r.db.Table("galgame").Select("id, like_count, favorite_count, view, resource_update_time, creator_user_id, published").
		Where("id IN ?", ids).Scan(&rows)
	out := make(map[int]GalgameLocalRow, len(rows))
	for _, row := range rows {
		out[row.ID] = row
	}
	return out
}

// Owner, not actor. Catalog's by-uid claim face answers "every work this user
// TOUCHED" — approving someone else's submission puts it under the reviewer —
// so the profile used to list 13 entries a moderator had merely reviewed, and
// count them toward creator eligibility. The owner lives here.
func (r *GalgameRepository) PublishedIDsByCreator(userID, page, limit int) ([]int, int64, error) {
	base := r.db.Table("galgame").Where("published AND creator_user_id = ?", userID)

	var total int64
	if err := base.Count(&total).Error; err != nil {
		return nil, 0, err
	}

	var ids []int
	err := base.Order("created DESC").
		Offset((page-1)*limit).Limit(limit).
		Pluck("id", &ids).Error
	return ids, total, err
}

func (r *GalgameRepository) CountPublishedByCreatorSince(userID int, since time.Time) int {
	var n int64
	r.db.Table("galgame").
		Where("published AND creator_user_id = ? AND created >= ?", userID, since).
		Count(&n)
	return int(n)
}

func (r *GalgameRepository) IncrementView(id int) {
	r.db.Table("galgame").Where("id = ?", id).
		Update("view", gorm.Expr("view + 1"))
	_ = viewstats.BumpDaily(r.db, viewstats.GalgameDaily, id)
}

func (r *GalgameRepository) PublishLocal(tx *gorm.DB, galgameID int) error {
	return tx.Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "id"}},
		DoUpdates: clause.Assignments(map[string]any{"published": true}),
	}).Create(&model.GalgameLocal{ID: galgameID, Published: true}).Error
}

func (r *GalgameRepository) UnpublishLocal(galgameID int) error {
	return r.db.Model(&model.GalgameLocal{}).Where("id = ?", galgameID).
		UpdateColumn("published", false).Error
}

func (r *GalgameRepository) DeleteLocal(galgameID int) error {
	return r.db.Where("id = ?", galgameID).Delete(&model.GalgameLocal{}).Error
}

func (r *GalgameRepository) EnsureLocalStub(tx *gorm.DB, galgameID int) error {
	return tx.Clauses(clause.OnConflict{DoNothing: true}).
		Create(&model.GalgameLocal{ID: galgameID}).Error
}

func (r *GalgameRepository) SetCreatorIfUnset(tx *gorm.DB, galgameID, userID int) error {
	return tx.Model(&model.GalgameLocal{}).
		Where("id = ? AND creator_user_id IS NULL", galgameID).
		UpdateColumn("creator_user_id", userID).Error
}

func (r *GalgameRepository) Touch(tx *gorm.DB, galgameID int) error {
	if err := r.EnsureLocalStub(tx, galgameID); err != nil {
		return err
	}
	return tx.Model(&model.GalgameLocal{}).Where("id = ?", galgameID).
		UpdateColumn("resource_update_time", time.Now()).Error
}

func (r *GalgameRepository) SubmitLocal(tx *gorm.DB, galgameID, userID int) error {
	if err := r.EnsureLocalStub(tx, galgameID); err != nil {
		return err
	}
	return r.SetCreatorIfUnset(tx, galgameID, userID)
}

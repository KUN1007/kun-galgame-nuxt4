package repository

import (
	"encoding/json"
	"strings"
	"time"

	"kun-galgame-api/internal/galgame/model"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

func pgTextArrayLiteral(items []string) string {
	if len(items) == 0 {
		return "{}"
	}
	parts := make([]string, len(items))
	for i, v := range items {
		escaped := strings.ReplaceAll(v, `\`, `\\`)
		escaped = strings.ReplaceAll(escaped, `"`, `\"`)
		parts[i] = `"` + escaped + `"`
	}
	return "{" + strings.Join(parts, ",") + "}"
}

var onConflictNothing = clause.OnConflict{DoNothing: true}

type ResourceRepository struct {
	db *gorm.DB
}

func NewResourceRepository(db *gorm.DB) *ResourceRepository {
	return &ResourceRepository{db: db}
}

func (r *ResourceRepository) DB() *gorm.DB { return r.db }

// The count and the page must apply the same gate, and it has to be in SQL.
// Dropping the hidden rows only after the page was cut answered an anonymous
// reader 7 cards for a requested 30 while the pager still counted all 37,840
// resources. NULL is "the content-limit sync has not seen this galgame yet" and
// passes, same as on /galgame.
func sfwResources(q *gorm.DB, sfw bool) *gorm.DB {
	if !sfw {
		return q
	}
	return q.Joins("JOIN galgame g ON g.id = galgame_resource.galgame_id").
		Where("g.content_limit IS NULL OR g.content_limit = 'sfw'")
}

func (r *ResourceRepository) CountAll(sfw bool) int64 {
	var total int64
	sfwResources(r.db.Table("galgame_resource"), sfw).Count(&total)
	return total
}

func (r *ResourceRepository) ListPaginated(page, limit int, sfw bool) []model.GalgameResourceRow {
	offset := (page - 1) * limit
	var rows []model.GalgameResourceRow
	sfwResources(r.db.Table("galgame_resource"), sfw).
		Select("galgame_resource.*").
		Order("galgame_resource.created DESC").
		Offset(offset).Limit(limit).
		Scan(&rows)
	return rows
}

func (r *ResourceRepository) FindByID(id int) (model.GalgameResourceRow, bool) {
	var row model.GalgameResourceRow
	if err := r.db.Table("galgame_resource").Where("id = ?", id).Scan(&row).Error; err != nil || row.ID == 0 {
		return row, false
	}
	return row, true
}

func (r *ResourceRepository) IsResourcePublishBanned(galgameID int) bool {
	var banned []bool
	r.db.Table("galgame").Where("id = ?", galgameID).Pluck("resource_publish_banned", &banned)
	return len(banned) > 0 && banned[0]
}

func (r *ResourceRepository) SetResourcePublishBanned(galgameID int, banned bool) error {
	return r.db.Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "id"}},
		DoUpdates: clause.Assignments(map[string]any{"resource_publish_banned": banned}),
	}).Create(&model.GalgameLocal{ID: galgameID, ResourcePublishBanned: banned}).Error
}

func (r *ResourceRepository) FindByGalgameID(galgameID int) []model.GalgameResourceRow {
	var rows []model.GalgameResourceRow
	r.db.Table("galgame_resource").
		Where("galgame_id = ?", galgameID).
		Order("status ASC, RANDOM()").
		Scan(&rows)
	return rows
}

func (r *ResourceRepository) FindRecommendations(galgameID, excludeID, limit int) []model.GalgameResourceRow {
	var rows []model.GalgameResourceRow
	r.db.Table("galgame_resource").
		Where("galgame_id = ? AND id != ?", galgameID, excludeID).
		Order("like_count DESC").
		Limit(limit).
		Scan(&rows)
	return rows
}

func (r *ResourceRepository) FindLinks(resourceID int) []string {
	type linkRow struct {
		URL string `gorm:"column:url"`
	}
	var links []linkRow
	r.db.Table("galgame_resource_link").
		Where("galgame_resource_id = ?", resourceID).
		Scan(&links)
	out := make([]string, len(links))
	for i, l := range links {
		out[i] = l.URL
	}
	return out
}

func (r *ResourceRepository) AggregateByGalgame(galgameID int) []model.ResourceAggregate {
	var aggs []model.ResourceAggregate
	r.db.Table("galgame_resource").
		Select("DISTINCT platform, language, type").
		Where("galgame_id = ?", galgameID).
		Scan(&aggs)
	return aggs
}

func (r *ResourceRepository) IsLikedBy(resourceID, userID int) bool {
	if userID <= 0 {
		return false
	}
	var cnt int64
	r.db.Table("galgame_resource_like").
		Where("galgame_resource_id = ? AND user_id = ?", resourceID, userID).
		Count(&cnt)
	return cnt > 0
}

func (r *ResourceRepository) FindLikedSet(userID int, resourceIDs []int) map[int]bool {
	out := map[int]bool{}
	if userID <= 0 || len(resourceIDs) == 0 {
		return out
	}
	var rows []struct {
		ResourceID int `gorm:"column:galgame_resource_id"`
	}
	r.db.Table("galgame_resource_like").
		Where("user_id = ? AND galgame_resource_id IN ?", userID, resourceIDs).
		Select("galgame_resource_id").
		Scan(&rows)
	for _, row := range rows {
		out[row.ResourceID] = true
	}
	return out
}

func (r *ResourceRepository) FindGalgameLocal(galgameID int) GalgameLocalRow {
	var row GalgameLocalRow
	r.db.Table("galgame").
		Select("id, like_count, favorite_count, view, resource_update_time, creator_user_id, published").
		Where("id = ?", galgameID).Scan(&row)
	return row
}

func (r *ResourceRepository) IncrementView(resourceID int) {
	r.db.Exec("UPDATE galgame_resource SET view = view + 1 WHERE id = ?", resourceID)
}

func (r *ResourceRepository) IncrementDownload(resourceID int) {
	r.db.Exec("UPDATE galgame_resource SET download = download + 1 WHERE id = ?", resourceID)
}

func (r *ResourceRepository) Create(tx *gorm.DB, res *model.GalgameResource) error {
	return tx.Create(res).Error
}

func (r *ResourceRepository) ReplaceProviders(tx *gorm.DB, resourceID int, providers []string) error {
	return tx.Exec(
		"UPDATE galgame_resource SET provider = ?::text[] WHERE id = ?",
		pgTextArrayLiteral(providers), resourceID,
	).Error
}

func (r *ResourceRepository) ReplaceProviderNames(tx *gorm.DB, resourceID int, names []string) error {
	if names == nil {
		names = []string{}
	}
	encoded, err := json.Marshal(names)
	if err != nil {
		return err
	}
	return tx.Exec(
		"UPDATE galgame_resource SET provider_name = ?::jsonb WHERE id = ?",
		string(encoded), resourceID,
	).Error
}

func (r *ResourceRepository) CreateLinks(tx *gorm.DB, resourceID int, urls []string) error {
	if len(urls) == 0 {
		return nil
	}
	links := make([]model.GalgameResourceLink, len(urls))
	for i, u := range urls {
		links[i] = model.GalgameResourceLink{GalgameResourceID: resourceID, URL: u}
	}
	return tx.Clauses(onConflictNothing).Create(&links).Error
}

func (r *ResourceRepository) DeleteLinks(tx *gorm.DB, resourceID int) error {
	return tx.Where("galgame_resource_id = ?", resourceID).
		Delete(&model.GalgameResourceLink{}).Error
}

func (r *ResourceRepository) UpdateFields(tx *gorm.DB, resourceID int, fields map[string]any) error {
	if len(fields) == 0 {
		return nil
	}
	return tx.Table("galgame_resource").Where("id = ?", resourceID).
		Updates(fields).Error
}

func (r *ResourceRepository) UpdateStatus(tx *gorm.DB, resourceID, status int) error {
	return tx.Table("galgame_resource").Where("id = ?", resourceID).
		Update("status", status).Error
}

func (r *ResourceRepository) DeleteByID(tx *gorm.DB, resourceID int) error {
	return tx.Where("id = ?", resourceID).Delete(&model.GalgameResource{}).Error
}

func (r *ResourceRepository) FindLike(tx *gorm.DB, resourceID, userID int) (model.GalgameResourceLike, bool) {
	var like model.GalgameResourceLike
	err := tx.Where("galgame_resource_id = ? AND user_id = ?", resourceID, userID).
		First(&like).Error
	if err != nil {
		return like, false
	}
	return like, true
}

func (r *ResourceRepository) CreateLike(tx *gorm.DB, resourceID, userID int) error {
	return tx.Create(&model.GalgameResourceLike{
		GalgameResourceID: resourceID, UserID: userID,
	}).Error
}

func (r *ResourceRepository) DeleteLike(tx *gorm.DB, like model.GalgameResourceLike) error {
	return tx.Delete(&like).Error
}

func (r *ResourceRepository) AdjustLikeCount(tx *gorm.DB, resourceID, delta int) error {
	return tx.Table("galgame_resource").Where("id = ?", resourceID).
		Update("like_count", gorm.Expr("like_count + ?", delta)).Error
}

func (r *ResourceRepository) AdjustLocalResourceCount(tx *gorm.DB, galgameID, delta int) error {
	return tx.Table("galgame").Where("id = ?", galgameID).
		Update("resource_count", gorm.Expr("resource_count + ?", delta)).Error
}

func (r *ResourceRepository) TouchGalgameUpdated(tx *gorm.DB, galgameID int) error {
	return tx.Table("galgame").Where("id = ?", galgameID).
		UpdateColumn("resource_update_time", time.Now()).Error
}

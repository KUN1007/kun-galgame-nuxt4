package repository

import (
	"kun-galgame-api/internal/website/model"

	"gorm.io/gorm"
)

type TagGroupRepository struct {
	db *gorm.DB
}

func NewTagGroupRepository(db *gorm.DB) *TagGroupRepository {
	return &TagGroupRepository{db: db}
}

func (r *TagGroupRepository) FindAll() []model.GalgameWebsiteTagGroup {
	var groups []model.GalgameWebsiteTagGroup
	r.db.Order("sort_order ASC, id ASC").Find(&groups)
	return groups
}

func (r *TagGroupRepository) Create(group *model.GalgameWebsiteTagGroup) error {
	return r.db.Create(group).Error
}

func (r *TagGroupRepository) UpdateFields(id int, updates map[string]any) error {
	return r.db.Model(&model.GalgameWebsiteTagGroup{}).Where("id = ?", id).Updates(updates).Error
}

// The tags survive their group: group_id is ON DELETE SET NULL, so they fall
// back into the ungrouped bucket instead of disappearing from the selector.
func (r *TagGroupRepository) DeleteByID(id int) error {
	return r.db.Delete(&model.GalgameWebsiteTagGroup{}, id).Error
}

func (r *TagGroupRepository) TagCounts() map[int]int {
	var rows []struct {
		GroupID int
		Total   int
	}
	r.db.Table("galgame_website_tag").
		Select("group_id, COUNT(*) AS total").
		Where("group_id IS NOT NULL").
		Group("group_id").Scan(&rows)
	out := make(map[int]int, len(rows))
	for _, row := range rows {
		out[row.GroupID] = row.Total
	}
	return out
}

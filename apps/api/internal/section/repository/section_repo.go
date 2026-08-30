package repository

import (
	"time"

	"gorm.io/gorm"

	"kun-galgame-api/pkg/miniapp"
)

type SectionRepository struct {
	db *gorm.DB
}

func NewSectionRepository(db *gorm.DB) *SectionRepository {
	return &SectionRepository{db: db}
}

type SectionTopicRow struct {
	ID           int
	Title        string
	Content      string
	View         int
	LikeCount    int
	ReplyCount   int
	Status       int
	IsNSFW       bool
	BestAnswerID *int
	UserID       int
	Created      time.Time
}

type SectionStatRow struct {
	SectionID   int    `gorm:"column:section_id"`
	SectionName string `gorm:"column:section_name"`
	TopicCount  int64  `gorm:"column:topic_count"`
	ViewCount   int64  `gorm:"column:view_count"`
}

type LatestTopicRow struct {
	ID      int    `gorm:"column:id"`
	Title   string `gorm:"column:title"`
	Created string `gorm:"column:created"`
	UserID  int    `gorm:"column:user_id"`
}

func (r *SectionRepository) FindSectionTopics(
	section, sortOrder string, page, limit int,
) (rows []SectionTopicRow, total int64, err error) {
	query := r.db.Table("topic t").
		Select(`t.id, t.title, SUBSTRING(t.content, 1, 233) AS content,
			t.view, t.like_count, t.reply_count, t.status, t.is_nsfw,
			t.best_answer_id, t.user_id, t.created`).
		Joins("JOIN topic_section_relation tsr ON tsr.topic_id = t.id").
		Joins("JOIN topic_section ts ON ts.id = tsr.topic_section_id").
		Where("ts.name = ? AND t.status != 1", section)

	if err = query.Count(&total).Error; err != nil {
		return
	}

	err = query.Order("t.created " + sortOrder).
		Offset((page - 1) * limit).
		Limit(limit).
		Find(&rows).Error
	return
}

func (r *SectionRepository) FindCategoryStats(category string) ([]SectionStatRow, error) {
	var rows []SectionStatRow
	err := r.db.Raw(`
		SELECT ts.id AS section_id, ts.name AS section_name,
			COUNT(DISTINCT t.id) AS topic_count,
			COALESCE(SUM(t.view), 0) AS view_count
		FROM topic_section ts
		JOIN topic_section_relation tsr ON tsr.topic_section_id = ts.id
		JOIN topic t ON t.id = tsr.topic_id AND t.status != 1
			AND t.category = ?
		GROUP BY ts.id, ts.name
		ORDER BY ts.id
	`, category).Scan(&rows).Error
	return rows, err
}

func (r *SectionRepository) FindLatestTopicsInSection(sectionID int, category string, limit int) []LatestTopicRow {
	var rows []LatestTopicRow
	r.db.Raw(`
		SELECT t.id, t.title, t.created, t.user_id
		FROM topic t
		JOIN topic_section_relation tsr ON tsr.topic_id = t.id
		WHERE tsr.topic_section_id = ? AND t.status != 1
			AND t.category = ?
		ORDER BY t.created DESC LIMIT ?
	`, sectionID, category, limit).Scan(&rows)
	return rows
}

func (r *SectionRepository) FindTopicMiniApps(topicIDs []int) map[int][]string {
	return miniapp.ByTopic(r.db, topicIDs)
}

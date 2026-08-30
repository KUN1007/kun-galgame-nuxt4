package repository

import (
	"time"

	"gorm.io/gorm"

	"kun-galgame-api/pkg/miniapp"
)

type SearchRepository struct {
	db *gorm.DB
}

func NewSearchRepository(db *gorm.DB) *SearchRepository {
	return &SearchRepository{db: db}
}

type TopicRow struct {
	ID               int
	Title            string
	View             int
	Status           int
	LikeCount        int
	ReplyCount       int
	CommentCount     int
	StatusUpdateTime time.Time
	UserID           int
	IsNSFW           bool
	BestAnswerID     *int
	UpvoteTime       *time.Time
}

type TopicSectionRow struct {
	TopicID     int    `gorm:"column:topic_id"`
	SectionName string `gorm:"column:name"`
}

type ReplyRow struct {
	ID          int
	TopicID     int
	TopicTitle  string
	TopicUserID int
	Content     string
	Floor       int
	UserID      int
	Created     time.Time
}

type CommentRow struct {
	ID          int
	TopicID     int
	TopicTitle  string
	TopicUserID int
	Content     string
	UserID      int
	Created     time.Time
}

func (r *SearchRepository) SearchTopics(keywords []string, page, limit int) (rows []TopicRow, total int64) {
	query := r.db.Table("topic t").
		Select(`t.id, t.title, t.view, t.status, t.like_count, t.reply_count,
			t.comment_count, t.status_update_time, t.user_id,
			t.is_nsfw, t.best_answer_id, t.upvote_time`).
		Where("t.status != 1")
	for _, kw := range keywords {
		like := "%" + kw + "%"
		query = query.Where("(t.title ILIKE ? OR t.content ILIKE ? OR t.category ILIKE ?)",
			like, like, like)
	}

	query.Count(&total)
	query.Order("t.status_update_time DESC").
		Offset((page - 1) * limit).Limit(limit).
		Find(&rows)
	return
}

func (r *SearchRepository) FindTopicSections(topicIDs []int) []TopicSectionRow {
	if len(topicIDs) == 0 {
		return nil
	}
	var rows []TopicSectionRow
	r.db.Table("topic_section_relation tsr").
		Select("tsr.topic_id, ts.name").
		Joins("JOIN topic_section ts ON ts.id = tsr.topic_section_id").
		Where("tsr.topic_id IN ?", topicIDs).
		Find(&rows)
	return rows
}

func (r *SearchRepository) FindTopicMiniApps(topicIDs []int) map[int][]string {
	return miniapp.ByTopic(r.db, topicIDs)
}

func (r *SearchRepository) SearchReplies(keywords []string, page, limit int) (rows []ReplyRow, total int64) {
	query := r.db.Table("topic_reply r").
		Select(`r.id, r.topic_id, t.title AS topic_title, t.user_id AS topic_user_id,
			SUBSTRING(COALESCE(r.content, ''), 1, 233) AS content, r.floor,
			r.user_id, r.created`).
		Joins("LEFT JOIN topic t ON t.id = r.topic_id").
		Where("r.status = 0")
	for _, kw := range keywords {
		query = query.Where("r.content ILIKE ?", "%"+kw+"%")
	}

	query.Count(&total)
	query.Order("r.created DESC").
		Offset((page - 1) * limit).Limit(limit).
		Find(&rows)
	return
}

func (r *SearchRepository) SearchComments(keywords []string, page, limit int) (rows []CommentRow, total int64) {
	query := r.db.Table("topic_comment c").
		Select(`c.id, c.topic_id, t.title AS topic_title, t.user_id AS topic_user_id,
			SUBSTRING(c.content, 1, 233) AS content,
			c.user_id, c.created`).
		Joins("LEFT JOIN topic t ON t.id = c.topic_id").
		Where("c.status = 0")
	for _, kw := range keywords {
		query = query.Where("c.content ILIKE ?", "%"+kw+"%")
	}

	query.Count(&total)
	query.Order("c.created DESC").
		Offset((page - 1) * limit).Limit(limit).
		Find(&rows)
	return
}

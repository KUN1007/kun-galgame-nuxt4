package repository

import (
	"time"

	"kun-galgame-api/internal/constants"
	"kun-galgame-api/internal/topic/model"

	"gorm.io/gorm"
)

type TopicListRepository struct {
	db *gorm.DB
}

func NewTopicListRepository(db *gorm.DB) *TopicListRepository {
	return &TopicListRepository{db: db}
}

type TopicCardRow struct {
	ID               int
	Title            string
	View             int
	Status           int
	IsNSFW           bool
	LikeCount        int
	ReplyCount       int
	CommentCount     int
	BestAnswerID     *int
	StatusUpdateTime time.Time
	Created          time.Time
	UpvoteTime       *time.Time
	CoverImages      model.ImageTokens
	UserID           int
	UserName         string
	UserAvatar       string
}

func (r *TopicListRepository) FindList(
	page, limit int,
	sortField, sortOrder, category string,
	isNSFW, authenticated bool,
) ([]TopicCardRow, int64, error) {
	var rows []TopicCardRow
	var total int64

	query := r.db.Table("topic").
		Select(`topic.id, topic.title, topic.view, topic.status,
			topic.is_nsfw, topic.like_count, topic.reply_count,
			topic.comment_count, topic.best_answer_id,
			topic.status_update_time, topic.created, topic.upvote_time,
			topic.cover_images, topic.user_id`).
		Where("topic.status != 1").
		Where(SharedListPredicate("topic", authenticated))

	if !isNSFW {
		query = query.Where("topic.is_nsfw = false")
	}
	if category != "" && category != "all" {
		query = query.Where("topic.category = ?", category)
	}

	query.Count(&total)

	query = query.Order(topicOrderCol(sortField) + " " + sortOrder).
		Offset((page - 1) * limit).
		Limit(limit)

	err := query.Find(&rows).Error
	return rows, total, err
}

func topicOrderCol(sortField string) string {
	if sortField == "view_1d" {
		return "COALESCE((SELECT SUM(d.count) FROM topic_view_daily d " +
			"WHERE d.entity_id = topic.id AND d.day = CURRENT_DATE), 0)"
	}
	if col, ok := constants.ValidTopicSortFields[sortField]; ok {
		return "topic." + col
	}
	if col, ok := constants.ValidTopicCountSortFields[sortField]; ok {
		return "topic." + col
	}
	return "topic.created"
}

func (r *TopicListRepository) FindResourceList(
	page, limit int,
	sortField, sortOrder, category string,
	isNSFW, authenticated bool,
) ([]TopicCardRow, int64, error) {
	var rows []TopicCardRow
	var total int64

	query := r.db.Table("topic").
		Select(`topic.id, topic.title, topic.view, topic.status,
			topic.is_nsfw, topic.like_count, topic.reply_count,
			topic.comment_count, topic.best_answer_id,
			topic.status_update_time, topic.created, topic.upvote_time,
			topic.cover_images, topic.user_id`).
		Joins(`JOIN topic_section_relation tsr ON tsr.topic_id = topic.id`).
		Joins(`JOIN topic_section ts ON ts.id = tsr.topic_section_id`).
		Where("topic.status != 1").
		Where(SharedListPredicate("topic", authenticated)).
		Where("ts.name IN ?", []string{"g-seeking", "g-other", "t-help"}).
		Group("topic.id")

	if !isNSFW {
		query = query.Where("topic.is_nsfw = false")
	}
	if category != "" && category != "all" {
		query = query.Where("topic.category = ?", category)
	}

	query.Count(&total)

	query = query.Order(topicOrderCol(sortField) + " " + sortOrder).
		Offset((page - 1) * limit).
		Limit(limit)

	err := query.Find(&rows).Error
	return rows, total, err
}

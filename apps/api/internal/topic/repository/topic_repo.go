package repository

import (
	"time"

	"kun-galgame-api/internal/infrastructure/viewstats"
	"kun-galgame-api/internal/topic/model"
	"kun-galgame-api/pkg/miniapp"

	"gorm.io/gorm"
)

type TopicRepository struct {
	db *gorm.DB
}

func NewTopicRepository(db *gorm.DB) *TopicRepository {
	return &TopicRepository{db: db}
}

func (r *TopicRepository) DB() *gorm.DB {
	return r.db
}

func (r *TopicRepository) FindByID(id int) (*model.Topic, error) {
	var topic model.Topic
	err := r.db.First(&topic, id).Error
	return &topic, err
}

func (r *TopicRepository) FindReplyByID(id int) (*model.TopicReply, error) {
	var reply model.TopicReply
	err := r.db.First(&reply, id).Error
	return &reply, err
}

func (r *TopicRepository) UpdateFields(id int, fields map[string]any) error {
	return r.db.Model(&model.Topic{}).Where("id = ?", id).Updates(fields).Error
}

func (r *TopicRepository) IncrementView(id int) error {
	err := r.db.Model(&model.Topic{}).Where("id = ?", id).
		Update("view", gorm.Expr("view + 1")).Error
	_ = viewstats.BumpDaily(r.db, viewstats.TopicDaily, id)
	return err
}

func (r *TopicRepository) HasUserLiked(userID, topicID int) (bool, error) {
	var count int64
	err := r.db.Model(&model.TopicReaction{}).
		Where("user_id = ? AND topic_id = ? AND reaction = 'like'", userID, topicID).Count(&count).Error
	return count > 0, err
}

func (r *TopicRepository) HasUserDisliked(userID, topicID int) (bool, error) {
	var count int64
	err := r.db.Model(&model.TopicReaction{}).
		Where("user_id = ? AND topic_id = ? AND reaction = 'dislike'", userID, topicID).Count(&count).Error
	return count > 0, err
}

func (r *TopicRepository) HasUserFavorited(userID, topicID int) (bool, error) {
	var count int64
	err := r.db.Model(&model.TopicFavorite{}).Where("user_id = ? AND topic_id = ?", userID, topicID).Count(&count).Error
	return count > 0, err
}

func (r *TopicRepository) HasUserUpvoted(userID, topicID int) (bool, error) {
	var count int64
	err := r.db.Model(&model.TopicUpvote{}).Where("user_id = ? AND topic_id = ?", userID, topicID).Count(&count).Error
	return count > 0, err
}

func (r *TopicRepository) CountTodayTopicsByUser(tx *gorm.DB, userID int) (int64, error) {
	var count int64
	oneDayAgo := time.Now().Add(-24 * time.Hour)
	err := tx.Model(&model.Topic{}).
		Where("user_id = ? AND created >= ?", userID, oneDayAgo).
		Count(&count).Error
	return count, err
}

func (r *TopicRepository) FindTopicMiniApps(topicIDs []int) map[int][]string {
	return miniapp.ByTopic(r.db, topicIDs)
}

func (r *TopicRepository) FindByIDTx(tx *gorm.DB, topicID int) (*model.Topic, error) {
	var topic model.Topic
	err := tx.First(&topic, topicID).Error
	return &topic, err
}

func (r *TopicRepository) CreateTopic(tx *gorm.DB, topic *model.Topic) error {
	return tx.Create(topic).Error
}

func (r *TopicRepository) UpdateTopicFields(tx *gorm.DB, topicID int, fields map[string]any) error {
	return tx.Model(&model.Topic{}).Where("id = ?", topicID).Updates(fields).Error
}

func (r *TopicRepository) TouchStatusUpdateTime(tx *gorm.DB, topicID int, t time.Time) error {
	return tx.Model(&model.Topic{}).
		Where("id = ? AND created > ?", topicID, model.BumpCutoff(t)).
		Updates(map[string]any{"status_update_time": t}).Error
}

func (r *TopicRepository) FindTopicFavorite(tx *gorm.DB, userID, topicID int) (*model.TopicFavorite, error) {
	var existing model.TopicFavorite
	err := tx.Where("user_id = ? AND topic_id = ?", userID, topicID).First(&existing).Error
	return &existing, err
}

func (r *TopicRepository) CreateTopicFavorite(tx *gorm.DB, userID, topicID int) error {
	return tx.Create(&model.TopicFavorite{UserID: userID, TopicID: topicID}).Error
}

func (r *TopicRepository) DeleteTopicFavorite(tx *gorm.DB, fav *model.TopicFavorite) error {
	return tx.Delete(fav).Error
}

func (r *TopicRepository) UserTopicInteractions(userID int) ([]int, map[int][]string, error) {
	favorited := []int{}
	if err := r.db.Model(&model.TopicFavorite{}).
		Where("user_id = ?", userID).Pluck("topic_id", &favorited).Error; err != nil {
		return nil, nil, err
	}
	var rows []struct {
		TopicID  int    `gorm:"column:topic_id"`
		Reaction string `gorm:"column:reaction"`
	}
	if err := r.db.Table("topic_reaction").
		Select("topic_id, reaction").
		Where("user_id = ?", userID).Scan(&rows).Error; err != nil {
		return nil, nil, err
	}
	reactions := map[int][]string{}
	for _, row := range rows {
		reactions[row.TopicID] = append(reactions[row.TopicID], row.Reaction)
	}
	return favorited, reactions, nil
}

func (r *TopicRepository) CreateTopicUpvote(tx *gorm.DB, userID, topicID int, description string) error {
	return tx.Create(&model.TopicUpvote{UserID: userID, TopicID: topicID, Description: description}).Error
}

type TopicUpvoteRow struct {
	ID          int       `gorm:"column:id"`
	UserID      int       `gorm:"column:user_id"`
	Description string    `gorm:"column:description"`
	Created     time.Time `gorm:"column:created"`
}

func (r *TopicRepository) FetchTopicUpvotes(topicID, limit int) ([]TopicUpvoteRow, error) {
	var rows []TopicUpvoteRow
	err := r.db.Table("topic_upvote").
		Select("id, user_id, description, created").
		Where("topic_id = ?", topicID).
		Order("created DESC, id DESC").
		Limit(limit).
		Scan(&rows).Error
	return rows, err
}

func (r *TopicRepository) AdjustLikeCount(tx *gorm.DB, topicID, delta int) error {
	return tx.Model(&model.Topic{}).Where("id = ?", topicID).
		Update("like_count", gorm.Expr("like_count + ?", delta)).Error
}

func (r *TopicRepository) AdjustDislikeCount(tx *gorm.DB, topicID, delta int) error {
	return tx.Model(&model.Topic{}).Where("id = ?", topicID).
		Update("dislike_count", gorm.Expr("dislike_count + ?", delta)).Error
}

func (r *TopicRepository) AdjustFavoriteCount(tx *gorm.DB, topicID, delta int) error {
	return tx.Model(&model.Topic{}).Where("id = ?", topicID).
		Update("favorite_count", gorm.Expr("favorite_count + ?", delta)).Error
}

func (r *TopicRepository) ApplyUpvoteCountAndTime(tx *gorm.DB, topicID int, t time.Time) error {
	return tx.Model(&model.Topic{}).Where("id = ?", topicID).Updates(map[string]any{
		"upvote_count": gorm.Expr("upvote_count + 1"),
		"upvote_time":  &t,
		"status_update_time": gorm.Expr(
			"CASE WHEN created > ? THEN ? ELSE status_update_time END", model.BumpCutoff(t), t),
	}).Error
}

type TopicAuthorUser struct {
	ID          int    `json:"id"`
	Name        string `json:"name"`
	Avatar      string `json:"avatar"`
	Moemoepoint int    `json:"moemoepoint"`
}

package repository

import (
	"fmt"
	"gorm.io/gorm"
	"kun-galgame-api/internal/admin/dto"
	"strings"
	"time"
)

type TopicAdminRepository struct{ db *gorm.DB }

func NewTopicAdminRepository(db *gorm.DB) *TopicAdminRepository { return &TopicAdminRepository{db: db} }

type HiddenRow struct {
	ID                        int
	Title, HiddenBy           string
	ReplyCount                int
	StatusUpdateTime, Created time.Time
	UserID                    int
}

func (r *TopicAdminRepository) ListHidden(page, limit int, hiddenBy, keywords string) ([]HiddenRow, int64, error) {
	q := r.db.Table("topic").Where("status = ?", 1)
	if hiddenBy != "" {
		q = q.Where("hidden_by = ?", hiddenBy)
	}
	if keywords != "" {
		esc := strings.NewReplacer("\\", "\\\\", "%", "\\%", "_", "\\_").Replace(keywords)
		q = q.Where("title ILIKE ?", "%"+esc+"%")
	}
	var total int64
	if err := q.Count(&total).Error; err != nil {
		return nil, 0, err
	}
	var rows []HiddenRow
	err := q.Select("id,title,hidden_by,reply_count,status_update_time,created,user_id").Order("status_update_time DESC").Offset((page - 1) * limit).Limit(limit).Find(&rows).Error
	return rows, total, err
}
func (r *TopicAdminRepository) PurgeStats(id int) (dto.TopicPurgeStats, error) {
	var s dto.TopicPurgeStats
	var row struct {
		ID       int
		Title    string
		Status   int
		HiddenBy string
		UserID   int
	}
	if err := r.db.Table("topic").Select("id,title,status,hidden_by,user_id").Where("id = ?", id).Scan(&row).Error; err != nil {
		return s, err
	}
	if row.ID == 0 {
		return s, gorm.ErrRecordNotFound
	}
	s.ID = row.ID
	s.Title = row.Title
	s.Status = row.Status
	s.HiddenBy = row.HiddenBy
	s.Replies = r.count("topic_reply", id)
	s.Comments = r.count("topic_comment", id)
	s.Polls = r.count("topic_poll", id)
	s.Lotteries = r.count("topic_lottery", id)
	r.db.Table("topic_lottery").Where("topic_id = ? AND status = ?", id, "drawn").Count(&s.DrawnLotteries)
	s.Favorites = r.count("topic_favorite", id)
	s.User.ID = row.UserID
	return s, nil
}
func (r *TopicAdminRepository) count(t string, id int) int64 {
	var n int64
	r.db.Table(t).Where("topic_id = ?", id).Count(&n)
	return n
}
func (r *TopicAdminRepository) Delete(id int) (dto.TopicPurgeStats, error) {
	var s dto.TopicPurgeStats
	err := r.db.Transaction(func(tx *gorm.DB) error {
		var row struct {
			ID       int
			Title    string
			Status   int
			HiddenBy string
			UserID   int
		}
		if err := tx.Table("topic").Select("id,title,status,hidden_by,user_id").Where("id = ?", id).Scan(&row).Error; err != nil {
			return err
		}
		if row.ID == 0 {
			return gorm.ErrRecordNotFound
		}
		s.ID = row.ID
		s.Title = row.Title
		s.Status = row.Status
		s.HiddenBy = row.HiddenBy
		s.User.ID = row.UserID
		for _, x := range []struct {
			t string
			p *int64
		}{{"topic_reply", &s.Replies}, {"topic_comment", &s.Comments}, {"topic_poll", &s.Polls}, {"topic_lottery", &s.Lotteries}, {"topic_favorite", &s.Favorites}} {
			if err := tx.Table(x.t).Where("topic_id = ?", id).Count(x.p).Error; err != nil {
				return err
			}
		}
		if err := tx.Table("topic_lottery").Where("topic_id = ? AND status = ?", id, "drawn").Count(&s.DrawnLotteries).Error; err != nil {
			return err
		}
		// topic_view_daily has no FK to topic (migration 050 created it
		// without one), so the cascade on topic cannot reach it. Its key
		// column is entity_id, not topic_id — the viewstats bucket tables
		// share one shape across domains.
		if err := tx.Exec("DELETE FROM topic_view_daily WHERE entity_id = ?", id).Error; err != nil {
			return err
		}
		// BuildTopicLink produces only '/topic/<id>' and '/topic/<id>?...';
		// LIKE '/topic/<id>%' would also match topic 1234 when deleting 123.
		link := fmt.Sprintf("/topic/%d", id)
		if err := tx.Exec("DELETE FROM message WHERE link = ? OR link LIKE ?", link, link+"?%").Error; err != nil {
			return err
		}
		return tx.Exec("DELETE FROM topic WHERE id = ?", id).Error
	})
	return s, err
}

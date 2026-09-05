package repository

import (
	"gorm.io/gorm"
	"kun-galgame-api/internal/topic/model"
)

func (r *TopicRepository) FindAccessGrants(topicID int) ([]model.TopicAccessGrant, error) {
	var rows []model.TopicAccessGrant
	err := r.db.Where("topic_id = ?", topicID).Order("subject_type, subject_value").Find(&rows).Error
	return rows, err
}

func (r *TopicRepository) ReplaceAccessGrants(tx *gorm.DB, topicID int, grants []model.TopicAccessGrant) error {
	if err := tx.Where("topic_id = ?", topicID).Delete(&model.TopicAccessGrant{}).Error; err != nil {
		return err
	}
	if len(grants) == 0 {
		return nil
	}
	rows := make([]model.TopicAccessGrant, len(grants))
	copy(rows, grants)
	for i := range rows {
		rows[i].TopicID = topicID
	}
	return tx.Create(&rows).Error
}

func SharedListPredicate(table string, authenticated bool) string {
	column := "access_scope"
	if table != "" {
		column = table + "." + column
	}
	// Role and user grants never make a topic visible in shared lists.
	if authenticated {
		return column + " IN ('public','login')"
	}
	return column + " = 'public'"
}

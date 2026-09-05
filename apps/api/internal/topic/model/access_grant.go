package model

type TopicAccessGrant struct {
	TopicID      int    `gorm:"column:topic_id;primaryKey" json:"topic_id"`
	SubjectType  string `gorm:"column:subject_type;primaryKey" json:"subject_type"`
	SubjectValue string `gorm:"column:subject_value;primaryKey" json:"subject_value"`
}

func (TopicAccessGrant) TableName() string { return "topic_access_grant" }

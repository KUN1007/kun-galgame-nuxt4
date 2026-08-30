package model

type UserBrief struct {
	ID     int    `json:"id"`
	Name   string `json:"name"`
	Avatar string `json:"avatar"`
}

type UserStats struct {
	Topic                  int64 `gorm:"column:topic"`
	TopicPoll              int64 `gorm:"column:topic_poll"`
	TopicLottery           int64 `gorm:"column:topic_lottery"`
	ReplyCreated           int64 `gorm:"column:reply_created"`
	CommentCreated         int64 `gorm:"column:comment_created"`
	GalgameComment         int64 `gorm:"column:galgame_comment"`
	GalgameRating          int64 `gorm:"column:galgame_rating"`
	GalgameResource        int64 `gorm:"column:galgame_resource"`
	GalgameToolset         int64 `gorm:"column:galgame_toolset"`
	GalgameToolsetResource int64 `gorm:"column:galgame_toolset_resource"`
	Upvote                 int64 `gorm:"column:upvote"`
	Like                   int64 `gorm:"column:like"`
	Dislike                int64 `gorm:"column:dislike"`
	DailyTopicCount        int64 `gorm:"column:daily_topic_count"`
}

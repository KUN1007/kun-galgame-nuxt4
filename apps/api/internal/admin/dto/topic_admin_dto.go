package dto

import "time"

type TopicAdminUser struct {
	ID     int    `json:"id"`
	Name   string `json:"name"`
	Avatar string `json:"avatar"`
}
type HiddenTopic struct {
	ID               int            `json:"id"`
	Title            string         `json:"title"`
	HiddenBy         string         `json:"hidden_by"`
	ReplyCount       int            `json:"reply_count"`
	StatusUpdateTime time.Time      `json:"status_update_time"`
	Created          time.Time      `json:"created"`
	User             TopicAdminUser `json:"user"`
}
type HiddenTopicList struct {
	Topics []HiddenTopic `json:"topics"`
	Total  int64         `json:"total"`
}
type TopicPurgeStats struct {
	ID             int            `json:"id"`
	Title          string         `json:"title"`
	Status         int            `json:"status"`
	HiddenBy       string         `json:"hidden_by"`
	User           TopicAdminUser `json:"user"`
	Replies        int64          `json:"replies"`
	Comments       int64          `json:"comments"`
	Polls          int64          `json:"polls"`
	Lotteries      int64          `json:"lotteries"`
	DrawnLotteries int64          `json:"drawn_lotteries"`
	Favorites      int64          `json:"favorites"`
}

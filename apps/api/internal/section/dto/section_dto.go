package dto

import "time"

type SectionTopicsRequest struct {
	Section   string `query:"section" validate:"required"`
	Page      int    `query:"page" validate:"min=1"`
	Limit     int    `query:"limit" validate:"min=1,max=30"`
	SortOrder string `query:"sort_order" validate:"required,oneof=asc desc"`
}

type CategoriesRequest struct {
	Category string `query:"category" validate:"required"`
}

type UserBrief struct {
	ID     int    `json:"id"`
	Name   string `json:"name"`
	Avatar string `json:"avatar"`
}

type SectionTopicItem struct {
	ID            int       `json:"id"`
	Title         string    `json:"title"`
	Content       string    `json:"content"`
	View          int       `json:"view"`
	LikeCount     int       `json:"like_count"`
	ReplyCount    int       `json:"reply_count"`
	HasBestAnswer bool      `json:"has_best_answer"`
	IsNSFW        bool      `json:"is_nsfw_topic"`
	MiniApps      []string  `json:"mini_apps"`
	User          UserBrief `json:"user"`
	Created       time.Time `json:"created"`
}

type SectionTopicsResponse struct {
	Topics []SectionTopicItem `json:"topics"`
	Total  int64              `json:"total"`
}

type LatestTopic struct {
	ID      int    `json:"id"`
	Title   string `json:"title"`
	Created string `json:"created"`
}

type SectionStat struct {
	ID          int          `json:"id"`
	Name        string       `json:"name"`
	TopicCount  int64        `json:"topic_count"`
	ViewCount   int64        `json:"view_count"`
	LatestTopic *LatestTopic `json:"latest_topic"`
}

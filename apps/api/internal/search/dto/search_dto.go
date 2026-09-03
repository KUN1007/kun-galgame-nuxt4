package dto

import (
	"time"

	galgameDto "kun-galgame-api/internal/galgame/dto"
)

type SearchRequest struct {
	Keywords string `query:"keywords" validate:"required,max=107"`
	Type     string `query:"type" validate:"required,oneof=topic galgame user reply comment"`
	Page     int    `query:"page" validate:"min=1"`
	Limit    int    `query:"limit" validate:"min=1,max=12"`
}

type UserBrief struct {
	ID     int    `json:"id"`
	Name   string `json:"name"`
	Avatar string `json:"avatar"`
}

type TopicItem struct {
	ID               int        `json:"id"`
	Title            string     `json:"title"`
	View             int        `json:"view"`
	Status           int        `json:"status"`
	LikeCount        int        `json:"like_count"`
	ReplyCount       int        `json:"reply_count"`
	CommentCount     int        `json:"comment_count"`
	HasBestAnswer    bool       `json:"has_best_answer"`
	MiniApps         []string   `json:"mini_apps"`
	IsNSFWTopic      bool       `json:"is_nsfw_topic"`
	Section          []string   `json:"section"`
	UpvoteTime       *time.Time `json:"upvote_time"`
	StatusUpdateTime time.Time  `json:"status_update_time"`
	User             UserBrief  `json:"user"`
}

type UserItem struct {
	ID          int       `json:"id"`
	Name        string    `json:"name"`
	Avatar      string    `json:"avatar"`
	Bio         string    `json:"bio"`
	Moemoepoint int       `json:"moemoepoint"`
	Created     time.Time `json:"created"`
}

type ReplyItem struct {
	ID         int       `json:"id"`
	TopicID    int       `json:"topic_id"`
	TopicTitle string    `json:"topic_title"`
	Content    string    `json:"content"`
	Floor      int       `json:"floor"`
	User       UserBrief `json:"user"`
	Created    time.Time `json:"created"`
}

type CommentItem struct {
	ID         int       `json:"id"`
	TopicID    int       `json:"topic_id"`
	TopicTitle string    `json:"topic_title"`
	Content    string    `json:"content"`
	User       UserBrief `json:"user"`
	Created    time.Time `json:"created"`
}

type QuickSearchRequest struct {
	Keywords string `query:"keywords" validate:"required,max=107"`
}

type QuickSearchTotals struct {
	Topic   int64 `json:"topic"`
	Galgame int64 `json:"galgame"`
	User    int64 `json:"user"`
}

type QuickSearchResult struct {
	Topics   []TopicItem              `json:"topics"`
	Galgames []galgameDto.GalgameCard `json:"galgames"`
	Users    []UserItem               `json:"users"`
	Totals   QuickSearchTotals        `json:"totals"`
}

type PaginatedResult[T any] struct {
	Items []T
	Total int64
}

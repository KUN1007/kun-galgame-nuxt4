package dto

import (
	"time"

	galgameDto "kun-galgame-api/internal/galgame/dto"
	toolsetDto "kun-galgame-api/internal/toolset/dto"
)

type SearchRequest struct {
	Keywords string `query:"keywords" validate:"required,max=107"`
	Type     string `query:"type" validate:"required,oneof=topic galgame resource user reply comment"`
	Page     int    `query:"page" validate:"min=1"`
	Limit    int    `query:"limit" validate:"min=1,max=12"`
}

type OverviewRequest struct {
	Keywords string `query:"keywords" validate:"required,max=107"`
}

type EntitySearchRequest struct {
	Keywords string `query:"keywords" validate:"required,max=107"`
	Family   string `query:"family" validate:"omitempty,oneof=character company staff tag"`
	Limit    int    `query:"limit" validate:"min=1,max=60"`
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
	ID     int      `json:"id"`
	Name   string   `json:"name"`
	Avatar string   `json:"avatar"`
	Bio    string   `json:"bio"`
	Roles  []string `json:"roles"`
	// OAuth owns the registration date and /users/search may answer without it,
	// so this is a pointer: a zero time.Time serialises as year 1 and the card
	// printed 0001年1月1日 for every account.
	Created    *time.Time `json:"created"`
	TopicCount int        `json:"topic_count"`
	ReplyCount int        `json:"reply_count"`
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

type OverviewTotals struct {
	Topic    int64 `json:"topic"`
	Galgame  int64 `json:"galgame"`
	Entity   int64 `json:"entity"`
	Resource int64 `json:"resource"`
	User     int64 `json:"user"`
	Reply    int64 `json:"reply"`
	Comment  int64 `json:"comment"`
	Toolset  int64 `json:"toolset"`
}

type OverviewResult struct {
	Topics    []TopicItem                    `json:"topics"`
	Galgames  []galgameDto.GalgameCard       `json:"galgames"`
	Entities  []galgameDto.EntitySearchGroup `json:"entities"`
	Resources []galgameDto.ResourceCard      `json:"resources"`
	Users     []UserItem                     `json:"users"`
	Replies   []ReplyItem                    `json:"replies"`
	Comments  []CommentItem                  `json:"comments"`
	Toolsets  []toolsetDto.ToolsetCard       `json:"toolsets"`
	Totals    OverviewTotals                 `json:"totals"`
}

type EntitySearchResult struct {
	Groups []galgameDto.EntitySearchGroup `json:"groups"`
	Total  int64                          `json:"total"`
}

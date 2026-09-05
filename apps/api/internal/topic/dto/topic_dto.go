package dto

import (
	"time"

	"kun-galgame-api/pkg/imageclient"
)

type KunUser struct {
	ID     int    `json:"id"`
	Name   string `json:"name"`
	Avatar string `json:"avatar"`
}

type KunUserWithMoemoepoint struct {
	ID          int    `json:"id"`
	Name        string `json:"name"`
	Avatar      string `json:"avatar"`
	Moemoepoint int    `json:"moemoepoint"`
}

type TopicUpvoteRecord struct {
	ID          int       `json:"id"`
	User        KunUser   `json:"user"`
	Description string    `json:"description"`
	Created     time.Time `json:"created"`
}

type ReactionHistoryItem struct {
	User     KunUser   `json:"user"`
	Reaction string    `json:"reaction"`
	Created  time.Time `json:"created"`
}

type ListTopicsRequest struct {
	Page      int    `query:"page" validate:"min=1"`
	Limit     int    `query:"limit" validate:"min=1,max=50"`
	SortField string `query:"sort_field"`
	SortOrder string `query:"sort_order" validate:"omitempty,oneof=asc desc"`
	Category  string `query:"category"`
}

type TopicCard struct {
	ID               int                              `json:"id"`
	Title            string                           `json:"title"`
	View             int                              `json:"view"`
	Sections         []string                         `json:"section"`
	CoverImages      []string                         `json:"cover_images"`
	CoverImageMeta   map[string]imageclient.ImageMeta `json:"cover_image_meta,omitempty"`
	User             KunUser                          `json:"user"`
	Status           int                              `json:"status"`
	HasBestAnswer    bool                             `json:"has_best_answer"`
	MiniApps         []string                         `json:"mini_apps"`
	IsNSFW           bool                             `json:"is_nsfw_topic"`
	LikeCount        int                              `json:"like_count"`
	ReplyCount       int                              `json:"reply_count"`
	CommentCount     int                              `json:"comment_count"`
	StatusUpdateTime time.Time                        `json:"status_update_time"`
	Created          time.Time                        `json:"created"`
	UpvoteTime       *time.Time                       `json:"upvote_time"`
}

type TopicListResponse struct {
	Topics []TopicCard `json:"topics"`
	Total  int64       `json:"total"`
}

type ReactionSummary struct {
	Reaction string    `json:"reaction"`
	Count    int       `json:"count"`
	Mine     bool      `json:"mine"`
	Reactors []KunUser `json:"reactors,omitempty"`
}

type MyTopicInteractions struct {
	Favorited []int            `json:"favorited"`
	Reactions map[int][]string `json:"reactions"`
}

type TopicDetail struct {
	ID               int                              `json:"id"`
	Title            string                           `json:"title"`
	Content          string                           `json:"content_markdown"`
	ContentHtml      string                           `json:"content_html"`
	View             int                              `json:"view"`
	Status           int                              `json:"status"`
	HiddenBy         string                           `json:"hidden_by"`
	IsNSFW           bool                             `json:"is_nsfw"`
	Category         string                           `json:"category"`
	Sections         []string                         `json:"section"`
	CoverImages      []string                         `json:"cover_images"`
	CoverImageMeta   map[string]imageclient.ImageMeta `json:"cover_image_meta,omitempty"`
	User             KunUserWithMoemoepoint           `json:"user"`
	LikeCount        int                              `json:"like_count"`
	IsLiked          bool                             `json:"is_liked"`
	DislikeCount     int                              `json:"dislike_count"`
	IsDisliked       bool                             `json:"is_disliked"`
	FavoriteCount    int                              `json:"favorite_count"`
	IsFavorited      bool                             `json:"is_favorited"`
	UpvoteCount      int                              `json:"upvote_count"`
	IsUpvoted        bool                             `json:"is_upvoted"`
	Reactions        []ReactionSummary                `json:"reactions"`
	ReplyCount       int                              `json:"reply_count"`
	MiniApps         []string                         `json:"mini_apps"`
	StatusUpdateTime time.Time                        `json:"status_update_time"`
	UpvoteTime       *time.Time                       `json:"upvote_time"`
	Edited           *time.Time                       `json:"edited"`
	Created          time.Time                        `json:"created"`
	BestAnswer       *TopicBestAnswer                 `json:"best_answer,omitempty"`
}

type TopicBestAnswer struct {
	ID              int       `json:"id"`
	Floor           int       `json:"floor"`
	User            KunUser   `json:"user"`
	ContentMarkdown string    `json:"content_markdown"`
	ContentHtml     string    `json:"content_html"`
	Created         time.Time `json:"created"`
}

type CreateTopicRequest struct {
	Title       string   `json:"title" validate:"required,min=1,max=233"`
	Content     string   `json:"content" validate:"required,min=1,max=100007"`
	Category    string   `json:"category" validate:"required,oneof=galgame technique others"`
	Sections    []string `json:"section" validate:"required,min=1,max=3"`
	IsNSFW      bool     `json:"is_nsfw"`
	CoverImages []string `json:"cover_images" validate:"omitempty,max=9"`
}

type UpdateTopicRequest struct {
	Title       string   `json:"title" validate:"required,min=1,max=233"`
	Content     string   `json:"content" validate:"required,min=1,max=100007"`
	Category    string   `json:"category" validate:"required,oneof=galgame technique others"`
	Sections    []string `json:"section" validate:"required,min=1,max=3"`
	IsNSFW      bool     `json:"is_nsfw"`
	CoverImages []string `json:"cover_images" validate:"omitempty,max=9"`
}

type TopicInteractionRequest struct {
	TopicID int `json:"topic_id" validate:"required,min=1"`
}

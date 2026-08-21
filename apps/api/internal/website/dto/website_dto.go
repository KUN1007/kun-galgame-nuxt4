package dto

import (
	"time"
)

type CreateWebsiteRequest struct {
	Name          string   `json:"name" validate:"required,max=233"`
	URL           string   `json:"url" validate:"required,fqdn,max=500"`
	Description   string   `json:"description" validate:"required,min=10,max=1000"`
	Icon          string   `json:"icon" validate:"omitempty,url,max=500"`
	IconImageHash string   `json:"icon_image_hash" validate:"max=128"`
	CategoryID    int      `json:"category_id" validate:"required,min=1"`
	AgeLimit      string   `json:"age_limit" validate:"required,oneof=all r18"`
	Status        string   `json:"status" validate:"omitempty,oneof=normal unreachable closed"`
	Language      string   `json:"language" validate:"max=10"`
	TagIDs        []int    `json:"tag_ids"`
	Domain        []string `json:"domain" validate:"max=10,dive,max=100"`
	CreateTime    string   `json:"create_time" validate:"max=20"`
}

type UpdateWebsiteRequest struct {
	WebsiteID     int      `json:"website_id" validate:"required,min=1"`
	Name          string   `json:"name" validate:"required,max=233"`
	URL           string   `json:"url" validate:"required,fqdn,max=500"`
	Description   string   `json:"description" validate:"required,min=10,max=1000"`
	Icon          string   `json:"icon" validate:"omitempty,url,max=500"`
	IconImageHash string   `json:"icon_image_hash" validate:"max=128"`
	CategoryID    int      `json:"category_id" validate:"required,min=1"`
	AgeLimit      string   `json:"age_limit" validate:"required,oneof=all r18"`
	Status        string   `json:"status" validate:"omitempty,oneof=normal unreachable closed"`
	Language      string   `json:"language" validate:"max=10"`
	TagIDs        []int    `json:"tag_ids"`
	Domain        []string `json:"domain" validate:"max=10,dive,max=100"`
	CreateTime    string   `json:"create_time" validate:"max=20"`
}

type DeleteWebsiteRequest struct {
	WebsiteID int `query:"website_id" validate:"required,min=1"`
}

type ToggleInteractionRequest struct {
	WebsiteID int `json:"website_id" validate:"required,min=1"`
}

type WebsiteCard struct {
	ID            int    `json:"id"`
	Name          string `json:"name"`
	Description   string `json:"description"`
	Domain        string `json:"domain"`
	AgeLimit      string `json:"age_limit"`
	Status        string `json:"status"`
	Level         int    `json:"level"`
	Icon          string `json:"icon"`
	IconImageHash string `json:"icon_image_hash"`
	IconURL       string `json:"icon_url"`
	Price         int    `json:"price"`
	Category      string `json:"category"`
}

type WebsiteCategoryBrief struct {
	ID          int    `json:"id"`
	Name        string `json:"name"`
	Label       string `json:"label"`
	Description string `json:"description"`
}

type WebsiteTagBrief struct {
	ID          int    `json:"id"`
	Name        string `json:"name"`
	Description string `json:"description"`
	Label       string `json:"label"`
	Level       int    `json:"level"`
	GroupID     *int   `json:"group_id"`
}

type UserBriefCompact struct {
	ID     int    `json:"id"`
	Name   string `json:"name"`
	Avatar string `json:"avatar"`
}

type WebsiteDetailComment struct {
	ID      int              `json:"id"`
	Content string           `json:"content"`
	User    UserBriefCompact `json:"user"`
	Created string           `json:"created"`
	Updated string           `json:"updated"`
}

type WebsiteDetailResponse struct {
	ID            int                    `json:"id"`
	Name          string                 `json:"name"`
	URL           string                 `json:"url"`
	Description   string                 `json:"description"`
	Icon          string                 `json:"icon"`
	IconImageHash string                 `json:"icon_image_hash"`
	IconURL       string                 `json:"icon_url"`
	View          int                    `json:"view"`
	Language      string                 `json:"language"`
	AgeLimit      string                 `json:"age_limit"`
	Status        string                 `json:"status"`
	Category      WebsiteCategoryBrief   `json:"category"`
	Tags          []WebsiteTagBrief      `json:"tags"`
	LikeCount     int                    `json:"like_count"`
	IsLiked       bool                   `json:"is_liked"`
	FavoriteCount int                    `json:"favorite_count"`
	IsFavorited   bool                   `json:"is_favorited"`
	Domain        any                    `json:"domain"`
	CreateTime    string                 `json:"create_time"`
	Comment       []WebsiteDetailComment `json:"comment"`
	Created       time.Time              `json:"created"`
	Updated       time.Time              `json:"updated"`
}

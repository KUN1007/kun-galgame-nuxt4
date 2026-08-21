package dto

import "time"

type CreateWebsiteTagRequest struct {
	Name        string `json:"name" validate:"required,min=1,max=30"`
	Label       string `json:"label" validate:"required,min=1,max=30"`
	Description string `json:"description" validate:"max=300"`
	Level       int    `json:"level" validate:"min=-100,max=20"`
	GroupID     *int   `json:"group_id" validate:"omitempty,min=1"`
}

type UpdateWebsiteTagRequest struct {
	TagID       int    `json:"tag_id" validate:"required,min=1"`
	Name        string `json:"name" validate:"required,min=1,max=30"`
	Label       string `json:"label" validate:"required,min=1,max=30"`
	Description string `json:"description" validate:"max=300"`
	Level       int    `json:"level" validate:"min=-100,max=20"`
	GroupID     *int   `json:"group_id" validate:"omitempty,min=1"`
}

type DeleteWebsiteTagRequest struct {
	TagID int `query:"tag_id" validate:"required,min=1"`
}

type WebsiteTagDetailResponse struct {
	ID           int           `json:"id"`
	Name         string        `json:"name"`
	Label        string        `json:"label"`
	Level        int           `json:"level"`
	Description  string        `json:"description"`
	GroupID      *int          `json:"group_id"`
	WebsiteCount int           `json:"website_count"`
	Websites     []WebsiteCard `json:"websites"`
	Created      time.Time     `json:"created"`
	Updated      time.Time     `json:"updated"`
}

type CreateWebsiteTagGroupRequest struct {
	Name        string `json:"name" validate:"required,min=1,max=30"`
	Label       string `json:"label" validate:"required,min=1,max=30"`
	Description string `json:"description" validate:"max=300"`
	SortOrder   int    `json:"sort_order" validate:"min=0,max=9999"`
	MultiSelect bool   `json:"multi_select"`
}

type UpdateWebsiteTagGroupRequest struct {
	GroupID     int    `json:"group_id" validate:"required,min=1"`
	Name        string `json:"name" validate:"required,min=1,max=30"`
	Label       string `json:"label" validate:"required,min=1,max=30"`
	Description string `json:"description" validate:"max=300"`
	SortOrder   int    `json:"sort_order" validate:"min=0,max=9999"`
	MultiSelect bool   `json:"multi_select"`
}

type DeleteWebsiteTagGroupRequest struct {
	GroupID int `query:"group_id" validate:"required,min=1"`
}

type WebsiteTagGroupBrief struct {
	ID          int    `json:"id"`
	Name        string `json:"name"`
	Label       string `json:"label"`
	Description string `json:"description"`
	SortOrder   int    `json:"sort_order"`
	MultiSelect bool   `json:"multi_select"`
	TagCount    int    `json:"tag_count"`
}

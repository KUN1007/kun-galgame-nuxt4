package dto

import "time"

type UpdateWebsiteCategoryRequest struct {
	CategoryID  int    `json:"category_id" validate:"required,min=1"`
	Name        string `json:"name" validate:"required,min=1,max=30"`
	Label       string `json:"label" validate:"required,min=1,max=30"`
	Description string `json:"description" validate:"max=300"`
	SortOrder   int    `json:"sort_order" validate:"min=0,max=9999"`
}

type WebsiteCategoryDetailResponse struct {
	ID           int           `json:"id"`
	Name         string        `json:"name"`
	Label        string        `json:"label"`
	Description  string        `json:"description"`
	SortOrder    int           `json:"sort_order"`
	WebsiteCount int           `json:"website_count"`
	Websites     []WebsiteCard `json:"websites"`
	Created      time.Time     `json:"created"`
	Updated      time.Time     `json:"updated"`
}

type CreateWebsiteCategoryRequest struct {
	Name        string `json:"name" validate:"required,min=1,max=30"`
	Label       string `json:"label" validate:"required,min=1,max=30"`
	Description string `json:"description" validate:"max=300"`
	SortOrder   int    `json:"sort_order" validate:"min=0,max=9999"`
}

type DeleteWebsiteCategoryRequest struct {
	CategoryID int `query:"category_id" validate:"required,min=1"`
}

type WebsiteCategoryListItem struct {
	ID           int    `json:"id"`
	Name         string `json:"name"`
	Label        string `json:"label"`
	Description  string `json:"description"`
	SortOrder    int    `json:"sort_order"`
	WebsiteCount int    `json:"website_count"`
}

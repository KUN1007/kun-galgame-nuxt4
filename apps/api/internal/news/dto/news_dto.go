package dto

import "time"

type UserBrief struct {
	ID     int    `json:"id"`
	Name   string `json:"name"`
	Avatar string `json:"avatar"`
}

// NewsSource is sent once per page and named from each item by SourceKey. The
// upstream inlines the whole block on every item instead, because a consumer
// that renders one item in isolation must still be able to show the partner's
// attribution and link back — the two conditions 月幕 and Galgame 批评 attached
// to republication. Collapsing to a map keeps that true here: an item is not
// renderable without a lookup that necessarily yields both.
type NewsSource struct {
	Key         string     `json:"key"`
	Name        string     `json:"name"`
	HomepageURL string     `json:"homepage_url"`
	ColumnURL   string     `json:"column_url"`
	Attribution string     `json:"attribution"`
	Publisher   *UserBrief `json:"publisher"`
}

// NewsItem carries no article body on purpose: the partners authorised an
// index, so preview plus banner is the whole of it and SourceURL is the only
// route to the full text.
type NewsItem struct {
	ID          int64     `json:"id"`
	SourceKey   string    `json:"source_key"`
	Lane        string    `json:"lane"`
	Title       string    `json:"title"`
	Preview     string    `json:"preview"`
	SourceURL   string    `json:"source_url"`
	BannerURL   string    `json:"banner_url"`
	PublishedAt time.Time `json:"published_at"`
}

type NewsFeed struct {
	Items      []NewsItem            `json:"items"`
	Sources    map[string]NewsSource `json:"sources"`
	Count      int64                 `json:"count"`
	NextCursor string                `json:"next_cursor"`
}

type NewsArchiveMonth struct {
	Month int   `json:"month"`
	Count int64 `json:"count"`
}

type NewsArchiveYear struct {
	Year  int   `json:"year"`
	Count int64 `json:"count"`
}

// NewsArchive carries Months for the one year the caller asked about, not for
// every year: each month is a separate upstream count, so building all of them
// eagerly would cost a request per month of the whole corpus.
type NewsArchive struct {
	Years  []NewsArchiveYear  `json:"years"`
	Months []NewsArchiveMonth `json:"months"`
}

type NewsDay struct {
	Day   int `json:"day"`
	Count int `json:"count"`
}

// NewsMonth is one calendar month, page by page. Total is the whole month and
// Count is what is left after the day filter, so the header can keep naming the
// month while the paginator divides the narrowed list.
type NewsMonth struct {
	Items   []NewsItem            `json:"items"`
	Sources map[string]NewsSource `json:"sources"`
	Days    []NewsDay             `json:"days"`
	Total   int                   `json:"total"`
	Count   int                   `json:"count"`
	Page    int                   `json:"page"`
	Limit   int                   `json:"limit"`
}

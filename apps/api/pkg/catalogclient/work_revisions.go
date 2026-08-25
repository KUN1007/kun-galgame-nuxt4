package catalogclient

import (
	"context"
	"net/url"
	"strconv"
	"time"
)

type WorkRevisionFeedItem struct {
	ID            int64     `json:"id"`
	ActorUID      int64     `json:"actor_uid"`
	AmenderUID    *int64    `json:"amender_uid"`
	Site          string    `json:"site"`
	ProductWorkID *int64    `json:"product_work_id"`
	CreatedAt     time.Time `json:"created_at"`
}

type WorkRevisionFeedPage struct {
	Items []WorkRevisionFeedItem `json:"items"`
}

func (c *Client) WorkRevisionsAfter(
	ctx context.Context,
	after int64,
	limit int,
	site string,
) (*WorkRevisionFeedPage, error) {
	q := url.Values{
		"sort":   {"recorded_asc"},
		"object": {"work"},
		"limit":  {strconv.Itoa(clampV2Limit(limit))},
	}
	if site != "" {
		q.Set("site", site)
	}
	if cur := encodeWatermark(after); cur != "" {
		q.Set("cursor", cur)
	}
	var page v2List[v2Revision]
	if err := c.appV2JSON(ctx, "/v2/catalog/revisions", q, &page); err != nil {
		return nil, err
	}
	rows := page.rows()
	out := &WorkRevisionFeedPage{Items: make([]WorkRevisionFeedItem, 0, len(rows))}
	for _, it := range rows {
		item := WorkRevisionFeedItem{
			ID: parseFlexID(it.ID), ActorUID: parseFlexID(it.ActorUID),
			Site: it.Site, CreatedAt: it.time(),
		}
		if n := parseFlexID(it.AmenderUID); n != 0 {
			item.AmenderUID = &n
		}
		if n := parseFlexID(it.SiteWorkID); n != 0 {
			item.ProductWorkID = &n
		}
		out.Items = append(out.Items, item)
	}
	return out, nil
}

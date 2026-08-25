package catalogclient

import (
	"context"
	"net/url"
	"strconv"
	"time"
)

type EditRevisionFeedItem struct {
	ID            int64     `json:"id"`
	EntityFamily  string    `json:"entity_family"`
	EntityType    string    `json:"entity_type"`
	EntityID      int64     `json:"entity_id"`
	Seq           int       `json:"seq"`
	Action        int16     `json:"action"`
	ChangedFields []string  `json:"changed_fields"`
	ActorUID      int64     `json:"actor_uid"`
	AmenderUID    *int64    `json:"amender_uid"`
	ProposalID    *int64    `json:"proposal_id"`
	Site          string    `json:"site"`
	CreatedAt     time.Time `json:"created_at"`
}

const (
	EditActionCreated  int16 = 0
	EditActionMerged   int16 = 1
	EditActionDirect   int16 = 2
	EditActionReverted int16 = 3
)

type EditRevisionFeedPage struct {
	Items     []EditRevisionFeedItem `json:"items"`
	NextSince int64                  `json:"next_since"`
}

func (c *Client) EditRevisionsSince(
	ctx context.Context,
	since int64,
	limit int,
	entityType string,
) (*EditRevisionFeedPage, error) {
	q := url.Values{
		"sort":  {"recorded_asc"},
		"limit": {strconv.Itoa(clampV2Limit(limit))},
	}
	if object := objectFamily(entityType); object != "" {
		q.Set("object", object)
	}
	if cur := encodeWatermark(since); cur != "" {
		q.Set("cursor", cur)
	}
	var page v2List[v2Revision]
	if err := c.appV2JSON(ctx, "/v2/catalog/revisions", q, &page); err != nil {
		return nil, err
	}
	rows := page.rows()
	out := &EditRevisionFeedPage{Items: make([]EditRevisionFeedItem, 0, len(rows))}
	for _, it := range rows {
		item := EditRevisionFeedItem{
			ID: parseFlexID(it.ID), EntityType: entityTypeFromObject(it.TargetObject),
			EntityID: parseFlexID(it.EntityID), Seq: it.Seq, Action: parseRevisionAction(it.Action),
			ChangedFields: it.ChangedFields, ActorUID: parseFlexID(it.ActorUID),
			Site: it.Site, CreatedAt: it.time(),
		}
		if n := parseFlexID(it.AmenderUID); n != 0 {
			item.AmenderUID = &n
		}
		if n := parseFlexID(it.ProposalID); n != 0 {
			item.ProposalID = &n
		}
		out.Items = append(out.Items, item)
		if item.ID > out.NextSince {
			out.NextSince = item.ID
		}
	}
	if out.NextSince == 0 {
		out.NextSince = since
	}
	return out, nil
}

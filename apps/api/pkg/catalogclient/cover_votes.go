package catalogclient

import (
	"context"
	"encoding/json"
	"net/url"
	"strconv"
)

type CoverTally struct {
	ID        int64  `json:"id"`
	ImageHash string `json:"image_hash"`
	VoteCount int    `json:"vote_count"`
	Voted     bool   `json:"voted"`
}

func (c *Client) WorkCoverVotes(ctx context.Context, workID int64) ([]CoverTally, error) {
	// Without the population gate catalog answers 404 "work not found" for every
	// r18 work, and the tallies silently vanished on ~94% of detail pages while
	// logging 11.5k WARNs an hour. The age gate is never a population cut.
	q := url.Values{"include": {"covers"}, "nsfw": {"true"}}
	var rec struct {
		Covers []struct {
			ID        json.RawMessage `json:"id"`
			Hash      string          `json:"hash"`
			ImageHash string          `json:"image_hash"`
			VoteCount int             `json:"vote_count"`
			Voted     bool            `json:"voted"`
		} `json:"covers"`
	}
	if err := c.appV2JSON(ctx, "/v2/catalog/works/"+strconv.FormatInt(workID, 10), q, &rec); err != nil {
		return nil, err
	}
	out := make([]CoverTally, 0, len(rec.Covers))
	for _, it := range rec.Covers {
		hash := it.Hash
		if hash == "" {
			hash = it.ImageHash
		}
		out = append(out, CoverTally{
			ID: parseFlexID(it.ID), ImageHash: hash, VoteCount: it.VoteCount, Voted: it.Voted,
		})
	}
	return out, nil
}

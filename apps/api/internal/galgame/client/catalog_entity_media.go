package client

import (
	"context"
	"encoding/json"
	"net/url"
	"strconv"
	"strings"

	"kun-galgame-api/pkg/errors"
)

// CatalogEntityMedia is what an entity card needs beyond a name. The search
// face answers names and ids only, so it has to be fetched by id afterwards.
type CatalogEntityMedia struct {
	Image     string
	WorkCount int
}

const catalogEntityMediaCap = 100

// CatalogEntityMediaBatch fetches the picture and the work count for ids the
// search face already returned. entity is a v1 collection name — characters,
// labels or tags — because CatalogGet still speaks v1 and rewrites on the way.
func (c *GalgameClient) CatalogEntityMediaBatch(
	ctx context.Context, entity string, ids []int64,
) (map[int64]CatalogEntityMedia, *errors.AppError) {
	if len(ids) == 0 {
		return map[int64]CatalogEntityMedia{}, nil
	}
	if len(ids) > catalogEntityMediaCap {
		ids = ids[:catalogEntityMediaCap]
	}

	raw := make([]string, len(ids))
	for i, id := range ids {
		raw[i] = strconv.FormatInt(id, 10)
	}
	q := url.Values{
		"ids":   {strings.Join(raw, ",")},
		"limit": {strconv.Itoa(len(ids))},
	}
	// Companies get their include forced by the v2 query rewriter; characters
	// do not, and without this the row comes back with no image at all.
	if entity == "characters" {
		q.Set("include", "image")
	}
	openPopulation(q)

	data, appErr := c.CatalogGet(ctx, "/catalog/"+entity, q)
	if appErr != nil {
		return nil, appErr
	}
	var parsed struct {
		Items []catEntityMediaRow `json:"items"`
	}
	if err := json.Unmarshal(data, &parsed); err != nil {
		return nil, errors.ErrInternal("解析 Catalog 实体配图响应失败")
	}

	out := make(map[int64]CatalogEntityMedia, len(parsed.Items))
	for _, row := range parsed.Items {
		media := CatalogEntityMedia{Image: row.Image, WorkCount: row.WorkCount}
		if media.Image == "" && row.Logo != nil {
			media.Image = row.Logo.URL
		}
		if media.Image == "" && row.LogoHash != "" {
			media.Image = c.ImageURLFromHash(row.LogoHash)
		}
		out[row.ID] = media
	}
	return out, nil
}

// The rewriter flattens a character's image object down to its URL but leaves a
// company's logo an object beside the logo_hash it derives, so both spellings
// have to be read.
type catEntityMediaRow struct {
	ID        int64  `json:"id"`
	Image     string `json:"image"`
	WorkCount int    `json:"work_count"`
	LogoHash  string `json:"logo_hash"`
	Logo      *struct {
		URL string `json:"url"`
	} `json:"logo"`
}

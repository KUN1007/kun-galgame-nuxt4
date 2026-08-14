package service

import (
	"context"
	"net/url"
	"strconv"

	"kun-galgame-api/internal/galgame/client"
	"kun-galgame-api/internal/galgame/dto"
	"kun-galgame-api/pkg/errors"
)

// catalogWorksSort maps the forum's sortField/sortOrder onto a catalog works
// search sort lane. Only release_date has a direction-aware pair; the forum's
// view/rating/time lanes are approximated by popularity/updated (they are
// forum-side counters with no catalog equivalent), and anything else falls
// back to newest-first.
func catalogWorksSort(sortField, sortOrder string) string {
	switch sortField {
	case "release_date":
		if sortOrder == "asc" {
			return "released_asc"
		}
		return "released_desc"
	case "view", "view_1d", "view_7d", "view_30d", "rating":
		return "popularity"
	case "time":
		return "updated"
	default:
		return "released_desc"
	}
}

// fetchCatalogMemberCards lists a taxonomy entity's LIVE members straight from
// the catalog works search and renders them as cards. The legacy path
// (CatalogMemberGIDs → hydrateListCards) only surfaced works the forum itself
// had claimed, so a work claimed by a foreign site — counted by the badge —
// silently vanished from the list. Listing the catalog directly keeps every
// live member, and each card carries IsOnForum so the frontend can render the
// unclaimed ones without a forum detail link.
func fetchCatalogMemberCards(
	ctx context.Context,
	galgameClient *client.GalgameClient,
	enricher *GalgameEnricher,
	filter url.Values,
	rawQuery url.Values,
	isSFW bool,
) ([]dto.GalgameCard, int64, *errors.AppError) {
	q := url.Values{
		"page":        {strconv.Itoa(atoiOr(rawQuery.Get("page"), 1))},
		"limit":       {strconv.Itoa(atoiOr(rawQuery.Get("limit"), 24))},
		"claim_state": {"live"},
		"include":     {CatalogCardInclude},
		"sort":        {catalogWorksSort(rawQuery.Get("sortField"), rawQuery.Get("sortOrder"))},
	}
	for k, v := range filter {
		q[k] = v
	}
	client.ApplyWorksGate(q, isSFW)

	res, appErr := galgameClient.CatalogWorksSearch(ctx, q)
	if appErr != nil {
		return nil, 0, appErr
	}
	return enricher.ToCards(ctx, catalogItemsToNextMoe(res.Items)), res.Total, nil
}

// paginateEntityCards applies the page/limit query params to an in-memory card
// slice, for lanes that already fetched every member up front (the list-endpoint
// rollup path has no server-side sort, so pagination happens here).
func paginateEntityCards(cards []dto.GalgameCard, rawQuery url.Values) []dto.GalgameCard {
	page := atoiOr(rawQuery.Get("page"), 1)
	limit := atoiOr(rawQuery.Get("limit"), 24)
	if page < 1 {
		page = 1
	}
	if limit < 1 {
		limit = 24
	}
	start := (page - 1) * limit
	if start >= len(cards) {
		return []dto.GalgameCard{}
	}
	end := start + limit
	if end > len(cards) {
		end = len(cards)
	}
	return cards[start:end]
}

package service

import (
	"context"
	"net/url"
	"strconv"

	"kun-galgame-api/internal/galgame/client"
	galgameDto "kun-galgame-api/internal/galgame/dto"
	"kun-galgame-api/internal/search/dto"
	"kun-galgame-api/pkg/errors"
)

// A keyword reaches a download resource two ways: through the note its uploader
// wrote, and through the name of the game it hangs off. Only the first is local
// — the forum stores no game names — so the second one is a catalog search
// whose ids are handed to SQL, capped at the one page catalog will answer.
const resourceGalgameMatchCap = 100

func (s *SearchService) SearchResources(
	ctx context.Context,
	raw string,
	page, limit int,
	isSFW bool,
) (*dto.PaginatedResult[galgameDto.ResourceCard], *errors.AppError) {
	keywords, appErr := tokenize(raw)
	if appErr != nil {
		return nil, appErr
	}
	if s.resource == nil {
		return nil, errors.ErrInternal("Galgame 资源搜索未启用")
	}

	res := s.resource.Search(
		ctx, keywords, s.matchedGalgameIDs(ctx, raw, isSFW), page, limit, 0, isSFW)
	return &dto.PaginatedResult[galgameDto.ResourceCard]{
		Items: res.Resources,
		Total: res.Total,
	}, nil
}

// A failed catalog search costs the game-name half of the match, not the lane:
// the note half is local and answers on its own.
func (s *SearchService) matchedGalgameIDs(ctx context.Context, raw string, isSFW bool) []int {
	if s.galgameClient == nil {
		return nil
	}
	q := url.Values{
		"q":     {raw},
		"page":  {"1"},
		"limit": {strconv.Itoa(resourceGalgameMatchCap)},
		"sort":  {"relevance"},
	}
	client.ApplyWorksGate(q, isSFW)

	res, appErr := s.galgameClient.CatalogWorksSearch(ctx, q)
	if appErr != nil {
		return nil
	}
	ids := make([]int, 0, len(res.Items))
	for i := range res.Items {
		if gid := client.CatalogItemGID(&res.Items[i]); gid > 0 {
			ids = append(ids, gid)
		}
	}
	return ids
}

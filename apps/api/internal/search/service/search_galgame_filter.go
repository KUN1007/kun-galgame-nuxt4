package service

import (
	"context"
	"net/url"
	"strconv"
	"strings"

	galgameDto "kun-galgame-api/internal/galgame/dto"
	"kun-galgame-api/internal/search/dto"
	"kun-galgame-api/pkg/errors"
	"kun-galgame-api/pkg/utils"
)

// Catalog answers 400 past ten tag ids.
const maxSearchTagIDs = 10

func applyGalgameFilter(q url.Values, filter dto.GalgameFilter) *errors.AppError {
	if filter.CompanyID != "" && filter.CompanyID != "0" {
		q.Set("company_id", filter.CompanyID)
	}
	if ids := parseIDList(filter.TagIDs, maxSearchTagIDs); len(ids) > 0 {
		q.Set("tag_id", strings.Join(ids, ","))
	}
	from, err := utils.ParseReleaseLowerBound(filter.ReleasedFrom)
	if err != nil {
		return errors.ErrBadRequest(err.Error())
	}
	to, err := utils.ParseReleaseUpperBound(filter.ReleasedTo)
	if err != nil {
		return errors.ErrBadRequest(err.Error())
	}
	if from != "" {
		q.Set("released_after", from)
	}
	if to != "" {
		q.Set("released_before", to)
	}
	return nil
}

func parseIDList(raw string, cap int) []string {
	out := make([]string, 0, cap)
	for _, part := range strings.Split(raw, ",") {
		part = strings.TrimSpace(part)
		if n, err := strconv.Atoi(part); err != nil || n <= 0 {
			continue
		}
		out = append(out, part)
		if len(out) == cap {
			break
		}
	}
	return out
}

func (s *SearchService) ResolveEntities(
	ctx context.Context,
	family, rawIDs string,
	isSFW bool,
) ([]galgameDto.EntitySearchItem, *errors.AppError) {
	if s.entityService == nil {
		return nil, errors.ErrInternal("Galgame 资料库搜索未启用")
	}
	raw := parseIDList(rawIDs, maxSearchTagIDs)
	ids := make([]int, 0, len(raw))
	for _, id := range raw {
		n, _ := strconv.Atoi(id)
		ids = append(ids, n)
	}
	return s.entityService.Resolve(ctx, family, ids, isSFW)
}

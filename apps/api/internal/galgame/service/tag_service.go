package service

import (
	"context"
	"net/url"
	"strconv"
	"strings"

	"kun-galgame-api/internal/galgame/client"
	"kun-galgame-api/internal/galgame/dto"
	"kun-galgame-api/pkg/errors"
)

type TagService struct {
	galgameClient *client.GalgameClient
	enricher      *GalgameEnricher
	galgameSvc    *GalgameService
	index         staleCache[indexedTag]
}

func NewTagService(galgameClient *client.GalgameClient, enricher *GalgameEnricher, galgameSvc *GalgameService) *TagService {
	return &TagService{galgameClient: galgameClient, enricher: enricher, galgameSvc: galgameSvc}
}

type TagMultiPage struct {
	Galgames []dto.GalgameCard `json:"galgames"`
	Total    int64             `json:"total"`
}

const taxonomyMemberPageCap = 200

const maxTagFilterIDs = 10

const CatalogCardInclude = "titles,covers,companies"

func tagCategory(kind string, sexual bool) string {
	if sexual {
		return "sexual"
	}
	return kind
}

func (s *TagService) Search(
	ctx context.Context,
	rawQuery url.Values,
	isSFW bool,
) ([]dto.TaxonomySearchItem, *errors.AppError) {
	hits, appErr := s.galgameClient.CatalogEntitySearch(ctx, "tags",
		rawQuery.Get("q"), atoiOr(rawQuery.Get("limit"), 20))
	if appErr != nil {
		return nil, appErr
	}

	kept := make([]client.CatalogEntityHit, 0, len(hits))
	ids := make([]int, 0, len(hits))
	for _, h := range hits {
		if h.Tier == client.TagTierHidden {
			continue
		}
		kept = append(kept, h)
		ids = append(ids, int(h.ID))
	}
	sexual := map[int]bool{}
	if isSFW && len(ids) > 0 {
		indexed, missing := s.sexualByID(ctx, ids)
		sexual = indexed
		for id, isSexual := range s.galgameClient.CatalogSexualTagIDs(ctx, missing) {
			sexual[id] = isSexual
		}
	}

	items := make([]dto.TaxonomySearchItem, 0, len(kept))
	for _, h := range kept {
		if sexual[int(h.ID)] {
			continue
		}
		items = append(items, dto.TaxonomySearchItem{ID: int(h.ID), Name: h.VocabularyName()})
	}
	return items, nil
}

func (s *TagService) GetByMultiTag(
	ctx context.Context,
	rawQuery url.Values,
	isSFW bool,
) (*TagMultiPage, *errors.AppError) {
	ids := rawQuery.Get("tag_ids")

	q := url.Values{
		"page":    {strconv.Itoa(atoiOr(rawQuery.Get("page"), 1))},
		"limit":   {strconv.Itoa(atoiOr(rawQuery.Get("limit"), 24))},
		"include": {CatalogCardInclude},
		"sort":    {"released_desc"},
	}
	if selected := splitCSV(ids); len(selected) > 0 {
		if len(selected) > maxTagFilterIDs {
			selected = selected[:maxTagFilterIDs]
		}
		q.Set("tag_id", strings.Join(selected, ","))
	}
	client.ApplyWorksGate(q, isSFW)

	res, appErr := s.galgameClient.CatalogWorksSearch(ctx, q)
	if appErr != nil {
		return nil, appErr
	}
	return &TagMultiPage{
		Galgames: s.enricher.ToCards(ctx, catalogItemsToNextMoe(ctx, res.Items)),
		Total:    res.Total,
	}, nil
}

func (s *TagService) GetList(
	ctx context.Context,
	rawQuery url.Values,
	isSFW bool,
) (*dto.TagListPage, *errors.AppError) {
	index, appErr := s.indexRows(ctx)
	if appErr != nil {
		return nil, appErr
	}

	tags := make([]dto.TagListItem, 0, len(index))
	for _, t := range index {
		if isSFW && t.sexual {
			continue
		}
		tags = append(tags, t.item)
	}
	total := int64(len(tags))

	page, limit := atoiOr(rawQuery.Get("page"), 1), atoiOr(rawQuery.Get("limit"), 100)
	if page < 1 {
		page = 1
	}
	if limit < 1 {
		limit = 100
	}
	if start := (page - 1) * limit; start >= len(tags) {
		tags = nil
	} else {
		tags = tags[start:min(start+limit, len(tags))]
	}
	return &dto.TagListPage{Tags: tags, Total: total}, nil
}

func (s *TagService) GetDetail(
	ctx context.Context,
	id string,
	rawQuery url.Values,
	isSFW bool,
) (*dto.TagDetail, *errors.AppError) {
	t, found, appErr := s.galgameClient.CatalogTag(ctx, id)
	if appErr != nil {
		return nil, appErr
	}
	if !found {
		return nil, errors.ErrNotFound("未找到该标签")
	}

	memberIDs, appErr := s.galgameClient.CatalogMemberGIDs(ctx,
		url.Values{"tag_id": {id}}, isSFW, taxonomyMemberPageCap)
	if appErr != nil {
		return nil, appErr
	}
	page, appErr := s.galgameSvc.hydrateListCards(ctx, buildEntityFilter(rawQuery, memberIDs), isSFW)
	if appErr != nil {
		return nil, appErr
	}

	return &dto.TagDetail{
		ID:           int(t.ID),
		Name:         t.Label(),
		Category:     tagCategory(t.Kind, t.Sexual),
		Hidden:       t.Tier == client.TagTierHidden,
		Description:  preferredIntro(t.Intros).Intro,
		Alias:        []string{},
		Galgame:      listCardsToEntityCards(page.Galgames),
		GalgameCount: page.Total,
	}, nil
}

func catalogItemsToNextMoe(ctx context.Context, items []client.CatalogWorkListItem) []dto.NextMoeGalgameItem {
	out := make([]dto.NextMoeGalgameItem, 0, len(items))
	for i := range items {
		if !client.CatalogItemRenderable(&items[i]) {
			continue
		}
		out = append(out, client.CatalogItemToNextMoeItem(ctx, &items[i]))
	}
	return out
}

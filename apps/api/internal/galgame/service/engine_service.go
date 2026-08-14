package service

import (
	"context"
	"net/url"

	"kun-galgame-api/internal/galgame/client"
	"kun-galgame-api/internal/galgame/dto"
	"kun-galgame-api/pkg/errors"
)

type EngineService struct {
	galgameClient *client.GalgameClient
	galgameSvc    *GalgameService
	enricher      *GalgameEnricher
}

func NewEngineService(galgameClient *client.GalgameClient, enricher *GalgameEnricher, galgameSvc *GalgameService) *EngineService {
	return &EngineService{galgameClient: galgameClient, enricher: enricher, galgameSvc: galgameSvc}
}

const engineIndexPageCap = 20

func (s *EngineService) GetList(ctx context.Context) ([]dto.EngineListItem, *errors.AppError) {
	items := []dto.EngineListItem{}
	cursor := ""
	for page := 0; page < engineIndexPageCap; page++ {
		q := client.OpenPopulation(url.Values{"limit": {"100"}})
		if cursor != "" {
			q.Set("cursor", cursor)
		}
		res, appErr := s.galgameClient.CatalogTaxonomyList(ctx, "engines", q)
		if appErr != nil {
			return nil, appErr
		}
		for _, e := range res.Items {
			items = append(items, dto.EngineListItem{
				ID:           int(e.ID),
				Name:         e.Label(),
				Description:  e.Description,
				Alias:        emptyStrSliceIfNil(e.Aliases),
				GalgameCount: e.WorkCount,
			})
		}
		if res.NextCursor == nil || *res.NextCursor == "" {
			break
		}
		cursor = *res.NextCursor
	}
	return items, nil
}

func (s *EngineService) GetDetail(
	ctx context.Context,
	id string,
	rawQuery url.Values,
	isSFW bool,
) (*dto.EngineDetail, *errors.AppError) {
	e, found, appErr := s.galgameClient.CatalogEngine(ctx, id)
	if appErr != nil {
		return nil, appErr
	}
	if !found {
		return nil, errors.ErrNotFound("未找到该引擎")
	}

	cards, total, appErr := fetchCatalogMemberCards(ctx,
		s.galgameClient, s.enricher,
		url.Values{"engine_id": {id}}, rawQuery, isSFW)
	if appErr != nil {
		return nil, appErr
	}

	return &dto.EngineDetail{
		ID:           int(e.ID),
		Name:         e.Name,
		Description:  e.Description,
		Alias:        emptyStrSliceIfNil(e.Aliases),
		Galgame:      cards,
		GalgameCount: total,
	}, nil
}

func emptyStrSliceIfNil(s []string) []string {
	if s == nil {
		return []string{}
	}
	return s
}

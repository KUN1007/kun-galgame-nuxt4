package service

import (
	"context"
	"net/url"

	"kun-galgame-api/internal/galgame/client"
	"kun-galgame-api/internal/galgame/dto"
	"kun-galgame-api/pkg/errors"
)

type OfficialService struct {
	galgameClient *client.GalgameClient
	enricher      *GalgameEnricher
	index         *officialIndexCache
}

func NewOfficialService(galgameClient *client.GalgameClient, enricher *GalgameEnricher) *OfficialService {
	return &OfficialService{
		galgameClient: galgameClient,
		enricher:      enricher,
		index:         newOfficialIndexCache(),
	}
}

func (s *OfficialService) GetList(
	ctx context.Context,
	rawQuery url.Values,
) (*dto.OfficialListPage, *errors.AppError) {
	items, total, appErr := s.page(ctx, rawQuery.Get("kind"),
		atoiOr(rawQuery.Get("page"), 1), atoiOr(rawQuery.Get("limit"), 50))
	if appErr != nil {
		return nil, appErr
	}
	return &dto.OfficialListPage{Officials: items, Total: total}, nil
}

func (s *OfficialService) Search(
	ctx context.Context,
	rawQuery url.Values,
) ([]dto.TaxonomySearchItem, *errors.AppError) {
	hits, appErr := s.galgameClient.CatalogEntitySearch(ctx, "labels",
		rawQuery.Get("q"), atoiOr(rawQuery.Get("limit"), 20))
	if appErr != nil {
		return nil, appErr
	}
	items := make([]dto.TaxonomySearchItem, 0, len(hits))
	for _, o := range hits {
		items = append(items, dto.TaxonomySearchItem{
			ID:   int(o.ID),
			Name: o.Name,
			Logo: s.galgameClient.ImageURLFromHash(o.LogoHash),
		})
	}
	return items, nil
}

func (s *OfficialService) GetDetail(
	ctx context.Context,
	id string,
	rawQuery url.Values,
	isSFW bool,
) (*dto.OfficialDetail, *errors.AppError) {
	o, found, movedTo, appErr := s.galgameClient.CatalogLabel(ctx, id)
	if appErr != nil {
		return nil, appErr
	}
	if movedTo != 0 {
		return &dto.OfficialDetail{MovedTo: int(movedTo)}, nil
	}
	if !found {
		return nil, errors.ErrNotFound("未找到该会社")
	}

	items, appErr := s.galgameClient.CatalogLabelMemberItems(ctx, id, isSFW, taxonomyMemberPageCap)
	if appErr != nil {
		return nil, appErr
	}

	cards := s.enricher.ToCards(ctx, catalogItemsToNextMoe(items))
	total := int64(len(cards))
	var imprintCount int64
	for i := range cards {
		if cards[i].ViaOfficial != nil {
			imprintCount++
		}
	}
	cards = paginateEntityCards(cards, rawQuery)

	return &dto.OfficialDetail{
		ID:                  int(o.ID),
		Name:                o.DisplayName,
		Links:               officialLinks(o.Links),
		Link:                client.PrimaryLabelLink(o),
		Logo:                s.galgameClient.ImageURLFromHash(o.LogoHash),
		Category:            o.Kind,
		Lang:                o.Lang,
		Description:         preferredIntro(o.Intros),
		Alias:               emptyStrSliceIfNil(o.Aliases),
		Galgame:             cards,
		GalgameCount:        total,
		OwnGalgameCount:     total - imprintCount,
		ImprintGalgameCount: imprintCount,
	}, nil
}

func (s *OfficialService) GetRelationGraph(ctx context.Context, id string) (*dto.OfficialRelationGraph, *errors.AppError) {
	graph, found, appErr := s.galgameClient.CatalogLabelRelationGraph(ctx, id)
	if appErr != nil {
		return nil, appErr
	}
	if !found {
		return nil, errors.ErrNotFound("未找到该会社")
	}

	out := &dto.OfficialRelationGraph{
		Nodes: make([]dto.OfficialRelationNode, 0, len(graph.Nodes)),
		Edges: make([]dto.OfficialRelationEdge, 0, len(graph.Edges)),
	}
	for _, n := range graph.Nodes {
		out.Nodes = append(out.Nodes, dto.OfficialRelationNode{
			ID:        int(n.ID),
			Name:      n.Name,
			Logo:      s.galgameClient.ImageURLFromHash(n.LogoHash),
			WorkCount: n.WorkCount,
		})
	}
	for _, e := range graph.Edges {
		out.Edges = append(out.Edges, dto.OfficialRelationEdge{
			From:     int(e.From),
			To:       int(e.To),
			Relation: e.Relation,
		})
	}
	return out, nil
}

func officialLinks(links []client.CatalogLabelLink) []dto.OfficialLink {
	out := make([]dto.OfficialLink, 0, len(links))
	for _, l := range links {
		out = append(out, dto.OfficialLink{
			Source: l.Source,
			Name:   client.LinkDisplayName(l.Source, l.URL),
			URL:    l.URL,
		})
	}
	return out
}

func (s *OfficialService) ResolveLegacyID(ctx context.Context, wikiID int) (int64, bool, *errors.AppError) {
	return s.galgameClient.LookupWikiLabel(ctx, wikiID)
}

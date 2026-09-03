package service

import (
	"context"
	"net/url"

	"kun-galgame-api/internal/galgame/client"
	"kun-galgame-api/internal/galgame/dto"
	"kun-galgame-api/pkg/errors"
)

func searchCatalogEntities(
	ctx context.Context,
	c *client.GalgameClient,
	searchType string,
	rawQuery url.Values,
) ([]dto.TaxonomySearchItem, *errors.AppError) {
	hits, _, appErr := c.CatalogEntitySearch(ctx, searchType, rawQuery.Get("q"), 1,
		atoiOr(rawQuery.Get("limit"), 20))
	if appErr != nil {
		return nil, appErr
	}
	items := make([]dto.TaxonomySearchItem, 0, len(hits))
	for _, hit := range hits {
		items = append(items, dto.TaxonomySearchItem{ID: int(hit.ID), Name: hit.Name(ctx)})
	}
	return items, nil
}

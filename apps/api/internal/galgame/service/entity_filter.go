package service

import (
	"net/url"
	"strconv"

	"kun-galgame-api/internal/galgame/dto"
	"kun-galgame-api/internal/galgame/model"
)

func buildEntityFilter(q url.Values, restrictIDs []int) model.GalgameListFilter {
	sortOrder := q.Get("sortOrder")
	if sortOrder == "" {
		sortOrder = "desc"
	}
	return model.GalgameListFilter{
		Type:           q.Get("type"),
		Language:       q.Get("language"),
		Platform:       q.Get("platform"),
		GameType:       q.Get("gameType"),
		SortField:      q.Get("sortField"),
		SortOrder:      sortOrder,
		Page:           atoiOr(q.Get("page"), 1),
		Limit:          atoiOr(q.Get("limit"), 24),
		ShowNoResource: q.Get("showNoResource") == "true",
		RestrictIDs:    restrictIDs,
	}
}

func listCardsToEntityCards(cards []dto.GalgameListCard) []dto.GalgameCard {
	out := make([]dto.GalgameCard, len(cards))
	for i, c := range cards {
		out[i] = dto.GalgameCard{
			ID:                         c.ID,
			Name:                       c.Name,
			NameOriginal:               c.NameOriginal,
			User:                       c.User,
			ContentLimit:               c.ContentLimit,
			View:                       c.View,
			LikeCount:                  c.LikeCount,
			ResourceUpdateTime:         c.ResourceUpdateTime,
			Platform:                   c.Platform,
			Language:                   c.Language,
			ReleaseDate:                c.ReleaseDate,
			ReleaseDateTBA:             c.ReleaseDateTBA,
			EffectiveBannerHash:        c.EffectiveBannerHash,
			EffectiveBannerURL:         c.EffectiveBannerURL,
			EffectiveBannerWidth:       c.EffectiveBannerWidth,
			EffectiveBannerHeight:      c.EffectiveBannerHeight,
			EffectiveBannerThumbhash:   c.EffectiveBannerThumbhash,
			EffectivePortraitHash:      c.EffectivePortraitHash,
			EffectivePortraitURL:       c.EffectivePortraitURL,
			EffectivePortraitWidth:     c.EffectivePortraitWidth,
			EffectivePortraitHeight:    c.EffectivePortraitHeight,
			EffectivePortraitThumbhash: c.EffectivePortraitThumbhash,
			IsOnForum:                  c.IsOnForum,
		}
	}
	return out
}

func atoiOr(s string, fallback int) int {
	if s == "" {
		return fallback
	}
	n, err := strconv.Atoi(s)
	if err != nil {
		return fallback
	}
	return n
}

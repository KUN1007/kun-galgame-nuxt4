package service

import (
	"context"

	"kun-galgame-api/internal/galgame/dto"
	"kun-galgame-api/internal/galgame/repository"
	"kun-galgame-api/pkg/userclient"
)

type GalgameEnricher struct {
	galgameRepo *repository.GalgameRepository
	metaRepo    *repository.GalgameResourceMetaRepository
	userClient  *userclient.Client
}

func NewGalgameEnricher(
	galgameRepo *repository.GalgameRepository,
	metaRepo *repository.GalgameResourceMetaRepository,
	userClient *userclient.Client,
) *GalgameEnricher {
	return &GalgameEnricher{galgameRepo: galgameRepo, metaRepo: metaRepo, userClient: userClient}
}

func (e *GalgameEnricher) ToCards(ctx context.Context, items []dto.NextMoeGalgameItem) []dto.GalgameCard {
	if len(items) == 0 {
		return []dto.GalgameCard{}
	}

	galgameIDs := make([]int, len(items))
	for i, g := range items {
		galgameIDs[i] = g.ID
	}

	localMap := e.galgameRepo.FindLocalBatch(galgameIDs)
	userMap := e.userClient.Hydrate(ctx, frozenCreatorIDs(galgameIDs, localMap))
	platformMap, languageMap := groupResourceMeta(e.metaRepo.FindResourceMetaBatch(galgameIDs))

	cards := make([]dto.GalgameCard, len(items))
	for i, g := range items {
		_, onForum := localMap[g.ID]
		cards[i] = dto.GalgameCard{
			ID: g.ID,
			Name: dto.KunLanguage{
				EnUs: g.NameEnUs, JaJp: g.NameJaJp,
				ZhCn: g.NameZhCn, ZhTw: g.NameZhTw,
			},
			Banner:                   g.Banner,
			User:                     frozenCreatorBrief(localMap[g.ID], userMap),
			ContentLimit:             g.ContentLimit,
			View:                     localMap[g.ID].View,
			LikeCount:                localMap[g.ID].LikeCount,
			ResourceUpdateTime:       g.ResourceUpdateTime,
			ReleaseDate:              g.ReleaseDate,
			ReleaseDateTBA:           g.ReleaseDateTBA,
			ReleasePrecision:         g.ReleasePrecision,
			Status:                   g.Status,
			EffectiveBannerHash:      g.EffectiveBannerHash,
			EffectiveBannerURL:       g.EffectiveBannerURL,
			EffectiveBannerWidth:     g.EffectiveBannerWidth,
			EffectiveBannerHeight:    g.EffectiveBannerHeight,
			EffectiveBannerThumbhash: g.EffectiveBannerThumbhash,
			Platform:                 emptyStrSliceIfNil(platformMap[g.ID]),
			Language:                 emptyStrSliceIfNil(languageMap[g.ID]),
			IsOnForum:                onForum,
			ViaOfficial:              g.ViaOfficial,
		}
	}
	return cards
}

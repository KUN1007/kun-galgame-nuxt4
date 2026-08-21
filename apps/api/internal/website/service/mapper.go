package service

import (
	"kun-galgame-api/internal/website/dto"
	"kun-galgame-api/internal/website/repository"
	"kun-galgame-api/pkg/imageclient"
)

func websiteCardsFromRows(
	rows []repository.WebsiteListRow,
	catMap map[int]string,
	levelMap map[int]int,
	cdnBase string,
) []dto.WebsiteCard {
	cards := make([]dto.WebsiteCard, len(rows))
	for i, r := range rows {
		lvl := levelMap[r.ID]
		cards[i] = dto.WebsiteCard{
			ID:            r.ID,
			Name:          r.Name,
			Description:   r.Description,
			Domain:        r.URL,
			AgeLimit:      r.AgeLimit,
			Status:        r.Status,
			Level:         lvl,
			Icon:          r.Icon,
			IconImageHash: r.IconImageHash,
			IconURL:       imageclient.ResolveURL(cdnBase, r.IconImageHash, r.Icon),
			Price:         lvl,
			Category:      catMap[r.CategoryID],
		}
	}
	return cards
}

func websiteCardsFromRowsSingleCategory(
	rows []repository.WebsiteListRow,
	categoryName string,
	levelMap map[int]int,
	cdnBase string,
) []dto.WebsiteCard {
	cards := make([]dto.WebsiteCard, len(rows))
	for i, r := range rows {
		lvl := levelMap[r.ID]
		cards[i] = dto.WebsiteCard{
			ID:            r.ID,
			Name:          r.Name,
			Description:   r.Description,
			Domain:        r.URL,
			AgeLimit:      r.AgeLimit,
			Status:        r.Status,
			Level:         lvl,
			Icon:          r.Icon,
			IconImageHash: r.IconImageHash,
			IconURL:       imageclient.ResolveURL(cdnBase, r.IconImageHash, r.Icon),
			Price:         lvl,
			Category:      categoryName,
		}
	}
	return cards
}

func collectCategoryIDs(rows []repository.WebsiteListRow) []int {
	ids := make([]int, 0, len(rows))
	seen := make(map[int]struct{}, len(rows))
	for _, r := range rows {
		if _, ok := seen[r.CategoryID]; ok {
			continue
		}
		seen[r.CategoryID] = struct{}{}
		ids = append(ids, r.CategoryID)
	}
	return ids
}

func collectWebsiteIDs(rows []repository.WebsiteListRow) []int {
	ids := make([]int, len(rows))
	for i, r := range rows {
		ids[i] = r.ID
	}
	return ids
}

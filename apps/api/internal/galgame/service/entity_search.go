package service

import (
	"context"
	"log/slog"
	"sync"

	"kun-galgame-api/internal/galgame/client"
	"kun-galgame-api/internal/galgame/dto"
	"kun-galgame-api/pkg/errors"
)

const (
	EntityFamilyCharacter = "character"
	EntityFamilyCompany   = "company"
	EntityFamilyStaff     = "staff"
	EntityFamilyTag       = "tag"
)

// catalog answers one family per request and its search face carries no cursor,
// so a family is a whole page: what limit does not fetch is unreachable.
var entityFamilies = []string{
	EntityFamilyCharacter,
	EntityFamilyCompany,
	EntityFamilyStaff,
	EntityFamilyTag,
}

var entityCatalogType = map[string]string{
	EntityFamilyCharacter: "characters",
	EntityFamilyCompany:   "labels",
	EntityFamilyStaff:     "names",
	EntityFamilyTag:       "tags",
}

func IsEntityFamily(family string) bool {
	_, ok := entityCatalogType[family]
	return ok
}

type EntitySearchService struct {
	galgameClient *client.GalgameClient
	tagService    *TagService
}

func NewEntitySearchService(galgameClient *client.GalgameClient, tagService *TagService) *EntitySearchService {
	return &EntitySearchService{galgameClient: galgameClient, tagService: tagService}
}

// Search runs one catalog request per family. A family that fails is dropped so
// a single slow or broken family does not blank the page, but a failure of all
// of them is an error rather than an empty result set.
func (s *EntitySearchService) Search(
	ctx context.Context,
	keywords string,
	family string,
	limit int,
	isSFW bool,
) ([]dto.EntitySearchGroup, *errors.AppError) {
	families := entityFamilies
	if family != "" {
		if !IsEntityFamily(family) {
			return nil, errors.ErrBadRequest("未知的实体类型")
		}
		families = []string{family}
	}

	groups := make([]dto.EntitySearchGroup, len(families))
	var (
		wg     sync.WaitGroup
		mu     sync.Mutex
		failed int
	)
	for i, f := range families {
		wg.Add(1)
		go func(i int, f string) {
			defer func() {
				if r := recover(); r != nil {
					slog.Error("entity search family panicked", "family", f, "panic", r)
				}
			}()
			defer wg.Done()

			items, total, appErr := s.searchFamily(ctx, f, keywords, limit, isSFW)
			if appErr != nil {
				slog.Warn("entity search family failed", "family", f, "error", appErr.Message)
				mu.Lock()
				failed++
				mu.Unlock()
				items = []dto.EntitySearchItem{}
			}
			groups[i] = dto.EntitySearchGroup{Family: f, Total: total, Items: items}
		}(i, f)
	}
	wg.Wait()

	if failed == len(families) {
		return nil, errors.ErrInternal("Galgame 资料库搜索失败")
	}
	return groups, nil
}

func (s *EntitySearchService) searchFamily(
	ctx context.Context,
	family, keywords string,
	limit int,
	isSFW bool,
) ([]dto.EntitySearchItem, int64, *errors.AppError) {
	if family == EntityFamilyTag {
		hits, total, appErr := s.tagService.searchHits(ctx, keywords, limit, isSFW)
		if appErr != nil {
			return nil, 0, appErr
		}
		items := make([]dto.EntitySearchItem, 0, len(hits))
		for _, h := range hits {
			items = append(items, dto.EntitySearchItem{
				ID:     int(h.ID),
				Family: family,
				Name:   h.VocabularyName(),
			})
		}
		return items, total, nil
	}

	hits, total, appErr := s.galgameClient.CatalogEntitySearch(
		ctx, entityCatalogType[family], keywords, limit)
	if appErr != nil {
		return nil, 0, appErr
	}
	items := make([]dto.EntitySearchItem, 0, len(hits))
	for _, h := range hits {
		name := h.Name(ctx)
		items = append(items, dto.EntitySearchItem{
			ID:     int(h.ID),
			Family: family,
			Name:   name,
			Alias:  entityAlias(h, name),
		})
	}
	return items, total, nil
}

func entityAlias(h client.CatalogEntityHit, name string) string {
	for _, candidate := range []string{h.DisplayName, h.Latin} {
		if candidate != "" && candidate != name {
			return candidate
		}
	}
	return ""
}

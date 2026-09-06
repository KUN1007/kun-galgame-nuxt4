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
	EntityFamilySeries    = "series"
	EntityFamilyEngine    = "engine"
)

// catalog answers one family per request, so a page is a page of one family.
var entityFamilies = []string{
	EntityFamilyCharacter,
	EntityFamilyCompany,
	EntityFamilyStaff,
	EntityFamilyTag,
	EntityFamilySeries,
	EntityFamilyEngine,
}

var entityCatalogType = map[string]string{
	EntityFamilyCharacter: "characters",
	EntityFamilyCompany:   "labels",
	EntityFamilyStaff:     "names",
	EntityFamilyTag:       "tags",
	EntityFamilySeries:    "series",
	EntityFamilyEngine:    "engines",
}

// A credit name has neither a picture nor a work count of its own, so its cards
// stay on the family icon; the rest pay for one batch request each. Only the
// character batch comes back with art — a series and an engine have no logo.
var entityMediaType = map[string]string{
	EntityFamilyCharacter: "characters",
	EntityFamilyCompany:   "labels",
	EntityFamilyTag:       "tags",
	EntityFamilySeries:    "series",
	EntityFamilyEngine:    "engines",
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
	page, limit int,
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

			items, total, appErr := s.searchFamily(ctx, f, keywords, page, limit, isSFW)
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
	page, limit int,
	isSFW bool,
) ([]dto.EntitySearchItem, int64, *errors.AppError) {
	if family == EntityFamilyTag {
		hits, total, appErr := s.tagService.searchHits(ctx, keywords, page, limit, isSFW)
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
		s.attachMedia(ctx, family, items)
		return items, total, nil
	}

	hits, total, appErr := s.galgameClient.CatalogEntitySearch(
		ctx, entityCatalogType[family], keywords, page, limit)
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
	s.attachMedia(ctx, family, items)
	return items, total, nil
}

// attachMedia is decoration: a card without its face is still a usable result,
// so a failed batch is logged and the family keeps its rows.
func (s *EntitySearchService) attachMedia(
	ctx context.Context, family string, items []dto.EntitySearchItem,
) {
	entity, ok := entityMediaType[family]
	if !ok || len(items) == 0 {
		return
	}
	ids := make([]int64, len(items))
	for i, item := range items {
		ids[i] = int64(item.ID)
	}
	media, appErr := s.galgameClient.CatalogEntityMediaBatch(ctx, entity, ids)
	if appErr != nil {
		slog.Warn("entity media batch failed", "family", family, "error", appErr.Message)
		return
	}
	for i := range items {
		m, ok := media[int64(items[i].ID)]
		if !ok {
			continue
		}
		items[i].Image, items[i].WorkCount = m.Image, m.WorkCount
	}
}

func entityAlias(h client.CatalogEntityHit, name string) string {
	for _, candidate := range []string{h.DisplayName, h.Latin} {
		if candidate != "" && candidate != name {
			return candidate
		}
	}
	return ""
}

// A filter chip keys on a catalog id, so a shared link arrives with ids and no
// names. Only the two families the filter bar can pick are resolvable.
var entityResolveType = map[string]string{
	EntityFamilyCompany: "labels",
	EntityFamilyTag:     "tags",
}

// Resolve names ids the reader already chose. The tag gate is the same one
// searchFamily applies: a hidden-tier tag, or a sexual tag in front of an SFW
// reader, is dropped rather than named — the chip then reads as its bare id,
// which is the correct amount to tell a reader who is not allowed to see it.
func (s *EntitySearchService) Resolve(
	ctx context.Context,
	family string,
	ids []int,
	isSFW bool,
) ([]dto.EntitySearchItem, *errors.AppError) {
	entity, ok := entityResolveType[family]
	if !ok {
		return nil, errors.ErrBadRequest("该实体类型不支持按 ID 解析")
	}
	if len(ids) == 0 {
		return []dto.EntitySearchItem{}, nil
	}

	rows, appErr := s.galgameClient.CatalogTaxonomyByIDs(ctx, entity, ids)
	if appErr != nil {
		return nil, appErr
	}

	items := make([]dto.EntitySearchItem, 0, len(rows))
	for _, row := range rows {
		if family == EntityFamilyTag {
			if row.Tier == client.TagTierHidden || (isSFW && row.Sexual) {
				continue
			}
		}
		name := row.Label(ctx)
		if family == EntityFamilyTag {
			name = row.VocabularyLabel()
		}
		items = append(items, dto.EntitySearchItem{
			ID:        int(row.ID),
			Family:    family,
			Name:      name,
			WorkCount: row.WorkCount,
		})
	}
	return items, nil
}

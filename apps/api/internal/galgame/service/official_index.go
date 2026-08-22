package service

import (
	"context"
	"net/url"
	"sort"
	"strconv"
	"sync"
	"time"

	"kun-galgame-api/internal/galgame/client"
	"kun-galgame-api/internal/galgame/dto"
	"kun-galgame-api/pkg/errors"
)

const (
	officialIndexTTL       = 30 * time.Minute
	officialIndexPageSize  = 100
	officialIndexCap       = 200000
	officialRefreshTimeout = 3 * time.Minute
)

type officialIndex struct {
	items   []dto.OfficialListItem
	builtAt time.Time
}

type officialIndexCache struct {
	mu       sync.RWMutex
	byKind   map[string]*officialIndex
	building map[string]struct{}
}

func newOfficialIndexCache() *officialIndexCache {
	return &officialIndexCache{
		byKind:   map[string]*officialIndex{},
		building: map[string]struct{}{},
	}
}

func (c *officialIndexCache) get(kind string) *officialIndex {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.byKind[kind]
}

func (c *officialIndexCache) put(kind string, idx *officialIndex) {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.byKind[kind] = idx
}

func (c *officialIndexCache) claim(kind string) bool {
	c.mu.Lock()
	defer c.mu.Unlock()
	if _, busy := c.building[kind]; busy {
		return false
	}
	c.building[kind] = struct{}{}
	return true
}

func (c *officialIndexCache) release(kind string) {
	c.mu.Lock()
	defer c.mu.Unlock()
	delete(c.building, kind)
}

func (s *OfficialService) page(
	ctx context.Context, kind string, page, limit int,
) ([]dto.OfficialListItem, int64, *errors.AppError) {
	idx := s.index.get(kind)
	if idx == nil {
		built, appErr := s.buildIndex(ctx, kind)
		if appErr != nil {
			return nil, 0, appErr
		}
		s.index.put(kind, built)
		idx = built
	} else if time.Since(idx.builtAt) > officialIndexTTL {
		s.refreshIndexAsync(kind)
	}
	return sliceOfficialPage(idx.items, page, limit), int64(len(idx.items)), nil
}

func (s *OfficialService) refreshIndexAsync(kind string) {
	if !s.index.claim(kind) {
		return
	}
	go func() {
		defer s.index.release(kind)
		ctx, cancel := context.WithTimeout(context.Background(), officialRefreshTimeout)
		defer cancel()
		if built, appErr := s.buildIndex(ctx, kind); appErr == nil {
			s.index.put(kind, built)
		}
	}()
}

func (s *OfficialService) buildIndex(ctx context.Context, kind string) (*officialIndex, *errors.AppError) {
	base := client.OpenPopulation(url.Values{"has_works": {"1"}})
	if kind != "" {
		base.Set("kind", kind)
	}

	items := make([]dto.OfficialListItem, 0, 4096)
	seenCursor := map[string]struct{}{}
	cursor := ""
	for {
		q := url.Values{}
		for k, v := range base {
			q[k] = v
		}
		q.Set("limit", strconv.Itoa(officialIndexPageSize))
		if cursor != "" {
			q.Set("cursor", cursor)
		}
		res, appErr := s.galgameClient.CatalogTaxonomyList(ctx, "labels", q)
		if appErr != nil {
			return nil, appErr
		}
		for _, o := range res.Items {
			items = append(items, s.officialRow(ctx, o))
		}
		if res.NextCursor == nil || *res.NextCursor == "" || len(items) >= officialIndexCap {
			break
		}
		cursor = *res.NextCursor
		if _, seen := seenCursor[cursor]; seen {
			break
		}
		seenCursor[cursor] = struct{}{}
	}

	sortOfficialsByCount(items)
	return &officialIndex{items: items, builtAt: time.Now()}, nil
}

func sortOfficialsByCount(items []dto.OfficialListItem) {
	sort.Slice(items, func(i, j int) bool {
		if items[i].GalgameCount != items[j].GalgameCount {
			return items[i].GalgameCount > items[j].GalgameCount
		}
		return items[i].Name < items[j].Name
	})
}

func (s *OfficialService) officialRow(ctx context.Context, o client.CatalogTaxonomyItem) dto.OfficialListItem {
	name := o.Label(ctx)
	return dto.OfficialListItem{
		ID:           int(o.ID),
		Name:         name,
		Category:     o.Kind,
		Logo:         s.galgameClient.ImageURLFromHash(o.LogoHash),
		Alias:        o.Aliases.Values(name),
		GalgameCount: o.WorkCount,
	}
}

func sliceOfficialPage(items []dto.OfficialListItem, page, limit int) []dto.OfficialListItem {
	if page < 1 {
		page = 1
	}
	if limit < 1 {
		limit = 50
	}
	start := (page - 1) * limit
	if start >= len(items) {
		return []dto.OfficialListItem{}
	}
	end := min(start+limit, len(items))
	return items[start:end]
}

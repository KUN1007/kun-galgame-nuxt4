package client

import (
	"cmp"
	"context"
	"encoding/json"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"time"

	"kun-galgame-api/internal/galgame/dto"
	"kun-galgame-api/pkg/errors"
	"kun-galgame-api/pkg/namepref"
)

const briefCacheTTL = 2 * time.Minute

const batchCacheMaxEntries = 4096

// origName is part of the key because the cached values hold names this
// request already rendered: without it, the first reader to warm an entry
// decides which language every other reader sees for the next two minutes.
type batchCacheKey struct {
	id       int
	sfw      bool
	origName bool
}

type batchCacheEntry[T any] struct {
	found  bool
	val    T
	expire time.Time
}

func cachedBatch[T any](
	ctx context.Context,
	mu *sync.RWMutex,
	cache map[batchCacheKey]batchCacheEntry[T],
	ids []int,
	sfw bool,
	fetch func([]int) (map[int]T, *errors.AppError),
) (map[int]T, *errors.AppError) {
	origName := namepref.PrefersOriginal(ctx)
	result := make(map[int]T, len(ids))
	var missing []int
	now := time.Now()
	mu.RLock()
	for _, id := range ids {
		if e, ok := cache[batchCacheKey{id, sfw, origName}]; ok && now.Before(e.expire) {
			if e.found {
				result[id] = e.val
			}
		} else {
			missing = append(missing, id)
		}
	}
	mu.RUnlock()
	if len(missing) == 0 {
		return result, nil
	}
	fetched, appErr := fetch(missing)
	if appErr != nil {
		return nil, appErr
	}
	expire := now.Add(briefCacheTTL)
	mu.Lock()
	if len(cache) > batchCacheMaxEntries {
		clear(cache)
	}
	for _, id := range missing {
		v, ok := fetched[id]
		cache[batchCacheKey{id, sfw, origName}] = batchCacheEntry[T]{found: ok, val: v, expire: expire}
		if ok {
			result[id] = v
		}
	}
	mu.Unlock()
	return result, nil
}

type GalgameClient struct {
	origin       string
	apiKey       string
	httpClient   *http.Client
	imageCDNBase string
	imageMeta    ImageMetaResolver

	briefMu     sync.RWMutex
	briefCache  map[batchCacheKey]batchCacheEntry[GalgameBrief]
	detailMu    sync.RWMutex
	detailCache map[batchCacheKey]batchCacheEntry[GalgameDetailBrief]

	labelLinkMu    sync.RWMutex
	labelLinkCache map[batchCacheKey]batchCacheEntry[string]

	tagSexualMu    sync.RWMutex
	tagSexualCache map[batchCacheKey]batchCacheEntry[bool]

	gidMu    sync.RWMutex
	gidCache map[int]gidLookupEntry
}

func New(baseURL, apiKey, imageCDNBase string) *GalgameClient {
	base := strings.TrimRight(baseURL, "/")

	// net/http defaults MaxIdleConnsPerHost to 2, which throttles a single-host
	// S2S client: concurrent callers cannot reuse keep-alives past 2 and pay a
	// fresh dial per request.
	transport := http.DefaultTransport.(*http.Transport).Clone()
	transport.MaxIdleConns = 100
	transport.MaxIdleConnsPerHost = 64

	return &GalgameClient{
		origin: catalogOrigin(base),
		apiKey: apiKey,
		httpClient: &http.Client{
			Timeout:   10 * time.Second,
			Transport: transport,
			// Never follow a redirect. The catalog answers a merged entity id with
			// 301 + current_id so the caller can redirect the BROWSER in one hop;
			// auto-following swallows that and returns the survivor's record under
			// the dead id — a duplicate page on two URLs, which is what the 301
			// exists to prevent.
			CheckRedirect: func(*http.Request, []*http.Request) error {
				return http.ErrUseLastResponse
			},
		},
		imageCDNBase: imageCDNBase,
		briefCache:   map[batchCacheKey]batchCacheEntry[GalgameBrief]{},
		detailCache:  map[batchCacheKey]batchCacheEntry[GalgameDetailBrief]{},
		gidCache:     map[int]gidLookupEntry{},

		labelLinkCache: map[batchCacheKey]batchCacheEntry[string]{},
		tagSexualCache: map[batchCacheKey]batchCacheEntry[bool]{},
	}
}

type apiResponse struct {
	Code    int             `json:"code"`
	Message string          `json:"message"`
	Data    json.RawMessage `json:"data"`
}

func (c *GalgameClient) GetV1(ctx context.Context, path string, query url.Values) (json.RawMessage, *errors.AppError) {
	if catalogReadsV1(path, query) {
		status, env, appErr := c.doV1(ctx, path, query)
		if appErr != nil {
			return nil, appErr
		}
		if status >= 200 && status < 300 && env.Code == 0 {
			return rewriteBanners(env.Data, c.imageCDNBase), nil
		}
		return nil, errors.New(errors.CodeBiz, cmp.Or(env.Message, "Galgame 资源不存在"), status)
	}
	status, raw, appErr := c.doV2(ctx, http.MethodGet, path, query, nil)
	if appErr != nil {
		return nil, appErr
	}
	if status >= 200 && status < 300 {
		return rewriteBanners(raw, c.imageCDNBase), nil
	}
	st, _, err := c.v2Error(status, raw)
	if err != nil {
		return nil, err
	}
	return nil, errors.New(errors.CodeBiz, "Galgame 资源不存在", st)
}

const catalogMovedCode = 12

func (c *GalgameClient) getV1Envelope(ctx context.Context, path string, query url.Values) (int, *apiResponse, *errors.AppError) {
	if catalogReadsV1(path, query) {
		status, env, appErr := c.doV1(ctx, path, query)
		if appErr != nil {
			return 0, nil, appErr
		}
		if status >= 200 && status < 300 {
			env.Data = rewriteBanners(env.Data, c.imageCDNBase)
		}
		return status, env, nil
	}
	status, raw, appErr := c.doV2(ctx, http.MethodGet, path, query, nil)
	if appErr != nil {
		return 0, nil, appErr
	}
	if status >= 200 && status < 300 {
		return status, &apiResponse{Data: rewriteBanners(raw, c.imageCDNBase)}, nil
	}
	return c.v2Error(status, raw)
}

func BriefName(b *GalgameBrief) string {
	if b == nil {
		return ""
	}
	return b.Name
}

type GalgameBrief struct {
	ID                  int     `json:"id"`
	WorkID              int64   `json:"work_id"`
	VndbID              string  `json:"vndb_id"`
	Name                string  `json:"name"`
	NameOriginal        string  `json:"name_original"`
	Status              int     `json:"status"`
	ClaimState          string  `json:"claim_state,omitempty"`
	ContentLimit        string  `json:"content_limit"`
	UserID              int     `json:"user_id"`
	OriginalLanguage    string  `json:"original_language"`
	AgeLimit            string  `json:"age_limit"`
	ReleaseDate         *string `json:"release_date"`
	ReleaseDateTBA      bool    `json:"release_date_tba"`
	EffectiveBannerHash string  `json:"effective_banner_hash"`
	// rewriteBanners injects effective_banner_url into the raw JSON BEFORE this
	// struct is unmarshalled. The field must be declared or Go silently drops
	// the walker's work and every downstream DTO is left with only the hash.
	EffectiveBannerURL       string `json:"effective_banner_url"`
	EffectiveBannerWidth     int    `json:"effective_banner_width,omitempty"`
	EffectiveBannerHeight    int    `json:"effective_banner_height,omitempty"`
	EffectiveBannerThumbhash string `json:"effective_banner_thumbhash,omitempty"`
	// cover_slots.portrait, kept apart from the banner because the two slots can
	// be the same image: catalog fills portrait from any cover when the work has
	// no portrait-shaped one, so a card must crop rather than trust the ratio.
	EffectivePortraitHash      string            `json:"effective_portrait_hash,omitempty"`
	EffectivePortraitURL       string            `json:"effective_portrait_url,omitempty"`
	EffectivePortraitWidth     int               `json:"effective_portrait_width,omitempty"`
	EffectivePortraitHeight    int               `json:"effective_portrait_height,omitempty"`
	EffectivePortraitThumbhash string            `json:"effective_portrait_thumbhash,omitempty"`
	Refs                       map[string]string `json:"refs,omitempty"`
	Company                    string            `json:"company,omitempty"`
}

func (b GalgameBrief) DlsiteWorkno() string { return b.Refs["dlsite"] }

type GalgameDetailBrief struct {
	GalgameBrief
	Intros    []dto.GalgameIntro `json:"intros"`
	Officials []string           `json:"officials"`
}

func (c *GalgameClient) GetBatchDetailPublic(ctx context.Context, ids []int, isSFW bool) (map[int]GalgameDetailBrief, *errors.AppError) {
	return cachedBatch(ctx, &c.detailMu, c.detailCache, ids, isSFW, func(miss []int) (map[int]GalgameDetailBrief, *errors.AppError) {
		rows, appErr := c.CatalogRowsByGIDs(ctx, miss, catalogDetailBriefInclude, contentLimitFor(isSFW))
		if appErr != nil {
			return nil, appErr
		}
		result := make(map[int]GalgameDetailBrief, len(rows))
		for gid := range rows {
			row := rows[gid]
			result[gid] = CatalogItemToDetailBrief(ctx, &row)
		}
		return result, nil
	})
}

func (c *GalgameClient) GetBatch(ctx context.Context, ids []int) (map[int]GalgameBrief, *errors.AppError) {
	return c.batchByGIDs(ctx, ids, "all")
}

func (c *GalgameClient) GetBatchPublic(ctx context.Context, ids []int, isSFW bool) (map[int]GalgameBrief, *errors.AppError) {
	return cachedBatch(ctx, &c.briefMu, c.briefCache, ids, isSFW, func(miss []int) (map[int]GalgameBrief, *errors.AppError) {
		return c.batchByGIDs(ctx, miss, contentLimitFor(isSFW))
	})
}

func (c *GalgameClient) batchByGIDs(ctx context.Context, ids []int, contentLimit string) (map[int]GalgameBrief, *errors.AppError) {
	if len(ids) == 0 {
		return map[int]GalgameBrief{}, nil
	}
	rows, appErr := c.CatalogRowsByGIDs(ctx, ids, catalogBriefInclude, contentLimit)
	if appErr != nil {
		return nil, appErr
	}
	result := make(map[int]GalgameBrief, len(rows))
	for gid := range rows {
		row := rows[gid]
		result[gid] = CatalogItemToBrief(ctx, &row)
	}
	return result, nil
}

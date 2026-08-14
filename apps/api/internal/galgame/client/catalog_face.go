package client

import (
	"context"
	"encoding/json"
	"net/url"
	"strconv"
	"strings"
	"time"

	"kun-galgame-api/pkg/errors"
)

const catalogIDsChunk = 100

var anchorSourceKeys = []string{"curated", "galgame_wiki"}

const (
	gidLookupHitTTL  = 30 * time.Minute
	gidLookupMissTTL = 2 * time.Minute
)

type gidLookupEntry struct {
	catalogID int64
	found     bool
	expire    time.Time
}

const (
	catalogBriefInclude       = "names,covers,refs"
	catalogDetailBriefInclude = "names,intros,labels,covers,refs"
)

func openPopulation(q url.Values) url.Values {
	q.Set("nsfw", "1")
	return q
}

func OpenPopulation(q url.Values) url.Values { return openPopulation(q) }

func applyWorksGate(q url.Values, contentLimit string) url.Values {
	openPopulation(q)
	switch contentLimit {
	case "sfw", "nsfw":
		q.Set("content_limit", contentLimit)
	}
	return q
}

func ApplyWorksGate(q url.Values, isSFW bool) url.Values {
	return applyWorksGate(q, contentLimitFor(isSFW))
}

func contentLimitFor(isSFW bool) string {
	if isSFW {
		return "sfw"
	}
	return "all"
}

func (c *GalgameClient) catalogIDsForGIDs(ctx context.Context, gids []int) (map[int]int64, *errors.AppError) {
	out := make(map[int]int64, len(gids))
	var missing []int
	now := time.Now()

	c.gidMu.RLock()
	for _, gid := range gids {
		if e, ok := c.gidCache[gid]; ok && now.Before(e.expire) {
			if e.found {
				out[gid] = e.catalogID
			}
		} else {
			missing = append(missing, gid)
		}
	}
	c.gidMu.RUnlock()
	if len(missing) == 0 {
		return out, nil
	}

	gidStride := max(catalogIDsChunk/len(anchorSourceKeys), 1)

	resolved := make(map[int]int64, len(missing))
	for start := 0; start < len(missing); start += gidStride {
		end := min(start+gidStride, len(missing))
		chunk := missing[start:end]

		items := make([]map[string]string, 0, len(chunk)*len(anchorSourceKeys))
		for _, gid := range chunk {
			for _, source := range anchorSourceKeys {
				items = append(items, map[string]string{
					"source":      source,
					"external_id": strconv.Itoa(gid),
					"type":        "work",
				})
			}
		}
		data, appErr := c.postFace(ctx, c.v1Base, "/catalog/lookup/batch",
			url.Values{"nsfw": {"1"}}, map[string]any{"items": items}, c.apiKey)
		if appErr != nil {
			return nil, appErr
		}
		var parsed struct {
			Items []struct {
				ExternalID string `json:"external_id"`
				Work       *struct {
					ID int64 `json:"id"`
				} `json:"work"`
			} `json:"items"`
		}
		if err := json.Unmarshal(data, &parsed); err != nil {
			return nil, errors.ErrInternal("解析 Catalog 批量解析响应失败")
		}
		for _, it := range parsed.Items {
			if it.Work == nil {
				continue
			}
			gid, err := strconv.Atoi(it.ExternalID)
			if err != nil {
				continue
			}
			resolved[gid] = it.Work.ID
		}
	}

	var unresolved []int
	for _, gid := range missing {
		if _, ok := resolved[gid]; !ok {
			unresolved = append(unresolved, gid)
		}
	}
	if len(unresolved) > 0 {
		adopted, appErr := c.adoptedWorkIDs(ctx, unresolved)
		if appErr != nil {
			return nil, appErr
		}
		for gid, id := range adopted {
			resolved[gid] = id
			out[gid] = id
		}
	}

	c.gidMu.Lock()
	if len(c.gidCache) > batchCacheMaxEntries {
		clear(c.gidCache)
	}
	for _, gid := range missing {
		id, ok := resolved[gid]
		ttl := gidLookupMissTTL
		if ok {
			ttl = gidLookupHitTTL
			out[gid] = id
		}
		c.gidCache[gid] = gidLookupEntry{catalogID: id, found: ok, expire: now.Add(ttl)}
	}
	c.gidMu.Unlock()
	return out, nil
}

// The second half of the gid bridge. A work minted through the submission face
// carries NO external_ref anchor — there is no upstream to have issued one — so
// the anchor lookup answers "no such work" rather than an error, and every page
// of a post-switchover entry would 404 silently.
//
// THE ROUND-TRIP CHECK IS NOT DEFENSIVE, IT IS THE WHOLE CORRECTNESS ARGUMENT:
// a legacy gid is also a syntactically valid work id, so resolving gid 42 by
// fetching work 42 would hand back a different game. An adopted id satisfies
// `claimed_by.work_id == the id asked for` by construction; a coincidence does not.
//
// The content gates are open because this resolves an IDENTITY. Filtering it by
// the reader's preference would make an r18 entry unresolvable rather than
// merely invisible; visibility belongs to the row fetch that follows.
func (c *GalgameClient) adoptedWorkIDs(ctx context.Context, gids []int) (map[int]int64, *errors.AppError) {
	ids := make([]int64, len(gids))
	for i, gid := range gids {
		ids[i] = int64(gid)
	}
	rows, appErr := c.worksByCatalogIDs(ctx, ids, "", "all")
	if appErr != nil {
		return nil, appErr
	}
	out := make(map[int]int64, len(rows))
	for i := range rows {
		row := &rows[i]
		if gid := row.gid(); gid > 0 && int64(gid) == row.ID {
			out[gid] = row.ID
		}
	}
	return out, nil
}

func (c *GalgameClient) worksByCatalogIDs(ctx context.Context, ids []int64, include, contentLimit string) ([]CatalogWorkListItem, *errors.AppError) {
	var out []CatalogWorkListItem
	for start := 0; start < len(ids); start += catalogIDsChunk {
		end := min(start+catalogIDsChunk, len(ids))
		chunk := ids[start:end]

		raw := make([]string, len(chunk))
		for i, id := range chunk {
			raw[i] = strconv.FormatInt(id, 10)
		}
		q := url.Values{
			"ids":   {strings.Join(raw, ",")},
			"limit": {strconv.Itoa(catalogIDsChunk)},
		}
		if include != "" {
			q.Set("include", include)
		}
		applyWorksGate(q, contentLimit)

		data, appErr := c.GetV1(ctx, "/catalog/works", q)
		if appErr != nil {
			return nil, appErr
		}
		var parsed catWorksListData
		if err := json.Unmarshal(data, &parsed); err != nil {
			return nil, errors.ErrInternal("解析 Catalog 作品列表响应失败")
		}
		out = append(out, parsed.Items...)
	}
	return out, nil
}

func (c *GalgameClient) CatalogWorkIDs(ctx context.Context, gids []int) (map[int]int64, *errors.AppError) {
	if len(gids) == 0 {
		return map[int]int64{}, nil
	}
	return c.catalogIDsForGIDs(ctx, gids)
}

func (c *GalgameClient) GIDsByCatalogIDs(ctx context.Context, ids []int64) (map[int64]int, *errors.AppError) {
	if len(ids) == 0 {
		return map[int64]int{}, nil
	}
	rows, appErr := c.worksByCatalogIDs(ctx, ids, "", "all")
	if appErr != nil {
		return nil, appErr
	}
	out := make(map[int64]int, len(rows))
	for i := range rows {
		if gid := rows[i].gid(); gid > 0 {
			out[rows[i].ID] = gid
		}
	}
	return out, nil
}

func (c *GalgameClient) CatalogRowsByGIDs(ctx context.Context, gids []int, include, contentLimit string) (map[int]CatalogWorkListItem, *errors.AppError) {
	if len(gids) == 0 {
		return map[int]CatalogWorkListItem{}, nil
	}
	idMap, appErr := c.catalogIDsForGIDs(ctx, gids)
	if appErr != nil {
		return nil, appErr
	}
	if len(idMap) == 0 {
		return map[int]CatalogWorkListItem{}, nil
	}
	ids := make([]int64, 0, len(idMap))
	for _, id := range idMap {
		ids = append(ids, id)
	}
	rows, appErr := c.worksByCatalogIDs(ctx, ids, include, contentLimit)
	if appErr != nil {
		return nil, appErr
	}
	out := make(map[int]CatalogWorkListItem, len(rows))
	for i := range rows {
		row := rows[i]
		if !row.isRenderable() {
			continue
		}
		if gid := row.gid(); gid > 0 {
			out[gid] = row
		}
	}
	return out, nil
}

type catWorkDetail struct {
	ID            int64         `json:"id"`
	DisplayName   string        `json:"display_name"`
	OLang         string        `json:"olang"`
	ContentRating string        `json:"content_rating"`
	ReleaseDate   *string       `json:"release_date"`
	Updated       string        `json:"updated"`
	Created       string        `json:"created"`
	ClaimedBy     *catClaimedBy `json:"claimed_by"`

	Titles []struct {
		Lang  string `json:"lang"`
		Title string `json:"title"`
		Kind  string `json:"kind"`
	} `json:"titles"`
	Refs []catRef `json:"refs"`
	Tags []struct {
		Name        string `json:"name"`
		Source      string `json:"source"`
		CanonicalID int64  `json:"canonical_id"`
		Tier        string `json:"tier"`
		Kind        string `json:"kind"`
		Spoiler     int    `json:"spoiler"`
		Sexual      bool   `json:"sexual"`
		WorkCount   int    `json:"work_count"`
	} `json:"tags"`
	Intro []struct {
		Lang    string `json:"lang"`
		Intro   string `json:"intro"`
		Machine bool   `json:"machine"`
	} `json:"intro"`
	Covers []struct {
		URL       string `json:"url"`
		Kind      string `json:"kind"`
		Sexual    int    `json:"sexual"`
		Violence  int    `json:"violence"`
		Width     int    `json:"width"`
		Height    int    `json:"height"`
		Thumbhash string `json:"thumbhash"`
	} `json:"covers"`
	CoverSlots  *catCoverSlots `json:"cover_slots"`
	Screenshots []struct {
		URL       string `json:"url"`
		Caption   string `json:"caption"`
		Sexual    int    `json:"sexual"`
		Violence  int    `json:"violence"`
		Width     int    `json:"width"`
		Height    int    `json:"height"`
		Thumbhash string `json:"thumbhash"`
	} `json:"screenshots"`
	Labels     []catWorkLabel     `json:"labels"`
	Engines    []catWorkEngine    `json:"engines"`
	Links      []catWorkLink      `json:"links"`
	Series     []catWorkSeries    `json:"series"`
	Credits    []catCreditGroup   `json:"credits"`
	Characters []catWorkCharacter `json:"characters"`
	Ratings    []catRating        `json:"ratings"`
}

type catWorkCharacter struct {
	ID      int64  `json:"id"`
	Name    string `json:"name"`
	Latin   string `json:"latin"`
	Kind    string `json:"kind"`
	Spoiler int    `json:"spoiler"`
	Image   string `json:"image"`
	Figure  string `json:"figure"`
	Voices  []struct {
		ID   int64  `json:"id"`
		Name string `json:"name"`
	} `json:"voices"`
	ImageMeta  ArtMeta `json:"-"`
	FigureMeta ArtMeta `json:"-"`
}

type catCreditGroup struct {
	RoleKey  string          `json:"role_key"`
	RoleName string          `json:"role_name"`
	Credits  []catCreditItem `json:"credits"`
}

type catCreditItem struct {
	ID          int64  `json:"id"`
	Name        string `json:"name"`
	Latin       string `json:"latin"`
	CharacterID int64  `json:"character_id"`
	Character   string `json:"character"`
}

type catWorkSeries struct {
	ID   int64  `json:"id"`
	Name string `json:"name"`
}

func (c *GalgameClient) CatalogWorkDetail(ctx context.Context, gid int) (*catWorkDetail, bool, *errors.AppError) {
	idMap, appErr := c.catalogIDsForGIDs(ctx, []int{gid})
	if appErr != nil {
		return nil, false, appErr
	}
	catalogID, ok := idMap[gid]
	if !ok {
		return nil, false, nil
	}
	q := url.Values{"spoilers": {"0"}, "include": {"credits"}}
	openPopulation(q)
	data, appErr := c.GetV1(ctx, "/catalog/works/"+strconv.FormatInt(catalogID, 10), q)
	if appErr != nil {
		if appErr.StatusCode == 404 {
			return nil, false, nil
		}
		return nil, false, appErr
	}
	var d catWorkDetail
	if err := json.Unmarshal(data, &d); err != nil {
		return nil, false, errors.ErrInternal("解析 Catalog 作品详情响应失败")
	}
	if d.ClaimedBy != nil && d.ClaimedBy.State == claimStateHidden {
		return nil, false, nil
	}
	c.hydrateRosterArt(d.Characters)
	return &d, true, nil
}

func (c *GalgameClient) hydrateRosterArt(chars []catWorkCharacter) {
	urls := make([]string, 0, len(chars)*2)
	for _, ch := range chars {
		urls = append(urls, ch.Image, ch.Figure)
	}
	meta := c.resolveArtMeta(urls)
	if meta == nil {
		return
	}
	for i := range chars {
		chars[i].ImageMeta = meta[chars[i].Image]
		chars[i].FigureMeta = meta[chars[i].Figure]
	}
}

type CatalogWorksPage struct {
	Items      []CatalogWorkListItem
	NextCursor string
	Total      int64
	Count      int64
	Month      string
	Year       string
	Meta       catCalendarMeta
}

func (c *GalgameClient) CatalogWorksList(ctx context.Context, q url.Values) (*CatalogWorksPage, *errors.AppError) {
	data, appErr := c.GetV1(ctx, "/catalog/works", q)
	if appErr != nil {
		return nil, appErr
	}
	var parsed catWorksListData
	if err := json.Unmarshal(data, &parsed); err != nil {
		return nil, errors.ErrInternal("解析 Catalog 作品列表响应失败")
	}
	page := &CatalogWorksPage{Items: parsed.Items}
	if parsed.NextCursor != nil {
		page.NextCursor = *parsed.NextCursor
	}
	return page, nil
}

func (c *GalgameClient) CatalogMemberGIDs(ctx context.Context, filter url.Values, isSFW bool, pageCap int) ([]int, *errors.AppError) {
	members, appErr := c.catalogMembers(ctx, filter, isSFW, pageCap)
	if appErr != nil {
		return nil, appErr
	}
	gids := make([]int, 0, len(members))
	for _, m := range members {
		gids = append(gids, m.GID)
	}
	return gids, nil
}

type CatalogRollupMember struct {
	GID int
	Via *CatalogLabelVia
}

func (c *GalgameClient) CatalogLabelRollupMembers(ctx context.Context, labelID string, isSFW bool, pageCap int) ([]CatalogRollupMember, *errors.AppError) {
	return c.catalogMembers(ctx,
		url.Values{"label_id": {labelID}, "label_rollup": {"1"}}, isSFW, pageCap)
}

// CatalogLabelMemberItems returns the label's LIVE members (direct + imprint
// rollup) as full catalog items. Unlike CatalogLabelRollupMembers it keeps
// works the forum hasn't claimed — they carry is_on_forum=false downstream —
// because the official detail page lists the catalog, not the forum's own
// galgame table.
func (c *GalgameClient) CatalogLabelMemberItems(ctx context.Context, labelID string, isSFW bool, pageCap int) ([]CatalogWorkListItem, *errors.AppError) {
	return c.catalogMemberItems(ctx,
		url.Values{"label_id": {labelID}, "label_rollup": {"1"}}, isSFW, pageCap)
}

func (c *GalgameClient) catalogMemberItems(ctx context.Context, filter url.Values, isSFW bool, pageCap int) ([]CatalogWorkListItem, *errors.AppError) {
	items := []CatalogWorkListItem{}
	cursor := ""
	for page := 0; page < pageCap; page++ {
		q := url.Values{}
		for k, v := range filter {
			q[k] = v
		}
		q.Set("limit", strconv.Itoa(catalogIDsChunk))
		q.Set("claimed", "true")
		q.Set("claim_state", claimStateLive)
		ApplyWorksGate(q, isSFW)
		if cursor != "" {
			q.Set("cursor", cursor)
		}
		res, appErr := c.CatalogWorksList(ctx, q)
		if appErr != nil {
			return nil, appErr
		}
		for i := range res.Items {
			if !res.Items[i].isRenderable() {
				continue
			}
			items = append(items, res.Items[i])
		}
		if res.NextCursor == "" {
			break
		}
		cursor = res.NextCursor
	}
	return items, nil
}

func (c *GalgameClient) catalogMembers(ctx context.Context, filter url.Values, isSFW bool, pageCap int) ([]CatalogRollupMember, *errors.AppError) {
	items, appErr := c.catalogMemberItems(ctx, filter, isSFW, pageCap)
	if appErr != nil {
		return nil, appErr
	}
	members := make([]CatalogRollupMember, 0, len(items))
	for i := range items {
		if gid := items[i].gid(); gid > 0 {
			members = append(members, CatalogRollupMember{
				GID: gid,
				Via: items[i].ViaLabel,
			})
		}
	}
	return members, nil
}

func (c *GalgameClient) CatalogWorksSearch(ctx context.Context, q url.Values) (*CatalogWorksPage, *errors.AppError) {
	data, appErr := c.GetV1(ctx, "/catalog/works/search", q)
	if appErr != nil {
		return nil, appErr
	}
	var parsed catWorksSearchData
	if err := json.Unmarshal(data, &parsed); err != nil {
		return nil, errors.ErrInternal("解析 Catalog 搜索响应失败")
	}
	return &CatalogWorksPage{Items: parsed.Items, Total: parsed.Total}, nil
}

func (c *GalgameClient) CatalogCalendar(ctx context.Context, bucket string, q url.Values) (*CatalogWorksPage, *errors.AppError) {
	data, appErr := c.GetV1(ctx, "/catalog/calendar"+bucket, q)
	if appErr != nil {
		return nil, appErr
	}
	var parsed catWorksListData
	if err := json.Unmarshal(data, &parsed); err != nil {
		return nil, errors.ErrInternal("解析 Catalog 月历响应失败")
	}
	page := &CatalogWorksPage{
		Items: parsed.Items, Count: parsed.Count,
		Month: parsed.Month, Year: parsed.Year, Meta: parsed.Meta,
	}
	if parsed.NextCursor != nil {
		page.NextCursor = *parsed.NextCursor
	}
	return page, nil
}

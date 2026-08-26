package client

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"net/url"
	"strconv"
	"strings"

	"kun-galgame-api/pkg/errors"
)

type v2Problem struct {
	Code      string `json:"code"`
	Title     string `json:"title"`
	Detail    string `json:"detail"`
	Status    int    `json:"status"`
	CurrentID string `json:"current_id"`
}

func catalogOrigin(raw string) string {
	u := strings.TrimRight(strings.TrimSpace(raw), "/")
	for _, suf := range []string{"/api/v1", "/v2", "/v1"} {
		if strings.HasSuffix(u, suf) {
			return strings.TrimSuffix(u, suf)
		}
	}
	return u
}

// Catalog reads answer from /v1. The v2 read faces are a strict subset and the
// cutover to them took the galgame detail page, every entity page and the whole
// site's request budget with it:
//
//   - The entity repr is id + name and nothing else. No intros, aliases, links,
//     logo_hash, traits, has_nsfw, engine description, series work_count, or tag
//     `sexual` — collect.CompanySpec/TagSpec/SeriesSpec/EngineSpec/CharacterSpec
//     declare no include tokens at all, so there is no query that asks for them.
//     The missing tag flag silently opened the NSFW tag gate: 3080 tags listed
//     for an SFW reader where v1 lists 2339.
//   - The work detail types covers[].sexual as "safe" where v1 sends 0, so
//     include=covers makes the whole detail fail to decode, and its ratings
//     block carries no distribution.
//   - The companies list dropped has_works=, and /v2 rate-limits on CLIENT IP at
//     100/min ignoring the key's tier (apiv2/protocol/limit.go) where /v1 limits
//     per key and this one is unlimited. Paging 38,317 companies instead of
//     2,985 spent the backend's whole minute on one page view, and every read on
//     the site then answered "Short-window rate limit exceeded." The same ceiling
//     is fatal to every fan-out the forum runs: a tag with 7,465 works is 75
//     member pages, and the series card index is 1,562 searches.
//
// refs= is the exception and the reason this is a predicate rather than a
// deletion: only v2 resolves the source:external_id batch lane, and that lane is
// how a legacy gid finds its catalog work. v1 accepts the parameter and ignores
// it, which would answer page 1 of the whole catalogue.
func catalogReadsV1(path string, query url.Values) bool {
	if !strings.HasPrefix(strings.TrimPrefix(path, "/v1"), "/catalog") {
		return false
	}
	return query.Get("refs") == ""
}

func v2CatalogPath(path string) string {
	path = strings.TrimPrefix(path, "/v1")
	switch {
	case path == "/catalog/works/search":
		return "/v2/catalog/works"
	case strings.HasPrefix(path, "/catalog/labels/") && strings.HasSuffix(path, "/relation-graph"):
		id := strings.TrimSuffix(strings.TrimPrefix(path, "/catalog/labels/"), "/relation-graph")
		return "/v2/catalog/companies/" + id + "/graph"
	case strings.HasPrefix(path, "/catalog/labels"):
		return "/v2/catalog/companies" + strings.TrimPrefix(path, "/catalog/labels")
	case strings.HasPrefix(path, "/catalog/names"):
		return "/v2/catalog/credit-names" + strings.TrimPrefix(path, "/catalog/names")
	case path == "/catalog/calendar/pending", path == "/catalog/calendar/tba":
		return "/v2/catalog/calendar"
	case strings.HasPrefix(path, "/catalog/"):
		return "/v2" + path
	default:
		return "/v2/catalog" + path
	}
}

func v2CatalogQuery(path string, query url.Values) url.Values {
	out := url.Values{}
	for k, vs := range query {
		out[k] = append([]string(nil), vs...)
	}
	switch nsfw := out.Get("nsfw"); nsfw {
	case "1", "on", "yes":
		out.Set("nsfw", "true")
	case "0", "off", "no":
		out.Del("nsfw")
	}
	if inc := out.Get("include"); inc != "" {
		inc = strings.ReplaceAll(inc, "names", "titles")
		inc = strings.ReplaceAll(inc, "labels", "companies")
		out.Set("include", inc)
	}
	if lid := out.Get("label_id"); lid != "" {
		out.Del("label_id")
		out.Set("company_id", lid)
	}
	if out.Get("label_rollup") == "1" || out.Get("label_rollup") == "true" {
		out.Del("label_rollup")
		out.Set("company_rollup", "true")
	}
	if t := out.Get("type"); t != "" && (path == "/catalog/search" || strings.HasSuffix(v2CatalogPath(path), "/search")) {
		out.Del("type")
		out.Set("object", v2SearchObject(t))
	}
	if path == "/catalog/works/search" || path == "/catalog/works" {
		if out.Get("include_total") == "" {
			out.Set("include_total", "true")
		}
		convertedPage := false
		if page := out.Get("page"); page != "" {
			out.Del("page")
			if n, err := strconv.Atoi(page); err == nil && n > 1 {
				out.Set("cursor", encodePageCursor(n))
				convertedPage = true
			}
		}
		if out.Get("q") == "" && out.Get("ids") == "" && out.Get("refs") == "" && out.Get("facets") == "" {
			if path == "/catalog/works/search" || convertedPage {
				out.Set("facets", "olang,tag_id")
			}
		}
	}
	// v2 folded v1's three calendar routes into one windowed face, and both
	// "announced" and "unknown" resolve to the SAME undated bucket there — so
	// 年内待定 and 发售日期未定 served identical rows. The year-known bucket is
	// selected by precision, not by a status.
	switch path {
	case "/catalog/calendar/pending":
		out.Set("precision", "year")
	case "/catalog/calendar/tba":
		out.Set("status", "unknown")
	}
	// v2's calendar answers with items and next_cursor only, so the month
	// header's 共 N 部 has no source unless the total is asked for.
	if strings.HasPrefix(path, "/catalog/calendar") && out.Get("include_total") == "" {
		out.Set("include_total", "true")
	}
	return out
}

func v2SearchObject(t string) string {
	switch t {
	case "labels", "label", "officials", "official":
		return "company"
	case "names", "name", "staff":
		return "credit_name"
	case "characters", "character":
		return "character"
	case "tags", "tag":
		return "tag"
	case "engines", "engine":
		return "engine"
	case "series":
		return "series"
	case "works", "work", "galgame":
		return "work"
	default:
		return t
	}
}

func encodePageCursor(page int) string {
	return "cur_" + base64.RawURLEncoding.EncodeToString([]byte(strconv.Itoa(page)))
}

func (c *GalgameClient) doV2(ctx context.Context, method, path string, query url.Values, body any) (int, []byte, *errors.AppError) {
	v2Path := v2CatalogPath(path)
	q := v2CatalogQuery(path, query)
	reqURL := c.origin + v2Path
	if len(q) > 0 {
		reqURL += "?" + q.Encode()
	}
	var rdr io.Reader
	if body != nil {
		raw, err := json.Marshal(body)
		if err != nil {
			return 0, nil, errors.ErrInternal("序列化请求失败")
		}
		rdr = bytes.NewReader(raw)
	}
	req, err := http.NewRequestWithContext(ctx, method, reqURL, rdr)
	if err != nil {
		return 0, nil, errors.ErrInternal("创建请求失败")
	}
	if c.apiKey != "" {
		req.Header.Set("Authorization", "Bearer "+c.apiKey)
	}
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	resp, err := c.httpClient.Do(req)
	if err != nil {
		slog.Error("Galgame 服务请求失败 (传输层)",
			"method", req.Method, "url", req.URL.String(), "error", err)
		return 0, nil, errors.ErrInternal(fmt.Sprintf("Galgame 服务不可达: %v", err))
	}
	defer resp.Body.Close()
	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return resp.StatusCode, nil, errors.ErrInternal("读取 Galgame 响应失败")
	}
	return resp.StatusCode, rewriteV2JSON(respBody, c.imageCDNBase), nil
}

func (c *GalgameClient) doV1(ctx context.Context, path string, query url.Values) (int, *apiResponse, *errors.AppError) {
	reqURL := c.origin + "/v1" + strings.TrimPrefix(path, "/v1")
	if len(query) > 0 {
		reqURL += "?" + query.Encode()
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, reqURL, nil)
	if err != nil {
		return 0, nil, errors.ErrInternal("创建请求失败")
	}
	if c.apiKey != "" {
		req.Header.Set("Authorization", "Bearer "+c.apiKey)
	}
	resp, err := c.httpClient.Do(req)
	if err != nil {
		slog.Error("Galgame 服务请求失败 (传输层)",
			"method", req.Method, "url", req.URL.String(), "error", err)
		return 0, nil, errors.ErrInternal(fmt.Sprintf("Galgame 服务不可达: %v", err))
	}
	defer resp.Body.Close()
	raw, err := io.ReadAll(resp.Body)
	if err != nil {
		return resp.StatusCode, nil, errors.ErrInternal("读取 Galgame 响应失败")
	}
	var env apiResponse
	if len(bytes.TrimSpace(raw)) > 0 {
		if err := json.Unmarshal(raw, &env); err != nil {
			return resp.StatusCode, nil, errors.ErrInternal("解析 Galgame 响应失败")
		}
	}
	return resp.StatusCode, &env, nil
}

func (c *GalgameClient) v2Error(status int, raw []byte) (int, *apiResponse, *errors.AppError) {
	var p v2Problem
	_ = json.Unmarshal(raw, &p)
	if status == http.StatusMovedPermanently {
		var env apiResponse
		if json.Unmarshal(raw, &env) == nil && env.Code == catalogMovedCode {
			return status, &env, nil
		}
	}
	if p.Code == "ENTITY_MERGED" && p.CurrentID != "" {
		id, _ := strconv.ParseInt(p.CurrentID, 10, 64)
		data, _ := json.Marshal(map[string]any{"current_id": id})
		return http.StatusMovedPermanently, &apiResponse{Code: catalogMovedCode, Data: data}, nil
	}
	msg := p.Detail
	if msg == "" {
		msg = p.Title
	}
	if msg == "" {
		msg = fmt.Sprintf("Galgame 服务返回了非预期响应 (HTTP %d)", status)
	}
	code := errors.CodeBiz
	if status == http.StatusNotFound {
		return status, &apiResponse{Code: code, Message: msg}, nil
	}
	return status, nil, errors.New(code, msg, status)
}

func envelopeOK(m map[string]any) bool {
	switch c := m["code"].(type) {
	case json.Number:
		return c == "0"
	case float64:
		return c == 0
	case int:
		return c == 0
	default:
		return false
	}
}

func rewriteV2JSON(raw []byte, cdnBase string) []byte {
	if len(bytes.TrimSpace(raw)) == 0 {
		return raw
	}
	dec := json.NewDecoder(bytes.NewReader(raw))
	dec.UseNumber()
	var tree any
	if err := dec.Decode(&tree); err != nil {
		return raw
	}
	if m, ok := tree.(map[string]any); ok {
		if _, has := m["object"]; !has {
			if data, ok := m["data"]; ok && envelopeOK(m) {
				tree = data
			}
		}
	}
	if !walkV2(tree, cdnBase) {
		if _, ok := tree.(map[string]any); !ok && tree == nil {
			return raw
		}
		out, err := json.Marshal(tree)
		if err != nil {
			return raw
		}
		return out
	}
	out, err := json.Marshal(tree)
	if err != nil {
		return raw
	}
	return out
}

func walkV2(node any, cdnBase string) bool {
	changed := false
	switch v := node.(type) {
	case map[string]any:
		changed = rewriteV2Object(v, cdnBase) || changed
		for k, child := range v {
			if k == "cover_slots" {
				continue
			}
			if walkV2(child, cdnBase) {
				changed = true
			}
		}
	case []any:
		for _, child := range v {
			if walkV2(child, cdnBase) {
				changed = true
			}
		}
	}
	return changed
}

func rewriteV2Object(v map[string]any, cdnBase string) bool {
	changed := false
	for _, key := range []string{"id", "work_id", "canonical_id", "character_id", "person_id", "from", "to", "from_id", "to_id"} {
		if n, ok := digitString(v[key]); ok {
			v[aliasIDKey(key)] = json.Number(n)
			if key == "from_id" {
				delete(v, "from_id")
			}
			if key == "to_id" {
				delete(v, "to_id")
			}
			changed = true
		}
	}
	if claim, ok := v["claim"].(map[string]any); ok {
		if _, has := v["claimed_by"]; !has {
			mapped := map[string]any{
				"site":          claim["site"],
				"state":         claim["state"],
				"content_limit": claim["content_limit"],
			}
			if n, ok := digitString(claim["site_work_id"]); ok {
				mapped["work_id"] = json.Number(n)
			} else if n, ok := digitString(claim["work_id"]); ok {
				mapped["work_id"] = json.Number(n)
			}
			v["claimed_by"] = mapped
			changed = true
		}
	}
	if companies, ok := v["companies"]; ok {
		if _, has := v["labels"]; !has {
			v["labels"] = companies
			changed = true
		}
	}
	if _, has := v["updated"]; !has {
		if u, ok := v["updated_at"]; ok {
			v["updated"] = u
			changed = true
		}
	}
	if _, has := v["created"]; !has {
		if u, ok := v["created_at"]; ok {
			v["created"] = u
			changed = true
		}
	}
	if machine, ok := v["is_machine"]; ok {
		if _, has := v["machine"]; !has {
			v["machine"] = machine
			changed = true
		}
	}
	if intro, ok := v["value"].(string); ok && v["intro"] == nil && (v["lang"] != nil || v["source"] != nil) {
		v["intro"] = intro
		changed = true
	}
	if sexual, ok := v["is_sexual"]; ok {
		if _, has := v["sexual"]; !has {
			v["sexual"] = sexual
			changed = true
		}
	}
	if kind, ok := v["company_kind"].(string); ok {
		if _, has := v["kind"]; !has {
			v["kind"] = kind
			changed = true
		}
		if _, has := v["label_kind"]; !has {
			v["label_kind"] = kind
			changed = true
		}
	}
	// v1 sent one `kind` that meant developer/publisher; v2 split it into
	// company_kind (what the company IS) and attribution_role (what it DID on
	// this work). Only the first was ever mapped, so the detail page printed a
	// raw "game_brand" as a role chip beside the 游戏品牌 category chip.
	if role, ok := v["attribution_role"].(string); ok {
		if _, has := v["role"]; !has {
			v["role"] = role
			changed = true
		}
	}
	if kind, ok := v["tag_kind"].(string); ok {
		if _, has := v["kind"]; !has {
			v["kind"] = kind
			changed = true
		}
	}
	if kind, ok := v["title_kind"].(string); ok {
		if _, has := v["kind"]; !has {
			v["kind"] = kind
			changed = true
		}
	}
	if _, has := v["cover_slots"]; !has {
		if slots, ok := v["covers"].(map[string]any); ok {
			v["cover_slots"] = slots
			changed = true
		} else if slots := coverSlotsFromImages(v, cdnBase); slots != nil {
			v["cover_slots"] = slots
			changed = true
		}
	}
	if img, ok := v["cover"].(map[string]any); ok {
		if url := imageURL(img, cdnBase); url != "" {
			v["cover"] = url
			changed = true
		}
	}
	for _, key := range []string{"image", "figure"} {
		if img, ok := v[key].(map[string]any); ok {
			if url := imageURL(img, cdnBase); url != "" {
				v[key] = url
				changed = true
			}
		}
	}
	if hash, ok := v["hash"].(string); ok && hash != "" {
		if url, _ := v["url"].(string); url == "" {
			if u := bannerURLFromHash(cdnBase, hash); u != "" {
				v["url"] = u
				changed = true
			}
		}
	}
	if role, ok := v["roster_role"].(string); ok {
		if _, has := v["kind"]; !has {
			v["kind"] = role
			changed = true
		}
	}
	if spoiler, ok := v["spoiler"].(string); ok {
		v["spoiler"] = spoilerLevel(spoiler)
		changed = true
	}
	if obj, ok := v["object"].(string); ok && obj == "search_result" {
		if t, ok := v["target_object"].(string); ok {
			v["entity_type"] = searchEntityType(t)
			changed = true
		}
	}
	return changed
}

func aliasIDKey(key string) string {
	switch key {
	case "from_id":
		return "from"
	case "to_id":
		return "to"
	default:
		return key
	}
}

func digitString(v any) (string, bool) {
	s, ok := v.(string)
	if !ok || s == "" {
		return "", false
	}
	for _, r := range s {
		if r < '0' || r > '9' {
			return "", false
		}
	}
	return s, true
}

func coverSlotsFromImages(v map[string]any, cdnBase string) map[string]any {
	if _, hasPortrait := v["portrait"]; hasPortrait {
		return nil
	}
	if !looksLikeV2Image(v["cover"]) && !looksLikeV2Image(v["banner"]) {
		return nil
	}
	portrait := imageToSlot(v["cover"], cdnBase)
	banner := imageToSlot(v["banner"], cdnBase)
	if portrait == nil && banner == nil {
		return nil
	}
	return map[string]any{"portrait": portrait, "banner": banner}
}

func looksLikeV2Image(v any) bool {
	img, ok := v.(map[string]any)
	if !ok {
		return false
	}
	if url, _ := img["url"].(string); url != "" {
		return true
	}
	hash, _ := img["hash"].(string)
	return hash != ""
}

func imageURL(img map[string]any, cdnBase string) string {
	if url, _ := img["url"].(string); url != "" {
		return url
	}
	hash, _ := img["hash"].(string)
	return bannerURLFromHash(cdnBase, hash)
}

func imageToSlot(v any, cdnBase string) map[string]any {
	img, ok := v.(map[string]any)
	if !ok {
		return nil
	}
	url := imageURL(img, cdnBase)
	if url == "" {
		return nil
	}
	slot := map[string]any{"url": url, "source": img["source"]}
	if w, ok := img["width"]; ok {
		slot["width"] = w
	}
	if h, ok := img["height"]; ok {
		slot["height"] = h
	}
	if th, ok := img["thumbhash"]; ok {
		slot["thumbhash"] = th
	}
	return slot
}

func spoilerLevel(s string) int {
	switch s {
	case "minimum":
		return 1
	case "spoilered", "major":
		return 2
	default:
		return 0
	}
}

func searchEntityType(object string) string {
	switch object {
	case "company":
		return "label"
	case "credit_name":
		return "name"
	default:
		return object
	}
}

package client

import (
	"encoding/json"
	"net/url"
	"testing"
)

// The v2 read lane is dormant behind catalogReadsV1, so nothing exercises these
// mappings until the flip. Every case below was captured from a live catalog
// response; the last cutover shipped green because it rewrote its assertions
// alongside the code, and the pages just went blank.

func TestV2CatalogQuery_DetailFacesAskForEveryBlock(t *testing.T) {
	for path, want := range map[string]string{
		"/catalog/works/4":                 v2CatalogDetailInclude["works"],
		"/catalog/characters/7":            v2CatalogDetailInclude["characters"],
		"/catalog/labels/1":                v2CatalogDetailInclude["companies"],
		"/catalog/names/1183":              v2CatalogDetailInclude["credit-names"],
		"/catalog/tags/41":                 v2CatalogDetailInclude["tags"],
		"/catalog/series/1":                v2CatalogDetailInclude["series"],
		"/catalog/labels/1/relation-graph": v2CatalogDetailInclude["companies"],
	} {
		if got := v2CatalogQuery(path, url.Values{}).Get("include"); got != want {
			t.Errorf("%s include = %q, want %q — v2 answers a bare id+name row without it", path, got, want)
		}
	}
}

func TestV2CatalogQuery_ListFacesKeepTheirOwnVocabulary(t *testing.T) {
	if got := v2CatalogQuery("/catalog/labels", url.Values{}).Get("include"); got != "aliases,logo" {
		t.Errorf("companies list include = %q, want aliases,logo (the v1 browse row)", got)
	}
	// The works list must keep the caller's include, remapped, not a full detail set.
	q := v2CatalogQuery("/catalog/works/search", url.Values{"include": {"names,covers,labels"}})
	if got := q.Get("include"); got != "titles,covers,companies" {
		t.Errorf("works search include = %q, want the v1 tokens remapped", got)
	}
}

func TestV2CatalogQuery_SpoilerCeiling(t *testing.T) {
	for _, tc := range []struct{ path, spoilers, want string }{
		{"/catalog/works/4", "2", "major"},
		{"/catalog/works/4", "1", "minor"},
		{"/catalog/works/4", "0", "none"},
		{"/catalog/characters/7", "2", "major"},
	} {
		q := v2CatalogQuery(tc.path, url.Values{"spoilers": {tc.spoilers}})
		if got := q.Get("spoiler"); got != tc.want {
			t.Errorf("%s spoilers=%s -> spoiler=%q, want %q", tc.path, tc.spoilers, got, tc.want)
		}
		if q.Has("spoilers") {
			t.Errorf("%s still sends v1's spoilers=, which v2 drops silently", tc.path)
		}
	}
	// A face with no spoiler axis must not grow the parameter.
	if q := v2CatalogQuery("/catalog/labels/1", url.Values{"spoilers": {"2"}}); q.Has("spoiler") {
		t.Error("companies detail grew a spoiler parameter")
	}
}

func TestV2CatalogQuery_BooleanFlags(t *testing.T) {
	// v1 took 1/0; v2 types these as real booleans and 400s on "1".
	if got := v2CatalogQuery("/catalog/tags", url.Values{"has_works": {"1"}}).Get("has_works"); got != "true" {
		t.Errorf("has_works = %q, want true — \"1\" is a 400", got)
	}
	if q := v2CatalogQuery("/catalog/tags", url.Values{"has_works": {"0"}}); q.Has("has_works") {
		t.Error("has_works=0 should drop the filter, not send false")
	}
}

func rewriteOne(t *testing.T, raw string) map[string]any {
	t.Helper()
	var got map[string]any
	if err := json.Unmarshal(rewriteV2JSON([]byte(raw), ""), &got); err != nil {
		t.Fatalf("rewrite produced invalid json: %v", err)
	}
	return got
}

func TestRewriteV2JSON_WorkTagCarriesCanonicalID(t *testing.T) {
	// catalog_detail.go drops a tag row whose canonical_id is 0, so v2 naming
	// the same column `id` empties the whole tag panel rather than degrading it.
	got := rewriteOne(t, `{"id":"638","display_name":"key","tag_kind":"meta","spoiler":"minor","is_sexual":false,"tier":"core","work_count":12}`)
	if got["canonical_id"] != float64(638) {
		t.Errorf("canonical_id = %v, want 638", got["canonical_id"])
	}
	if got["spoiler"] != float64(1) {
		t.Errorf("spoiler = %v, want 1 — v2's enum is none|minor|major", got["spoiler"])
	}
	if got["kind"] != "meta" || got["sexual"] != false {
		t.Errorf("tag_kind/is_sexual not folded: %v", got)
	}
	// A standalone tag list row has no spoiler and must not claim a canonical id.
	list := rewriteOne(t, `{"object":"tag","id":"638","display_name":"key","tag_kind":"meta","is_sexual":false}`)
	if _, has := list["canonical_id"]; has {
		t.Error("a tag browse row grew a canonical_id it never had on v1")
	}
}

func TestRewriteV2JSON_ImageGradingIsNumeric(t *testing.T) {
	// The cover block is typed int, so one "safe" made the WHOLE work detail
	// fail to decode rather than just that field.
	for wire, want := range map[string]float64{"safe": 0, "suggestive": 1, "explicit": 2} {
		got := rewriteOne(t, `{"id":"1","url":"u","sexual":"`+wire+`","violence":null}`)
		if got["sexual"] != want {
			t.Errorf("sexual %q -> %v, want %v", wire, got["sexual"], want)
		}
	}
	// A tag's boolean `sexual` must survive untouched.
	got := rewriteOne(t, `{"object":"tag","id":"1","is_sexual":true}`)
	if got["sexual"] != true {
		t.Errorf("tag sexual = %v, want true", got["sexual"])
	}
}

func TestRewriteV2JSON_StaffAndCompanyShapes(t *testing.T) {
	got := rewriteOne(t, `{"object":"credit_name","id":"1183","display_name":"高森奈津美","gender":"female","birth_year":1987,"birth_month":2,"birth_day":14,"photo":{"url":"u","hash":"ph"},"aliases":[{"value":"a","alias_kind":"alias"}]}`)
	if got["gender"] != float64(2) {
		t.Errorf("gender = %v, want 2 (male=1, female=2)", got["gender"])
	}
	if got["birth_y"] != float64(1987) || got["birth_m"] != float64(2) || got["birth_d"] != float64(14) {
		t.Errorf("birth_* not folded: %v", got)
	}
	if got["photo_hash"] != "ph" {
		t.Errorf("photo_hash = %v", got["photo_hash"])
	}
	alias := got["aliases"].([]any)[0].(map[string]any)
	if alias["kind"] != "alias" {
		t.Errorf("alias_kind not folded: %v", alias)
	}
	company := rewriteOne(t, `{"object":"company","id":"1","logo":{"url":"u","hash":"lg"}}`)
	if company["logo_hash"] != "lg" {
		t.Errorf("logo_hash = %v", company["logo_hash"])
	}
}

func TestRewriteV2JSON_TraitLieAndCreditCharacter(t *testing.T) {
	trait := rewriteOne(t, `{"id":"1","display_name":"金发","is_lie":true,"is_sexual":false,"spoiler":"major"}`)
	if trait["lie"] != true {
		t.Errorf("is_lie not folded: %v", trait)
	}
	if trait["spoiler"] != float64(2) {
		t.Errorf("spoiler = %v, want 2", trait["spoiler"])
	}
	role := rewriteOne(t, `{"role_key":"voice-actor","character_id":"7936","character_name":"空門蒼"}`)
	if role["character"] != "空門蒼" {
		t.Errorf("character_name not folded: %v", role)
	}
}

func TestOffsetCursorRoundTrip(t *testing.T) {
	// The sub-faces drop `offset` silently and page by an opaque cursor that is
	// the next offset in base64 — cur_Mw skips exactly three rows on a live
	// catalog, and a limit=50 page answers cur_NTA where v1 said next_offset=50.
	if got := encodeOffsetCursor(3); got != "cur_Mw" {
		t.Errorf("encodeOffsetCursor(3) = %q, want cur_Mw", got)
	}
	if got, ok := decodeOffsetCursor("cur_NTA"); !ok || got != 50 {
		t.Errorf("decodeOffsetCursor(cur_NTA) = %d,%v, want 50,true", got, ok)
	}
	for _, bad := range []string{"", "50", "cur_", "cur_!!!", "cur_" + "eyJzIjoiaWQifQ"} {
		if _, ok := decodeOffsetCursor(bad); ok {
			t.Errorf("decodeOffsetCursor(%q) claimed an offset", bad)
		}
	}
}

func TestIncludeHasToken(t *testing.T) {
	q := url.Values{"include": {"names, credits ,covers"}}
	if !includeHasToken(q, "credits") {
		t.Error("credits not found in a spaced include list")
	}
	if includeHasToken(q, "cred") {
		t.Error("a prefix matched a token")
	}
	if includeHasToken(url.Values{}, "credits") {
		t.Error("an absent include matched")
	}
}

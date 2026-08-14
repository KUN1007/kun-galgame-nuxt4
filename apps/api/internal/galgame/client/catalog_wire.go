package client

import (
	"strings"

	"kun-galgame-api/internal/galgame/dto"
)

type catClaimedBy struct {
	Site         string `json:"site"`
	WorkID       int    `json:"work_id"`
	State        string `json:"state"`
	ContentLimit string `json:"content_limit"`
}

type catRef struct {
	Source     string `json:"source"`
	ExternalID string `json:"external_id"`
}

type catRelatedLink struct {
	Source string `json:"source"`
	URL    string `json:"url"`
}

type catCoverSlot struct {
	URL       string `json:"url"`
	Width     int    `json:"width"`
	Height    int    `json:"height"`
	Thumbhash string `json:"thumbhash"`
	Sexual    int    `json:"sexual"`
	Violence  int    `json:"violence"`
	Source    string `json:"source"`
}

type catCoverSlots struct {
	Portrait *catCoverSlot `json:"portrait"`
	Banner   *catCoverSlot `json:"banner"`
}

type catNames struct {
	JaJP string `json:"ja-jp"`
	ZhCN string `json:"zh-cn"`
	ZhTW string `json:"zh-tw"`
	EnUS string `json:"en-us"`
}

type catIntroSlot struct {
	Intro   string `json:"intro"`
	Source  string `json:"source"`
	Machine bool   `json:"machine"`
}

type catIntros struct {
	JaJP *catIntroSlot `json:"ja-jp"`
	ZhCN *catIntroSlot `json:"zh-cn"`
	ZhTW *catIntroSlot `json:"zh-tw"`
	EnUS *catIntroSlot `json:"en-us"`
}

type catWorkLabel struct {
	ID          int64  `json:"id"`
	DisplayName string `json:"display_name"`
	LabelKind   string `json:"label_kind"`
	Kind        string `json:"kind"`
	Lang        string `json:"lang"`
	WorkCount   int    `json:"work_count"`
}

type catWorkEngine struct {
	ID        int64  `json:"id"`
	Name      string `json:"name"`
	WorkCount int    `json:"work_count"`
}

type catWorkLink struct {
	Source string `json:"source"`
	URL    string `json:"url"`
}

type catRating struct {
	Source    string  `json:"source"`
	Score     float64 `json:"score"`
	VoteCount int     `json:"vote_count"`
	Rank      *int    `json:"rank"`
}

type CatalogWorkListItem struct {
	ID            int64         `json:"id"`
	Medium        string        `json:"medium"`
	DisplayName   string        `json:"display_name"`
	ContentRating string        `json:"content_rating"`
	OLang         string        `json:"olang"`
	ReleaseDate   *string       `json:"release_date"`
	ClaimedBy     *catClaimedBy `json:"claimed_by"`
	Cover         string        `json:"cover"`
	Updated       string        `json:"updated"`

	Names   *catNames      `json:"names"`
	Intros  *catIntros     `json:"intros"`
	Labels  []catWorkLabel `json:"labels"`
	Ratings []catRating    `json:"ratings"`
	Covers  *catCoverSlots `json:"covers"`
	Refs    []catRef       `json:"refs"`

	ViaLabel *CatalogLabelVia `json:"via_label"`
}

type CatalogLabelVia struct {
	ID   int64  `json:"id"`
	Name string `json:"name"`
}

type catWorksListData struct {
	Items      []CatalogWorkListItem `json:"items"`
	NextCursor *string               `json:"next_cursor"`
	Count      int64                 `json:"count"`
	Month      string                `json:"month"`
	Year       string                `json:"year"`
	Meta       catCalendarMeta       `json:"meta"`
}

type catCalendarMeta struct {
	Today    string `json:"today"`
	MinMonth string `json:"min_month"`
	MaxMonth string `json:"max_month"`
	HasPrev  *bool  `json:"has_prev"`
	HasNext  *bool  `json:"has_next"`
}

type catWorksSearchData struct {
	Total int64                 `json:"total"`
	Page  int                   `json:"page"`
	Limit int                   `json:"limit"`
	Items []CatalogWorkListItem `json:"items"`
}

func hashFromURL(u string) string {
	if u == "" {
		return ""
	}
	base := u
	if i := strings.LastIndexByte(base, '/'); i >= 0 {
		base = base[i+1:]
	}
	return strings.TrimSuffix(base, ".webp")
}

const zeroTS = "0001-01-01T00:00:00Z"

func contentLimitOf(claimed *catClaimedBy, rating string) string {
	if claimed != nil {
		switch claimed.ContentLimit {
		case "sfw", "nsfw":
			return claimed.ContentLimit
		}
	}
	return contentLimitFromRating(rating)
}

func contentLimitFromRating(rating string) string {
	if rating == "r18" {
		return "nsfw"
	}
	return "sfw"
}

func ageLimitFromRating(rating string) string {
	if rating == "r18" {
		return "r18"
	}
	return "all"
}

func productLocale(olang string) string {
	tag := strings.ToLower(strings.TrimSpace(olang))
	switch {
	case tag == "":
		return ""
	case tag == "ja" || strings.HasPrefix(tag, "ja-"):
		return "ja-jp"
	case tag == "zh-hant" || strings.HasPrefix(tag, "zh-hant-") ||
		tag == "zh-tw" || tag == "zh-hk":
		return "zh-tw"
	case tag == "zh" || strings.HasPrefix(tag, "zh"):
		return "zh-cn"
	case tag == "en" || strings.HasPrefix(tag, "en-"):
		return "en-us"
	default:
		return olang
	}
}

const (
	claimStateLive    = "live"
	claimStateDraft   = "draft"
	claimStatePending = "pending"
	claimStateHidden  = "hidden"
)

const ClaimStateWizard = claimStateLive + "," + claimStateDraft + "," + claimStatePending

func statusFromClaimState(state string) int {
	if state == claimStateLive {
		return galgameStatusPublished
	}
	return galgameStatusVndbDraft
}

const (
	galgameStatusPublished = 0
	galgameStatusVndbDraft = 2
)

func (it *CatalogWorkListItem) isRenderable() bool {
	return it.ClaimedBy == nil || it.ClaimedBy.State != claimStateHidden
}

func CatalogItemRenderable(it *CatalogWorkListItem) bool { return it.isRenderable() }

func CatalogItemGID(it *CatalogWorkListItem) int { return it.gid() }

func (it *CatalogWorkListItem) gid() int {
	if it.ClaimedBy == nil || !isKungalClaim(it.ClaimedBy.Site) {
		return 0
	}
	return it.ClaimedBy.WorkID
}

const ClaimSiteKungal = "kungal"

const claimSiteLegacy = "galgame_wiki"

func isKungalClaim(site string) bool {
	return site == ClaimSiteKungal || site == claimSiteLegacy
}

func refsMap(refs []catRef) map[string]string {
	if refs == nil {
		return nil
	}
	out := make(map[string]string, len(refs))
	for _, r := range refs {
		prev, seen := out[r.Source]
		switch {
		case !seen:
			out[r.Source] = r.ExternalID
		case r.Source == sourceVNDB && !isVndbWorkID(prev) && isVndbWorkID(r.ExternalID):
			out[r.Source] = r.ExternalID
		}
	}
	return out
}

const sourceVNDB = "vndb"

func isVndbWorkID(s string) bool {
	if len(s) < 2 || s[0] != 'v' {
		return false
	}
	for i := 1; i < len(s); i++ {
		if s[i] < '0' || s[i] > '9' {
			return false
		}
	}
	return true
}

func coverFields(covers *catCoverSlots, fallbackURL string) (hash, url string, w, h int, thumb string) {
	slot := (*catCoverSlot)(nil)
	if covers != nil {
		if covers.Banner != nil {
			slot = covers.Banner
		} else if covers.Portrait != nil {
			slot = covers.Portrait
		}
	}
	if slot == nil {
		return hashFromURL(fallbackURL), fallbackURL, 0, 0, ""
	}
	return hashFromURL(slot.URL), slot.URL, slot.Width, slot.Height, slot.Thumbhash
}

func CatalogItemToBrief(it *CatalogWorkListItem) GalgameBrief {
	b := GalgameBrief{
		ID:                 it.gid(),
		NameEnUs:           it.name("en-us"),
		NameJaJp:           it.name("ja-jp"),
		NameZhCn:           it.name("zh-cn"),
		NameZhTw:           it.name("zh-tw"),
		AgeLimit:           ageLimitFromRating(it.ContentRating),
		ContentLimit:       contentLimitOf(it.ClaimedBy, it.ContentRating),
		OriginalLanguage:   productLocale(it.OLang),
		ReleaseDate:        it.ReleaseDate,
		ResourceUpdateTime: zeroTS,
		Refs:               refsMap(it.Refs),
	}
	if it.ClaimedBy != nil {
		b.Status = statusFromClaimState(it.ClaimedBy.State)
	} else {
		b.Status = galgameStatusVndbDraft
	}
	b.VndbID = b.Refs["vndb"]
	b.EffectiveBannerHash, b.EffectiveBannerURL,
		b.EffectiveBannerWidth, b.EffectiveBannerHeight,
		b.EffectiveBannerThumbhash = coverFields(it.Covers, it.Cover)
	return b
}

func (it *CatalogWorkListItem) name(key string) string {
	if it.Names == nil {
		if key == "ja-jp" {
			return it.DisplayName
		}
		return ""
	}
	switch key {
	case "ja-jp":
		return it.Names.JaJP
	case "zh-cn":
		return it.Names.ZhCN
	case "zh-tw":
		return it.Names.ZhTW
	case "en-us":
		return it.Names.EnUS
	}
	return ""
}

func (it *CatalogWorkListItem) intro(key string) string {
	if it.Intros == nil {
		return ""
	}
	slot := (*catIntroSlot)(nil)
	switch key {
	case "ja-jp":
		slot = it.Intros.JaJP
	case "zh-cn":
		slot = it.Intros.ZhCN
	case "zh-tw":
		slot = it.Intros.ZhTW
	case "en-us":
		slot = it.Intros.EnUS
	}
	if slot == nil {
		return ""
	}
	return slot.Intro
}

func CatalogItemToDetailBrief(it *CatalogWorkListItem) GalgameDetailBrief {
	b := GalgameDetailBrief{GalgameBrief: CatalogItemToBrief(it)}
	b.IntroEnUS = it.intro("en-us")
	b.IntroJaJP = it.intro("ja-jp")
	b.IntroZhCN = it.intro("zh-cn")
	b.IntroZhTW = it.intro("zh-tw")
	for _, l := range it.Labels {
		b.Officials = append(b.Officials, l.DisplayName)
	}
	return b
}

func CatalogItemToNextMoeItem(it *CatalogWorkListItem) dto.NextMoeGalgameItem {
	m := dto.NextMoeGalgameItem{
		ID:               it.gid(),
		NameEnUs:         it.name("en-us"),
		NameJaJp:         it.name("ja-jp"),
		NameZhCn:         it.name("zh-cn"),
		NameZhTw:         it.name("zh-tw"),
		ReleaseDate:      it.ReleaseDate,
		ContentLimit:     contentLimitOf(it.ClaimedBy, it.ContentRating),
		ReleasePrecision: releasePrecisionOf(it.ReleaseDate),
	}
	if it.ClaimedBy != nil {
		m.Status = statusFromClaimState(it.ClaimedBy.State)
	} else {
		m.Status = galgameStatusVndbDraft
	}
	if it.ViaLabel != nil {
		m.ViaOfficial = &dto.OfficialBrief{ID: int(it.ViaLabel.ID), Name: it.ViaLabel.Name}
	}
	m.EffectiveBannerHash, m.EffectiveBannerURL,
		m.EffectiveBannerWidth, m.EffectiveBannerHeight,
		m.EffectiveBannerThumbhash = coverFields(it.Covers, it.Cover)
	return m
}

func releasePrecisionOf(date *string) string {
	if date == nil {
		return "tba"
	}
	switch len(*date) {
	case 4:
		return "year"
	case 7:
		return "month"
	case 10:
		return "day"
	}
	return ""
}

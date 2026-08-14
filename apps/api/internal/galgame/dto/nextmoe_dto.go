package dto

type NextMoeAlias struct {
	Name string `json:"name"`
}

type NextMoeGalgameItem struct {
	ID                       int     `json:"id"`
	NameEnUs                 string  `json:"name_en_us"`
	NameJaJp                 string  `json:"name_ja_jp"`
	NameZhCn                 string  `json:"name_zh_cn"`
	NameZhTw                 string  `json:"name_zh_tw"`
	Banner                   string  `json:"banner"`
	ContentLimit             string  `json:"content_limit"`
	ResourceUpdateTime       string  `json:"resource_update_time"`
	ReleaseDate              *string `json:"release_date"`
	ReleaseDateTBA           bool    `json:"release_date_tba"`
	EffectiveBannerHash      string  `json:"effective_banner_hash"`
	EffectiveBannerURL       string  `json:"effective_banner_url"`
	EffectiveBannerWidth     int     `json:"effective_banner_width,omitempty"`
	EffectiveBannerHeight    int     `json:"effective_banner_height,omitempty"`
	EffectiveBannerThumbhash string  `json:"effective_banner_thumbhash,omitempty"`
	UserID                   int            `json:"user_id"`
	ReleasePrecision         string         `json:"release_precision"`
	Status                   int            `json:"status"`
	ViaOfficial              *OfficialBrief `json:"via_official,omitempty"`
}

type NextMoeGalgameCover struct {
	ImageHash string `json:"image_hash"`
	SortOrder int    `json:"sort_order"`
	Sexual    int    `json:"sexual"`
	Violence  int    `json:"violence"`
	Source    string `json:"source"`
	SourceKey string `json:"source_key"`
	Kind      string `json:"kind,omitempty"`
	CDNURL    string `json:"cdn_url"`
	Width     int    `json:"width,omitempty"`
	Height    int    `json:"height,omitempty"`
	Thumbhash string `json:"thumbhash,omitempty"`
}

type NextMoeGalgameScreenshot struct {
	ImageHash string `json:"image_hash"`
	SortOrder int    `json:"sort_order"`
	Caption   string `json:"caption"`
	Sexual    int    `json:"sexual"`
	Violence  int    `json:"violence"`
	Source    string `json:"source"`
	SourceKey string `json:"source_key"`
	CDNURL    string `json:"cdn_url"`
	Width     int    `json:"width,omitempty"`
	Height    int    `json:"height,omitempty"`
	Thumbhash string `json:"thumbhash,omitempty"`
}

type NextMoeOfficial struct {
	ID           int            `json:"id"`
	Name         string         `json:"name"`
	Link         string         `json:"link"`
	Category     string         `json:"category"`
	Roles        []string       `json:"roles"`
	Lang         string         `json:"lang"`
	Alias        []NextMoeAlias `json:"alias"`
	GalgameCount int            `json:"galgame_count"`
}

type NextMoeOfficialRel struct {
	Official NextMoeOfficial `json:"official"`
}

type NextMoeEngine struct {
	ID    int    `json:"id"`
	Name  string `json:"name"`
	Link  string `json:"link"`
	Intro string `json:"intro"`
}

type NextMoeEngineRel struct {
	Engine NextMoeEngine `json:"engine"`
}

type NextMoeTag struct {
	ID           int            `json:"id"`
	Name         string         `json:"name"`
	Category     string         `json:"category"`
	Alias        []NextMoeAlias `json:"alias"`
	GalgameCount int            `json:"galgame_count"`
}

type NextMoeTagRel struct {
	Tag NextMoeTag `json:"tag"`
}

type NextMoeContributor struct {
	UserID int `json:"user_id"`
}

type NextMoeGalgameDetail struct {
	ID                       int                  `json:"id"`
	NameEnUs                 string               `json:"name_en_us"`
	NameJaJp                 string               `json:"name_ja_jp"`
	NameZhCn                 string               `json:"name_zh_cn"`
	NameZhTw                 string               `json:"name_zh_tw"`
	Banner                   string               `json:"banner"`
	EffectiveBannerHash      string               `json:"effective_banner_hash"`
	EffectiveBannerURL       string               `json:"effective_banner_url"`
	EffectiveBannerWidth     int                  `json:"effective_banner_width,omitempty"`
	EffectiveBannerHeight    int                  `json:"effective_banner_height,omitempty"`
	EffectiveBannerThumbhash string               `json:"effective_banner_thumbhash,omitempty"`
	ContentLimit             string               `json:"content_limit"`
	AgeLimit                 string               `json:"age_limit"`
	OriginalLanguage         string               `json:"original_language"`
	SeriesID                 *int                 `json:"series_id"`
	Official                 []NextMoeOfficialRel `json:"official"`
	Engine                   []NextMoeEngineRel   `json:"engine"`
	Tag                      []NextMoeTagRel      `json:"tag"`
	Contributors             []NextMoeContributor `json:"contributors"`
}

type NextMoeGalgameDetailResponse struct {
	Galgame NextMoeGalgameDetail `json:"galgame"`
}

type NextMoeUser struct {
	ID     int    `json:"id"`
	Name   string `json:"name"`
	Avatar string `json:"avatar"`
}

type NextMoeTagWithSpoiler struct {
	SpoilerLevel int        `json:"spoiler_level"`
	Tag          NextMoeTag `json:"tag"`
}

type NextMoeEngineAlias []string

type NextMoeGalgameDetailFull struct {
	ID                         int                        `json:"id"`
	VndbID                     string                     `json:"vndb_id"`
	NameEnUs                   string                     `json:"name_en_us"`
	NameJaJp                   string                     `json:"name_ja_jp"`
	NameZhCn                   string                     `json:"name_zh_cn"`
	NameZhTw                   string                     `json:"name_zh_tw"`
	Banner                     string                     `json:"banner"`
	IntroEnUs                  string                     `json:"intro_en_us"`
	IntroJaJp                  string                     `json:"intro_ja_jp"`
	IntroZhCn                  string                     `json:"intro_zh_cn"`
	IntroZhTw                  string                     `json:"intro_zh_tw"`
	ContentLimit               string                     `json:"content_limit"`
	ResourceUpdateTime         string                     `json:"resource_update_time"`
	ReleaseDate                *string                    `json:"release_date"`
	ReleaseDateTBA             bool                       `json:"release_date_tba"`
	OriginalLanguage           string                     `json:"original_language"`
	AgeLimit                   string                     `json:"age_limit"`
	UserID                     int                        `json:"user_id"`
	SeriesID                   *int                       `json:"series_id"`
	Status                     int                        `json:"status"`
	EffectiveBannerHash        string                     `json:"effective_banner_hash"`
	EffectiveBannerURL         string                     `json:"effective_banner_url"`
	EffectiveBannerWidth       int                        `json:"effective_banner_width,omitempty"`
	EffectiveBannerHeight      int                        `json:"effective_banner_height,omitempty"`
	EffectiveBannerThumbhash   string                     `json:"effective_banner_thumbhash,omitempty"`
	EffectivePortraitHash      string                     `json:"effective_portrait_hash,omitempty"`
	EffectivePortraitURL       string                     `json:"effective_portrait_url,omitempty"`
	EffectivePortraitWidth     int                        `json:"effective_portrait_width,omitempty"`
	EffectivePortraitHeight    int                        `json:"effective_portrait_height,omitempty"`
	EffectivePortraitThumbhash string                     `json:"effective_portrait_thumbhash,omitempty"`
	ExternalRatings            []GalgameExternalRating    `json:"external_ratings"`
	Covers                     []NextMoeGalgameCover      `json:"covers"`
	Screenshots                []NextMoeGalgameScreenshot `json:"screenshots"`
	Alias                      []NextMoeAlias             `json:"alias"`
	Official                   []NextMoeOfficialRel       `json:"official"`
	Engine                     []NextMoeEngineWithAlias   `json:"engine"`
	Series                     []NextMoeSeriesRef         `json:"series"`
	Tag                        []NextMoeTagWithSpoiler    `json:"tag"`
	Staff                      []NextMoeStaffGroup        `json:"staff"`
	Characters                 []NextMoeGalgameCharacter  `json:"characters"`
	Contributor                []NextMoeContributor       `json:"contributor"`
	Created                    string                     `json:"created"`
	Updated                    string                     `json:"updated"`
	Refs                       map[string]string          `json:"refs,omitempty"`
}

type NextMoeStaffGroup struct {
	RoleKey  string             `json:"role_key"`
	RoleName string             `json:"role_name"`
	People   []NextMoeStaffName `json:"people"`
}

type NextMoeStaffName struct {
	ID         int      `json:"id"`
	Name       string   `json:"name"`
	Latin      string   `json:"latin,omitempty"`
	Characters []string `json:"characters,omitempty"`
}

type NextMoeGalgameCharacter struct {
	ID         int                     `json:"id"`
	Name       string                  `json:"name"`
	Latin      string                  `json:"latin,omitempty"`
	Kind       string                  `json:"kind"`
	Spoiler    int                     `json:"spoiler"`
	Image      string                  `json:"image,omitempty"`
	Figure     string                  `json:"figure,omitempty"`
	ImageMeta  *GalgameArtMeta         `json:"image_meta,omitempty"`
	FigureMeta *GalgameArtMeta         `json:"figure_meta,omitempty"`
	Voices     []NextMoeCharacterVoice `json:"voices"`
}

type NextMoeCharacterVoice struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
}

type NextMoeSeriesRef struct {
	ID   int    `json:"id"`
	Name string `json:"name"`
}

type NextMoeEngineWithAlias struct {
	Engine struct {
		ID           int      `json:"id"`
		Name         string   `json:"name"`
		Alias        []string `json:"alias"`
		GalgameCount int      `json:"galgame_count"`
	} `json:"engine"`
}

type NextMoeGalgameDetailFullResp struct {
	Galgame NextMoeGalgameDetailFull `json:"galgame"`
	Users   map[string]NextMoeUser   `json:"users"`
}

type NextMoeSeriesSample struct {
	NameEnUs                 string `json:"name_en_us"`
	NameJaJp                 string `json:"name_ja_jp"`
	NameZhCn                 string `json:"name_zh_cn"`
	NameZhTw                 string `json:"name_zh_tw"`
	Banner                   string `json:"banner"`
	ContentLimit             string `json:"content_limit"`
	EffectiveBannerHash      string `json:"effective_banner_hash"`
	EffectiveBannerURL       string `json:"effective_banner_url"`
	EffectiveBannerWidth     int    `json:"effective_banner_width,omitempty"`
	EffectiveBannerHeight    int    `json:"effective_banner_height,omitempty"`
	EffectiveBannerThumbhash string `json:"effective_banner_thumbhash,omitempty"`
}

type NextMoeSeriesBrief struct {
	ID          int                   `json:"id"`
	Name        string                `json:"name"`
	Description string                `json:"description"`
	Galgame     []NextMoeSeriesSample `json:"galgame"`
	Created     string                `json:"created"`
	Updated     string                `json:"updated"`
}

type NextMoeCreatedResp struct {
	ID int `json:"id"`
}

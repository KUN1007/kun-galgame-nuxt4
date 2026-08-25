package config

import (
	"fmt"
	"os"
	"strconv"
)

type Config struct {
	Server         ServerConfig
	Database       DatabaseConfig
	Redis          RedisConfig
	OAuth          OAuthConfig
	FileStorage    S3Config
	Search         SearchConfig
	CORS           CORSConfig
	NextMoeAPI     NextMoeAPIConfig
	NewsAPI        NewsAPIConfig
	ImageClient    ImageClientConfig
	ArtifactClient ArtifactClientConfig
	LinkChecker    LinkCheckerConfig
	Trust          TrustConfig
	Catalog        CatalogClientConfig
	Community      CommunityConfig
	Dlsite         DlsiteConfig
}

type CommunityConfig struct {
	BaseURL      string
	ClientID     string
	ClientSecret string
}

type CatalogClientConfig struct {
	BaseURL string
}

type TrustConfig struct {
	BaseURL        string
	CallbackSecret string
	Site           string
	CheckEnabled   bool
	ScanEnabled    bool
}

type ArtifactClientConfig struct {
	BaseURL      string
	ClientID     string
	ClientSecret string
}

type LinkCheckerConfig struct {
	BaseURL              string
	APIKey               string
	CFAccessClientID     string
	CFAccessClientSecret string
}

type ImageClientConfig struct {
	BaseURL      string
	ClientID     string
	ClientSecret string
}

type NextMoeAPIConfig struct {
	BaseURL      string
	APIKey       string
	ImageCDNBase string
}

// NewsAPIConfig is a SECOND NextMoe credential, not a copy of the first: the
// news face is gated on scope news:read, which the catalog key does not carry.
// An empty APIKey leaves /news answering 503 instead of failing startup — the
// forum's catalogue must not stop booting over a partner index.
type NewsAPIConfig struct {
	BaseURL string
	APIKey  string
}

// The affiliate link is assembled SERVER-side and shipped as a ready URL: the
// affiliate id stays out of the browser bundle, and this project's frontend
// build cannot be trusted with env vars (NUXT_PUBLIC_* / process.env.* come out
// undefined in the generic prod image), so a template baked into the frontend
// would silently produce broken links in production.
//
// LinkTemplate is a whole template, not assembled parts: DLsite's affiliate path
// differs per site segment, so a path change stays an env edit.
type DlsiteConfig struct {
	LinkTemplate string
	CouponURL    string
}

func (c DlsiteConfig) Configured() bool { return c.LinkTemplate != "" }

type ServerConfig struct {
	Port string
	Mode string
}

type DatabaseConfig struct {
	URL             string
	MaxOpenConns    int
	MaxIdleConns    int
	ConnMaxLifetime int
}

type RedisConfig struct {
	Host     string
	Port     string
	Password string
	DB       int
}

type OAuthConfig struct {
	ServerURL    string
	ClientID     string
	ClientSecret string
	RedirectURI  string
	JWTSecret    string
}

type S3Config struct {
	Endpoint  string
	Region    string
	Bucket    string
	AccessKey string
	SecretKey string
}

type SearchConfig struct {
	MeilisearchURL string
	MeilisearchKey string
}

type CORSConfig struct {
	AllowOrigins string
}

func Load() (*Config, error) {
	dbURL, err := requireEnv("KUN_DATABASE_URL")
	if err != nil {
		return nil, err
	}
	oauthServerURL, err := requireEnv("OAUTH_SERVER_URL")
	if err != nil {
		return nil, err
	}
	oauthClientID, err := requireEnv("OAUTH_CLIENT_ID")
	if err != nil {
		return nil, err
	}
	oauthClientSecret, err := requireEnv("OAUTH_CLIENT_SECRET")
	if err != nil {
		return nil, err
	}
	oauthRedirectURI, err := requireEnv("OAUTH_REDIRECT_URI")
	if err != nil {
		return nil, err
	}

	nextMoeBase := envOrDefault("KUN_NEXTMOE_API_BASE", "http://127.0.0.1:19281")
	nextMoeKey := envOrDefault("KUN_NEXTMOE_API_KEY", "")
	if nextMoeBase != "" && nextMoeKey == "" {
		return nil, fmt.Errorf(
			"KUN_NEXTMOE_API_KEY 未设置: catalog /v2 读面硬依赖 nmk_ developer API key; 已配置 KUN_NEXTMOE_API_BASE=%q 但 KUN_NEXTMOE_API_KEY 为空, 不做静默降级",
			nextMoeBase,
		)
	}

	return &Config{
		Server: ServerConfig{
			Port: envOrDefault("SERVER_PORT", "2334"),
			Mode: envOrDefault("SERVER_MODE", "dev"),
		},
		Database: DatabaseConfig{
			URL:             dbURL,
			MaxOpenConns:    envOrDefaultInt("DB_MAX_OPEN_CONNS", 25),
			MaxIdleConns:    envOrDefaultInt("DB_MAX_IDLE_CONNS", 10),
			ConnMaxLifetime: envOrDefaultInt("DB_CONN_MAX_LIFETIME", 300),
		},
		Redis: RedisConfig{
			Host:     envOrDefault("REDIS_HOST", "127.0.0.1"),
			Port:     envOrDefault("REDIS_PORT", "6379"),
			Password: envOrDefault("REDIS_PASSWORD", ""),
			DB:       envOrDefaultInt("REDIS_DB", 0),
		},
		OAuth: OAuthConfig{
			ServerURL:    oauthServerURL,
			ClientID:     oauthClientID,
			ClientSecret: oauthClientSecret,
			RedirectURI:  oauthRedirectURI,
			JWTSecret:    envOrDefault("JWT_SECRET", ""),
		},
		FileStorage: S3Config{
			Endpoint:  envOrDefault("FILE_STORAGE_ENDPOINT", ""),
			Region:    envOrDefault("FILE_STORAGE_REGION", ""),
			Bucket:    envOrDefault("FILE_STORAGE_BUCKET", ""),
			AccessKey: envOrDefault("FILE_STORAGE_ACCESS_KEY", ""),
			SecretKey: envOrDefault("FILE_STORAGE_SECRET_KEY", ""),
		},
		Search: SearchConfig{
			MeilisearchURL: envOrDefault("MEILISEARCH_URL", "http://127.0.0.1:7700"),
			MeilisearchKey: envOrDefault("MEILISEARCH_KEY", ""),
		},
		CORS: CORSConfig{
			AllowOrigins: envOrDefault(
				"CORS_ALLOW_ORIGINS",
				"http://127.0.0.1:2333,https://www.kungal.com",
			),
		},
		NextMoeAPI: NextMoeAPIConfig{
			BaseURL:      nextMoeBase,
			APIKey:       nextMoeKey,
			ImageCDNBase: envOrDefault("KUN_IMAGE_PUBLIC_BASE_URL", "https://image.kungal.iloveren.link"),
		},
		NewsAPI: NewsAPIConfig{
			BaseURL: envOrDefault("KUN_NEWS_API_BASE", nextMoeBase),
			APIKey:  envOrDefault("KUN_NEWS_API_KEY", ""),
		},
		ImageClient: ImageClientConfig{
			BaseURL:      envOrDefault("KUN_IMAGE_CLIENT_BASE_URL", "http://127.0.0.1:9278"),
			ClientID:     envOrDefault("KUN_IMAGE_CLIENT_ID", ""),
			ClientSecret: envOrDefault("KUN_IMAGE_CLIENT_SECRET", ""),
		},
		ArtifactClient: ArtifactClientConfig{
			BaseURL:      envOrDefault("KUN_ARTIFACT_CLIENT_BASE_URL", "http://127.0.0.1:9279"),
			ClientID:     envOrDefault("KUN_ARTIFACT_CLIENT_ID", ""),
			ClientSecret: envOrDefault("KUN_ARTIFACT_CLIENT_SECRET", ""),
		},
		LinkChecker: LinkCheckerConfig{
			BaseURL:              envOrDefault("LINK_CHECKER_BASE_URL", ""),
			APIKey:               envOrDefault("LINK_CHECKER_API_KEY", ""),
			CFAccessClientID:     envOrDefault("CF_ACCESS_CLIENT_ID", ""),
			CFAccessClientSecret: envOrDefault("CF_ACCESS_CLIENT_SECRET", ""),
		},
		Trust: TrustConfig{
			BaseURL:        envOrDefault("KUN_TRUST_BASE_URL", "http://127.0.0.1:9283"),
			CallbackSecret: envOrDefault("KUN_TRUST_CALLBACK_SECRET", ""),
			Site:           envOrDefault("KUN_TRUST_SITE", "kungal"),
			CheckEnabled:   envOrDefaultBool("KUN_TRUST_CHECK_ENABLED", false),
			ScanEnabled:    envOrDefaultBool("KUN_TRUST_SCAN_ENABLED", false),
		},
		Catalog: CatalogClientConfig{
			BaseURL: envOrDefault("KUN_CATALOG_API_BASE", "http://127.0.0.1:19281"),
		},
		Community: CommunityConfig{
			BaseURL:      envOrDefault("KUN_COMMUNITY_API_BASE", ""),
			ClientID:     envOrDefault("KUN_COMMUNITY_CLIENT_ID", ""),
			ClientSecret: envOrDefault("KUN_COMMUNITY_CLIENT_SECRET", ""),
		},
		Dlsite: DlsiteConfig{
			LinkTemplate: envOrDefault("KUN_DLSITE_LINK_TEMPLATE", ""),
			CouponURL:    envOrDefault("KUN_DLSITE_COUPON_URL", ""),
		},
	}, nil
}

func requireEnv(key string) (string, error) {
	val := os.Getenv(key)
	if val == "" {
		return "", fmt.Errorf("环境变量 %s 未设置", key)
	}
	return val, nil
}

func envOrDefault(key, fallback string) string {
	if val := os.Getenv(key); val != "" {
		return val
	}
	return fallback
}

func envOrDefaultInt(key string, fallback int) int {
	if val := os.Getenv(key); val != "" {
		if n, err := strconv.Atoi(val); err == nil {
			return n
		}
	}
	return fallback
}

func envOrDefaultBool(key string, fallback bool) bool {
	if val := os.Getenv(key); val != "" {
		if b, err := strconv.ParseBool(val); err == nil {
			return b
		}
	}
	return fallback
}

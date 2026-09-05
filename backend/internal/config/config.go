package config

import (
	"os"
	"strconv"
	"strings"
	"time"
)

// Config is the single source of truth for runtime configuration.
// Everything comes from the environment so the same binary runs as API or worker.
type Config struct {
	AppEnv       string
	HTTPAddr     string
	PublicAppURL string
	PublicAPIURL string

	SessionCookieName string
	SessionTTL        time.Duration

	DatabaseURL string
	RedisURL    string

	MinioEndpoint       string
	MinioPublicEndpoint string
	MinioAccessKey      string
	MinioSecretKey      string
	MinioBucket         string
	MinioUseSSL         bool

	AllowDevLogin      bool
	GoogleClientID     string
	GoogleClientSecret string
	GoogleRedirectURL  string

	AIBaseURL     string
	AIAPIKey      string
	AIModel       string
	AITimeout     time.Duration
	AIMaxAttempts int

	ImageProvider     string
	UnsplashAccessKey string
	WorkerConcurrency int
}

func Load() Config {
	return Config{
		AppEnv:       env("APP_ENV", "development"),
		HTTPAddr:     env("HTTP_ADDR", ":8080"),
		PublicAppURL: env("PUBLIC_APP_URL", "http://localhost:3000"),
		PublicAPIURL: env("PUBLIC_API_URL", "http://localhost:8080"),

		SessionCookieName: env("SESSION_COOKIE_NAME", "tks_session"),
		SessionTTL:        time.Duration(envInt("SESSION_TTL_HOURS", 720)) * time.Hour,

		DatabaseURL: env("DATABASE_URL", "postgres://tks:tks@localhost:5432/tks?sslmode=disable"),
		RedisURL:    env("REDIS_URL", "redis://localhost:6379/0"),

		MinioEndpoint:       env("MINIO_ENDPOINT", "localhost:9100"),
		MinioPublicEndpoint: env("MINIO_PUBLIC_ENDPOINT", env("MINIO_ENDPOINT", "localhost:9100")),
		MinioAccessKey:      env("MINIO_ACCESS_KEY", "minioadmin"),
		MinioSecretKey:      env("MINIO_SECRET_KEY", "minioadmin"),
		MinioBucket:         env("MINIO_BUCKET", "tks"),
		MinioUseSSL:         envBool("MINIO_USE_SSL", false),

		AllowDevLogin:      envBool("ALLOW_DEV_LOGIN", true),
		GoogleClientID:     env("GOOGLE_CLIENT_ID", ""),
		GoogleClientSecret: env("GOOGLE_CLIENT_SECRET", ""),
		GoogleRedirectURL:  env("GOOGLE_REDIRECT_URL", "http://localhost:8080/api/v1/auth/callback"),

		AIBaseURL:     strings.TrimRight(env("AI_BASE_URL", "https://api.openai.com/v1"), "/"),
		AIAPIKey:      env("AI_API_KEY", ""),
		AIModel:       env("AI_MODEL", "gpt-4o-mini"),
		AITimeout:     time.Duration(envInt("AI_TIMEOUT_SECONDS", 90)) * time.Second,
		AIMaxAttempts: envInt("AI_MAX_ATTEMPTS", 3),

		ImageProvider:     env("IMAGE_PROVIDER", "unsplash"),
		UnsplashAccessKey: env("UNSPLASH_ACCESS_KEY", ""),
		WorkerConcurrency: envInt("WORKER_CONCURRENCY", 2),
	}
}

// GoogleEnabled reports whether real Google OAuth is configured.
func (c Config) GoogleEnabled() bool { return c.GoogleClientID != "" && c.GoogleClientSecret != "" }

func env(k, def string) string {
	if v := os.Getenv(k); v != "" {
		return v
	}
	return def
}

func envInt(k string, def int) int {
	if v, err := strconv.Atoi(os.Getenv(k)); err == nil {
		return v
	}
	return def
}

func envBool(k string, def bool) bool {
	if v, err := strconv.ParseBool(os.Getenv(k)); err == nil {
		return v
	}
	return def
}

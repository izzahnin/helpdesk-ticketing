package config

import (
	"log"
	"net/url"
	"os"
	"strings"
)

type Config struct {
	AppEnv             string
	Port               string
	DatabaseURL        string
	JWTAccessSecret    string
	RefreshTokenPepper string
	FrontendURL        string
	AuthDevTokenBody   bool
}

func Load() Config {
	cfg := Config{
		AppEnv:             getEnv("APP_ENV", "development"),
		Port:               getEnv("PORT", "4000"),
		DatabaseURL:        databaseURL(os.Getenv("DATABASE_URL"), getEnv("APP_ENV", "development")),
		JWTAccessSecret:    os.Getenv("JWT_ACCESS_SECRET"),
		RefreshTokenPepper: os.Getenv("REFRESH_TOKEN_PEPPER"),
		FrontendURL:        getEnv("FRONTEND_URL", "http://localhost:3000"),
		AuthDevTokenBody:   os.Getenv("AUTH_DEV_TOKEN_BODY") == "true",
	}
	if cfg.AppEnv == "production" {
		require("DATABASE_URL", cfg.DatabaseURL)
		require("JWT_ACCESS_SECRET", cfg.JWTAccessSecret)
		require("REFRESH_TOKEN_PEPPER", cfg.RefreshTokenPepper)
		require("FRONTEND_URL", cfg.FrontendURL)
	}
	return cfg
}

func databaseURL(raw, appEnv string) string {
	if strings.EqualFold(appEnv, "production") || raw == "" {
		return raw
	}

	parsed, err := url.Parse(raw)
	if err == nil && isPostgresURL(parsed) && isLocalHost(parsed.Hostname()) {
		query := parsed.Query()
		query.Set("sslmode", "disable")
		parsed.RawQuery = query.Encode()
		return parsed.String()
	}

	if isPostgresRawURL(raw) && !strings.Contains(raw, "sslmode=") {
		separator := "?"
		if strings.Contains(raw, "?") {
			separator = "&"
		}
		return raw + separator + "sslmode=disable"
	}
	return raw
}

func isPostgresURL(parsed *url.URL) bool {
	return parsed.Scheme == "postgres" || parsed.Scheme == "postgresql"
}

func isPostgresRawURL(raw string) bool {
	return strings.HasPrefix(raw, "postgres://") || strings.HasPrefix(raw, "postgresql://")
}

func isLocalHost(host string) bool {
	return host == "localhost" || host == "127.0.0.1" || host == "::1"
}

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

func require(key, value string) {
	if value == "" {
		log.Fatalf("missing required production env var: %s", key)
	}
}

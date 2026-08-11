package config

import (
	"log"
	"os"
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
		DatabaseURL:        os.Getenv("DATABASE_URL"),
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

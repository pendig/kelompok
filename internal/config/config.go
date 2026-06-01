package config

import (
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"
)

type Config struct {
	Env                       string
	APIAddr                   string
	AdminAPIKey               string
	AdminOrganizationSlugs    []string
	DatabaseURL               string
	DatabaseMaxConns          int32
	DatabaseMinConns          int32
	DatabaseMaxConnLifetime   time.Duration
	DatabaseMaxConnIdleTime   time.Duration
	DatabaseHealthCheckPeriod time.Duration
	GoogleOAuthClientID       string
	GoogleOAuthClientSecret   string
}

func Load() (Config, error) {
	if err := loadDotEnv(".env"); err != nil {
		return Config{}, err
	}

	maxConns, err := envInt32("KELOMPOK_DB_MAX_CONNS", 5)
	if err != nil {
		return Config{}, err
	}

	minConns, err := envInt32("KELOMPOK_DB_MIN_CONNS", 0)
	if err != nil {
		return Config{}, err
	}

	maxConnLifetime, err := envDuration("KELOMPOK_DB_MAX_CONN_LIFETIME", 30*time.Minute)
	if err != nil {
		return Config{}, err
	}

	maxConnIdleTime, err := envDuration("KELOMPOK_DB_MAX_CONN_IDLE_TIME", 5*time.Minute)
	if err != nil {
		return Config{}, err
	}

	healthCheckPeriod, err := envDuration("KELOMPOK_DB_HEALTH_CHECK_PERIOD", time.Minute)
	if err != nil {
		return Config{}, err
	}

	return Config{
		Env:                       env("KELOMPOK_ENV", "development"),
		APIAddr:                   env("KELOMPOK_API_ADDR", ":4621"),
		AdminAPIKey:               strings.TrimSpace(os.Getenv("KELOMPOK_ADMIN_API_KEY")),
		AdminOrganizationSlugs:    envCSV("KELOMPOK_ADMIN_ORGANIZATION_SLUGS"),
		DatabaseURL:               os.Getenv("KELOMPOK_DATABASE_URL"),
		DatabaseMaxConns:          maxConns,
		DatabaseMinConns:          minConns,
		DatabaseMaxConnLifetime:   maxConnLifetime,
		DatabaseMaxConnIdleTime:   maxConnIdleTime,
		DatabaseHealthCheckPeriod: healthCheckPeriod,
		GoogleOAuthClientID:       os.Getenv("GOOGLE_OAUTH_CLIENT_ID"),
		GoogleOAuthClientSecret:   os.Getenv("GOOGLE_OAUTH_CLIENT_SECRET"),
	}, nil
}

func envCSV(key string) []string {
	value := strings.TrimSpace(os.Getenv(key))
	if value == "" {
		return nil
	}

	parts := strings.Split(value, ",")
	items := make([]string, 0, len(parts))
	for _, part := range parts {
		item := strings.TrimSpace(strings.ToLower(part))
		if item != "" {
			items = append(items, item)
		}
	}

	return items
}

func env(key, fallback string) string {
	value := os.Getenv(key)
	if value == "" {
		return fallback
	}
	return value
}

func envInt32(key string, fallback int32) (int32, error) {
	value := os.Getenv(key)
	if value == "" {
		return fallback, nil
	}

	parsed, err := strconv.ParseInt(value, 10, 32)
	if err != nil {
		return 0, fmt.Errorf("parse %s: %w", key, err)
	}
	if parsed < 0 {
		return 0, fmt.Errorf("parse %s: must be non-negative", key)
	}

	return int32(parsed), nil
}

func envDuration(key string, fallback time.Duration) (time.Duration, error) {
	value := os.Getenv(key)
	if value == "" {
		return fallback, nil
	}

	parsed, err := time.ParseDuration(value)
	if err != nil {
		return 0, fmt.Errorf("parse %s: %w", key, err)
	}
	if parsed < 0 {
		return 0, fmt.Errorf("parse %s: must be non-negative", key)
	}

	return parsed, nil
}

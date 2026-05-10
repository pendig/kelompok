package config

import "os"

type Config struct {
	Env         string
	APIAddr     string
	DatabaseURL string
}

func Load() Config {
	_ = loadDotEnv(".env")

	return Config{
		Env:         env("KELOMPOK_ENV", "development"),
		APIAddr:     env("KELOMPOK_API_ADDR", ":4621"),
		DatabaseURL: os.Getenv("KELOMPOK_DATABASE_URL"),
	}
}

func env(key, fallback string) string {
	value := os.Getenv(key)
	if value == "" {
		return fallback
	}
	return value
}

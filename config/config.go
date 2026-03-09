package config

import (
	"fmt"
	"os"
)

type Config struct {
	Port      string
	Env       string
	DSN       string
	JWTSecret string
}

func Load() *Config {
	return &Config{
		Port:      getEnv("APP_PORT", "8080"),
		Env:       getEnv("APP_ENV", "development"),
		JWTSecret: getEnv("JWT_SECRET", "change-me-in-production"),
		DSN: fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?parseTime=true&charset=utf8mb4",
			getEnv("DB_USER", "root"),
			getEnv("DB_PASSWORD", ""),
			getEnv("DB_HOST", "localhost"),
			getEnv("DB_PORT", "3306"),
			getEnv("DB_NAME", "foodcourt"),
		),
	}
}

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
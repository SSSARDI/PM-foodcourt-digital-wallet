package config

import (
	"fmt"
	"os"
)

type Config struct {
	Port string
	Env  string
	DSN  string
}

func Load() *Config {
	return &Config{
		Port: getEnv("APP_PORT", "8080"),
		Env:  getEnv("APP_ENV", "development"),
		DSN: fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?parseTime=true&charset=utf8mb4",
			getEnv("DB_USER", "root"),
			getEnv("DB_PASSWORD", ""),
			getEnv("DB_HOST", "localhost"),
			getEnv("DB_PORT", "3306"),
			getEnv("DB_NAME", "myapp"),
		),
	}
}

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

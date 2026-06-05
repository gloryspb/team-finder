package config

import (
	"os"
	"strconv"
	"time"
)

type Config struct {
	AppEnv       string
	ServerPort   string
	DatabaseURL  string
	RedisAddr    string
	JWTSecret    string
	TokenTTL     time.Duration
	AllowOrigins string
}

func Load() Config {
	return Config{
		AppEnv:       getEnv("APP_ENV", "development"),
		ServerPort:   getEnv("SERVER_PORT", getEnv("PORT", "8080")),
		DatabaseURL:  getEnv("DATABASE_URL", "postgres://teamfinder:teamfinder@postgres:5432/teamfinder?sslmode=disable"),
		RedisAddr:    getEnv("REDIS_ADDR", "redis:6379"),
		JWTSecret:    getEnv("JWT_SECRET", "change-me-in-production"),
		TokenTTL:     time.Duration(getEnvInt("JWT_TTL_HOURS", 24)) * time.Hour,
		AllowOrigins: getEnv("ALLOW_ORIGINS", "http://localhost:5173,http://localhost:3000"),
	}
}

func getEnv(key, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return fallback
}

func getEnvInt(key string, fallback int) int {
	value := os.Getenv(key)
	if value == "" {
		return fallback
	}
	parsed, err := strconv.Atoi(value)
	if err != nil {
		return fallback
	}
	return parsed
}

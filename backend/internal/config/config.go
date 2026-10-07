package config

import (
	"fmt"
	"os"
	"strings"
	"time"
)

type Config struct {
	Port          string
	DatabaseURL   string
	RedisURL      string
	JWTSecret     string
	JWTTTL        time.Duration
	AllowedOrigin string
}

func Load() (*Config, error) {
	c := &Config{
		Port:          get("PORT", "8080"),
		DatabaseURL:   get("DATABASE_URL", "postgres://cloudpulse:cloudpulse@localhost:5432/cloudpulse?sslmode=disable"),
		RedisURL:      get("REDIS_URL", "redis://localhost:6379"),
		JWTSecret:     os.Getenv("JWT_SECRET"),
		JWTTTL:        24 * time.Hour,
		AllowedOrigin: get("ALLOWED_ORIGIN", "http://localhost:5173"),
	}
	if len(c.JWTSecret) < 32 {
		return nil, fmt.Errorf("JWT_SECRET must be set and at least 32 characters long")
	}
	return c, nil
}

func get(key, fallback string) string {
	if v := strings.TrimSpace(os.Getenv(key)); v != "" {
		return v
	}
	return fallback
}
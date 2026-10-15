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

	// Lets health checks target localhost and private networks.
	// For local testing only. Keep it off on a real server.
	AllowPrivateTargets bool

	// Where alert notifications go. Each is optional; an unset channel just logs a warning.
	SlackWebhookURL string
	AlertWebhookURL string
	SMTPHost        string
	SMTPPort        string
	SMTPUser        string
	SMTPPass        string
	SMTPFrom        string
}

func Load() (*Config, error) {
	c := &Config{
		Port:                get("PORT", "8080"),
		DatabaseURL:         get("DATABASE_URL", "postgres://cloudpulse:cloudpulse@localhost:5433/cloudpulse?sslmode=disable"),
		RedisURL:            get("REDIS_URL", "redis://localhost:6379"),
		JWTSecret:           os.Getenv("JWT_SECRET"),
		JWTTTL:              24 * time.Hour,
		AllowedOrigin:       get("ALLOWED_ORIGIN", "http://localhost:5173"),
		AllowPrivateTargets: get("ALLOW_PRIVATE_TARGETS", "false") == "true",

		SlackWebhookURL: get("SLACK_WEBHOOK_URL", ""),
		AlertWebhookURL: get("ALERT_WEBHOOK_URL", ""),
		SMTPHost:        get("SMTP_HOST", ""),
		SMTPPort:        get("SMTP_PORT", "587"),
		SMTPUser:        get("SMTP_USER", ""),
		SMTPPass:        get("SMTP_PASS", ""),
		SMTPFrom:        get("SMTP_FROM", ""),
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

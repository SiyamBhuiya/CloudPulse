package service

import (
	"context"
	"net"
	"net/url"
	"strings"
	"unicode/utf8"

	"cloudpulse/internal/healthcheck"
	"cloudpulse/internal/models"
	"cloudpulse/internal/repository"
)

const maxChecksPerUser = 50

// ValidationError is a problem with what the user typed. The handler returns it as a 400.
type ValidationError struct{ Msg string }

func (e ValidationError) Error() string { return e.Msg }

type HealthService struct {
	repo         *repository.HealthRepo
	allowPrivate bool
}

func NewHealthService(repo *repository.HealthRepo, allowPrivate bool) *HealthService {
	return &HealthService{repo: repo, allowPrivate: allowPrivate}
}

func (s *HealthService) List(ctx context.Context, userID string) ([]models.HealthCheck, error) {
	return s.repo.ListByUser(ctx, userID)
}

func (s *HealthService) Create(ctx context.Context, userID, name, rawURL string, intervalSec, expected int) (*models.HealthCheck, error) {
	name = strings.TrimSpace(name)
	rawURL = strings.TrimSpace(rawURL)

	switch {
	case name == "":
		return nil, ValidationError{"name is required"}
	case utf8.RuneCountInString(name) > 60:
		return nil, ValidationError{"name must be at most 60 characters"}
	case intervalSec != 30 && intervalSec != 60 && intervalSec != 300:
		return nil, ValidationError{"interval must be 30, 60 or 300 seconds"}
	case expected < 100 || expected > 599:
		return nil, ValidationError{"expected status must be between 100 and 599"}
	}
	if err := s.checkURL(rawURL); err != nil {
		return nil, err
	}

	n, err := s.repo.CountByUser(ctx, userID)
	if err != nil {
		return nil, err
	}
	if n >= maxChecksPerUser {
		return nil, ValidationError{"you have reached the limit of 50 health checks"}
	}
	return s.repo.Create(ctx, userID, name, rawURL, intervalSec, expected)
}

// checkURL gives early feedback. The checker enforces the same rule again at
// connection time, which is the part that cannot be bypassed.
func (s *HealthService) checkURL(raw string) error {
	if raw == "" || len(raw) > 500 {
		return ValidationError{"enter a URL of up to 500 characters"}
	}
	u, err := url.Parse(raw)
	if err != nil || (u.Scheme != "http" && u.Scheme != "https") || u.Hostname() == "" {
		return ValidationError{"enter a full URL that starts with http:// or https://"}
	}
	if u.User != nil {
		return ValidationError{"the URL cannot contain a username or password"}
	}
	if s.allowPrivate {
		return nil
	}
	host := strings.ToLower(u.Hostname())
	if host == "localhost" || strings.HasSuffix(host, ".localhost") {
		return ValidationError{"this address is not allowed"}
	}
	if ip := net.ParseIP(host); ip != nil && healthcheck.BlockedIP(ip) {
		return ValidationError{"this address is not allowed"}
	}
	return nil
}

func (s *HealthService) SetEnabled(ctx context.Context, userID, id string, enabled bool) error {
	return s.repo.SetEnabled(ctx, userID, id, enabled)
}

func (s *HealthService) Delete(ctx context.Context, userID, id string) error {
	return s.repo.Delete(ctx, userID, id)
}

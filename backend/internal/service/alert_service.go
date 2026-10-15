package service

import (
	"context"
	"errors"
	"strings"
	"unicode/utf8"

	"cloudpulse/internal/models"
	"cloudpulse/internal/repository"
)

const maxRulesPerUser = 50

var (
	validMetrics  = map[string]bool{"cpu": true, "memory": true, "disk": true}
	validChannels = map[string]bool{"email": true, "slack": true, "webhook": true}
)

type AlertService struct {
	repo    *repository.AlertRepo
	servers *repository.ServerRepo
}

func NewAlertService(repo *repository.AlertRepo, servers *repository.ServerRepo) *AlertService {
	return &AlertService{repo: repo, servers: servers}
}

// validate cleans the input and checks that the chosen server belongs to this user.
func (s *AlertService) validate(ctx context.Context, userID string, r *models.AlertRule) error {
	r.Name = strings.TrimSpace(r.Name)
	r.ServerID = strings.TrimSpace(r.ServerID)
	if r.ServerID == "" {
		r.ServerID = "all"
	}

	switch {
	case r.Name == "":
		return ValidationError{"name is required"}
	case utf8.RuneCountInString(r.Name) > 60:
		return ValidationError{"name must be at most 60 characters"}
	case !validMetrics[r.Metric]:
		return ValidationError{"metric must be cpu, memory or disk"}
	case !(r.Threshold > 0 && r.Threshold <= 100):
		return ValidationError{"threshold must be between 1 and 100"}
	case r.DurationMin < 1 || r.DurationMin > 120:
		return ValidationError{"duration must be between 1 and 120 minutes"}
	case !validChannels[r.Channel]:
		return ValidationError{"channel must be email, slack or webhook"}
	}

	if r.ServerID != "all" {
		if _, err := s.servers.Get(ctx, userID, r.ServerID); err != nil {
			if errors.Is(err, repository.ErrNotFound) {
				return ValidationError{"server not found"}
			}
			return err
		}
	}
	return nil
}

func (s *AlertService) ListRules(ctx context.Context, userID string) ([]models.AlertRule, error) {
	return s.repo.ListRules(ctx, userID)
}

func (s *AlertService) CreateRule(ctx context.Context, userID string, in models.AlertRule) (*models.AlertRule, error) {
	if err := s.validate(ctx, userID, &in); err != nil {
		return nil, err
	}
	n, err := s.repo.CountRules(ctx, userID)
	if err != nil {
		return nil, err
	}
	if n >= maxRulesPerUser {
		return nil, ValidationError{"you have reached the limit of 50 alert rules"}
	}
	in.Enabled = true
	return s.repo.CreateRule(ctx, userID, in)
}

func (s *AlertService) UpdateRule(ctx context.Context, userID, id string, in models.AlertRule) (*models.AlertRule, error) {
	if err := s.validate(ctx, userID, &in); err != nil {
		return nil, err
	}
	return s.repo.UpdateRule(ctx, userID, id, in)
}

func (s *AlertService) DeleteRule(ctx context.Context, userID, id string) error {
	return s.repo.DeleteRule(ctx, userID, id)
}

func (s *AlertService) History(ctx context.Context, userID string) ([]models.AlertEvent, error) {
	return s.repo.History(ctx, userID)
}

func (s *AlertService) ServerEvents(ctx context.Context, userID, serverID string) ([]models.ServerEvent, error) {
	if _, err := s.servers.Get(ctx, userID, serverID); err != nil {
		return nil, err
	}
	return s.repo.ServerEvents(ctx, userID, serverID)
}

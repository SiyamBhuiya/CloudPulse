package service

import (
	"context"
	"errors"
	"time"

	"cloudpulse/internal/models"
	"cloudpulse/internal/repository"
)

var ErrBadRange = errors.New("range must be 1h, 6h, 24h or 7d")

// Each range is cut into 72 buckets.
var ranges = map[string]struct {
	span      time.Duration
	bucketSec int
}{
	"1h":  {1 * time.Hour, 50},
	"6h":  {6 * time.Hour, 300},
	"24h": {24 * time.Hour, 1200},
	"7d":  {7 * 24 * time.Hour, 8400},
}

type MetricService struct {
	servers *repository.ServerRepo
	metrics *repository.MetricRepo
}

func NewMetricService(servers *repository.ServerRepo, metrics *repository.MetricRepo) *MetricService {
	return &MetricService{servers: servers, metrics: metrics}
}

// Both methods first check that the server belongs to this user.

func (s *MetricService) History(ctx context.Context, userID, serverID, rng string) ([]models.HistoryPoint, error) {
	cfg, ok := ranges[rng]
	if !ok {
		return nil, ErrBadRange
	}
	if _, err := s.servers.Get(ctx, userID, serverID); err != nil {
		return nil, err
	}
	return s.metrics.History(ctx, serverID, time.Now().Add(-cfg.span), cfg.bucketSec)
}

func (s *MetricService) Processes(ctx context.Context, userID, serverID string) ([]models.ProcessInfo, error) {
	if _, err := s.servers.Get(ctx, userID, serverID); err != nil {
		return nil, err
	}
	return s.metrics.Processes(ctx, serverID)
}

// Exists lets the events endpoint check ownership too.
func (s *MetricService) Exists(ctx context.Context, userID, serverID string) error {
	_, err := s.servers.Get(ctx, userID, serverID)
	return err
}

package service

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/hex"
	"time"

	"cloudpulse/internal/models"
	"cloudpulse/internal/repository"
)

// A server counts as online if its agent reported within this window.
const onlineWindow = 30 * time.Second

type ServerService struct {
	servers *repository.ServerRepo
}

func NewServerService(servers *repository.ServerRepo) *ServerService {
	return &ServerService{servers: servers}
}

// HashAgentKey is what gets stored. The agent sends the real key, and the
// collector will hash it the same way to find the server.
func HashAgentKey(key string) string {
	sum := sha256.Sum256([]byte(key))
	return hex.EncodeToString(sum[:])
}

func newAgentKey() (string, error) {
	b := make([]byte, 24)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return "cp_live_" + hex.EncodeToString(b), nil
}

// decorate fills in the fields the dashboard shows.
// CPU, memory and uptime stay at placeholder values until agents report real metrics.
func decorate(s *models.Server) {
	seen := s.CreatedAt
	if s.LastSeenAt != nil {
		seen = *s.LastSeenAt
	}
	s.LastSeen = seen.UnixMilli()

	if s.LastSeenAt != nil && time.Since(*s.LastSeenAt) < onlineWindow {
		s.Status = "online"
	} else {
		s.Status = "down"
	}
	s.Uptime = 100
}

// Create returns the new server and the plain agent key. The key is only
// ever available here, because the database keeps just its hash.
func (s *ServerService) Create(ctx context.Context, userID, name, region string) (*models.Server, string, error) {
	key, err := newAgentKey()
	if err != nil {
		return nil, "", err
	}
	server, err := s.servers.Create(ctx, userID, name, region, HashAgentKey(key))
	if err != nil {
		return nil, "", err
	}
	decorate(server)
	return server, key, nil
}

func (s *ServerService) List(ctx context.Context, userID string) ([]models.Server, error) {
	list, err := s.servers.ListByUser(ctx, userID)
	if err != nil {
		return nil, err
	}
	for i := range list {
		decorate(&list[i])
	}
	return list, nil
}

func (s *ServerService) Get(ctx context.Context, userID, id string) (*models.Server, error) {
	server, err := s.servers.Get(ctx, userID, id)
	if err != nil {
		return nil, err
	}
	decorate(server)
	return server, nil
}

func (s *ServerService) Delete(ctx context.Context, userID, id string) error {
	return s.servers.Delete(ctx, userID, id)
}
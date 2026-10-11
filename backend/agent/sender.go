package agent

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"cloudpulse/agent/collector"
)

var ErrUnauthorized = errors.New("the API rejected the agent key")

type Sender struct {
	url    string
	key    string
	client *http.Client
}

func NewSender(apiURL, key string) *Sender {
	return &Sender{
		url:    strings.TrimRight(apiURL, "/") + "/api/agent/metrics",
		key:    key,
		client: &http.Client{Timeout: 5 * time.Second},
	}
}

func (s *Sender) Send(ctx context.Context, sample collector.Sample) error {
	body, err := json.Marshal(sample)
	if err != nil {
		return err
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, s.url, bytes.NewReader(body))
	if err != nil {
		return err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Agent-Key", s.key)

	resp, err := s.client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	_, _ = io.Copy(io.Discard, resp.Body)

	switch {
	case resp.StatusCode == http.StatusUnauthorized:
		return ErrUnauthorized
	case resp.StatusCode >= 300:
		return fmt.Errorf("API returned status %d", resp.StatusCode)
	}
	return nil
}

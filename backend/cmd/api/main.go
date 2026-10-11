package main

import (
	"context"
	"errors"
	"flag"
	"log/slog"
	"os"
	"os/signal"
	"syscall"
	"time"

	"cloudpulse/agent"
	"cloudpulse/agent/collector"
	"cloudpulse/pkg/logger"
)

func main() {
	logger.Setup()

	key := flag.String("key", os.Getenv("CLOUDPULSE_KEY"), "agent key from the Add server dialog")
	apiURL := flag.String("api", env("CLOUDPULSE_API", "http://localhost:8080"), "CloudPulse API address")
	interval := flag.Duration("interval", time.Second, "how often to report")
	flag.Parse()

	if *key == "" {
		slog.Error("missing agent key", "usage", "agent -key cp_live_... [-api http://host:8080]")
		os.Exit(1)
	}

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	c := collector.New()
	sender := agent.NewSender(*apiURL, *key)
	slog.Info("agent started", "api", *apiURL, "interval", interval.String())

	ticker := time.NewTicker(*interval)
	defer ticker.Stop()

	failing := false // so a long outage logs once, not every second
	for {
		select {
		case <-ctx.Done():
			slog.Info("agent stopped")
			return
		case <-ticker.C:
			err := sender.Send(ctx, c.Collect())
			switch {
			case err == nil:
				if failing {
					slog.Info("reconnected")
					failing = false
				}
			case errors.Is(err, agent.ErrUnauthorized):
				slog.Error("agent key is wrong or the server was deleted")
				os.Exit(1)
			case !failing && ctx.Err() == nil:
				slog.Warn("cannot reach the API, will keep trying", "err", err)
				failing = true
			}
		}
	}
}

func env(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

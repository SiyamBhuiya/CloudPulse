package retention

import (
	"context"
	"log/slog"
	"time"

	"cloudpulse/internal/repository"
)

type Cleaner struct {
	metrics *repository.MetricRepo
	keep    time.Duration
}

func New(metrics *repository.MetricRepo, keepDays int) *Cleaner {
	return &Cleaner{metrics: metrics, keep: time.Duration(keepDays) * 24 * time.Hour}
}

// Run cleans up once shortly after startup, then every hour, until ctx is cancelled.
func (c *Cleaner) Run(ctx context.Context) {
	select {
	case <-ctx.Done():
		return
	case <-time.After(30 * time.Second):
	}
	c.once(ctx)

	tick := time.NewTicker(time.Hour)
	defer tick.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-tick.C:
			c.once(ctx)
		}
	}
}

func (c *Cleaner) once(ctx context.Context) {
	n, err := c.metrics.PruneOlderThan(ctx, time.Now().Add(-c.keep))
	switch {
	case err != nil && ctx.Err() == nil:
		slog.Error("prune metrics", "err", err)
	case n > 0:
		slog.Info("old metrics deleted", "rows", n)
	}
}

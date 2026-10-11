package collector

import (
	"context"
	"log/slog"
	"sync"
	"time"

	"cloudpulse/internal/models"
	"cloudpulse/internal/repository"
)

type Collector struct {
	in      chan models.Sample
	repo    *repository.MetricRepo
	workers int
	wg      sync.WaitGroup
}

func New(repo *repository.MetricRepo, buffer, workers int) *Collector {
	return &Collector{in: make(chan models.Sample, buffer), repo: repo, workers: workers}
}

// Start launches the writer goroutines. They stop when ctx is cancelled,
// after writing whatever is still waiting in the buffer.
func (c *Collector) Start(ctx context.Context) {
	for i := 0; i < c.workers; i++ {
		c.wg.Add(1)
		go c.worker(ctx)
	}
}

// Wait blocks until every worker has finished. Call it after cancelling ctx.
func (c *Collector) Wait() { c.wg.Wait() }

// Submit queues a sample without blocking. It returns false if the buffer is
// full, so the caller can tell the agent to retry.
func (c *Collector) Submit(s models.Sample) bool {
	select {
	case c.in <- s:
		return true
	default:
		return false
	}
}

func (c *Collector) worker(ctx context.Context) {
	defer c.wg.Done()
	for {
		select {
		case s := <-c.in:
			c.write(s)
		case <-ctx.Done():
			for { // drain what is left, then stop
				select {
				case s := <-c.in:
					c.write(s)
				default:
					return
				}
			}
		}
	}
}

func (c *Collector) write(s models.Sample) {
	// A fresh context, so a shutdown does not cancel the final writes
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := c.repo.Insert(ctx, s); err != nil {
		slog.Error("store metrics", "server", s.ServerID, "err", err)
	}
}

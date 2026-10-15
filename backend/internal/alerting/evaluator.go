package alerting

import (
	"context"
	"log/slog"
	"time"

	"cloudpulse/internal/repository"
)

const (
	evalEvery   = 10 * time.Second
	freshSec    = 30 // data older than this means the agent is offline
	coverageGap = 15 // seconds of slack when checking the window is fully covered
)

type Evaluator struct {
	repo     *repository.AlertRepo
	notifier *Notifier
}

func NewEvaluator(repo *repository.AlertRepo, n *Notifier) *Evaluator {
	return &Evaluator{repo: repo, notifier: n}
}

// Run checks every enabled rule every 10 seconds until ctx is cancelled.
func (e *Evaluator) Run(ctx context.Context) {
	tick := time.NewTicker(evalEvery)
	prune := time.NewTicker(time.Hour)
	defer tick.Stop()
	defer prune.Stop()

	for {
		select {
		case <-ctx.Done():
			e.notifier.Wait()
			return
		case <-tick.C:
			e.evaluate(ctx)
		case <-prune.C:
			if err := e.repo.PruneEvents(ctx); err != nil && ctx.Err() == nil {
				slog.Error("prune alert events", "err", err)
			}
		}
	}
}

func (e *Evaluator) evaluate(ctx context.Context) {
	cands, err := e.repo.Candidates(ctx)
	if err != nil {
		if ctx.Err() == nil {
			slog.Error("evaluate alerts", "err", err)
		}
		return
	}

	for _, c := range cands {
		// No recent data means the agent is offline. Leave the state as it is.
		if c.Latest == nil || c.LatestAgeSec == nil || *c.LatestAgeSec > freshSec {
			continue
		}

		if !c.Firing {
			windowSec := float64(c.DurationMin*60 - coverageGap)
			breach := c.Samples > 0 &&
				c.MinValue != nil && *c.MinValue > c.Threshold && // above the limit the whole time
				c.CoveredSec != nil && *c.CoveredSec >= windowSec // and we have data for the whole window
			if !breach {
				continue
			}
			opened, err := e.repo.Open(ctx, c, *c.Latest)
			if err != nil {
				slog.Error("open alert", "rule", c.RuleID, "err", err)
				continue
			}
			if opened {
				slog.Info("alert firing", "rule", c.RuleName, "server", c.ServerName, "value", *c.Latest)
				e.notifier.Notify(newAlert(c, *c.Latest, false))
			}
			continue
		}

		if *c.Latest <= c.Threshold {
			closed, err := e.repo.Close(ctx, c.RuleID, c.ServerID)
			if err != nil {
				slog.Error("close alert", "rule", c.RuleID, "err", err)
				continue
			}
			if closed {
				slog.Info("alert resolved", "rule", c.RuleName, "server", c.ServerName)
				e.notifier.Notify(newAlert(c, *c.Latest, true))
			}
		}
	}

	if err := e.repo.ResolveOrphans(ctx); err != nil && ctx.Err() == nil {
		slog.Error("resolve orphan alerts", "err", err)
	}
}

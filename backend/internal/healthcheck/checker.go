package healthcheck

import (
	"context"
	"errors"
	"log/slog"
	"net"
	"net/http"
	"sync"
	"syscall"
	"time"

	"cloudpulse/internal/repository"
)

const (
	maxParallel = 20
	slowAfterMs = 1000 // a correct answer slower than this counts as "slow"
)

// BlockedIP reports addresses a user-supplied URL must never reach:
// this machine, private networks, link-local (cloud metadata) and so on.
func BlockedIP(ip net.IP) bool {
	if ip.IsLoopback() || ip.IsPrivate() || ip.IsUnspecified() || ip.IsMulticast() ||
		ip.IsLinkLocalUnicast() || ip.IsLinkLocalMulticast() {
		return true
	}
	if v4 := ip.To4(); v4 != nil && v4[0] == 100 && v4[1] >= 64 && v4[1] <= 127 {
		return true // carrier-grade NAT range
	}
	return false
}

type Checker struct {
	repo   *repository.HealthRepo
	client *http.Client
	sem    chan struct{} // limits how many checks run at once
	wg     sync.WaitGroup
}

func New(repo *repository.HealthRepo, allowPrivate bool) *Checker {
	dialer := &net.Dialer{
		Timeout: 5 * time.Second,
		// Runs after DNS lookup, on the exact address about to be connected.
		// Checking here (not on the URL text) also covers redirects and DNS tricks.
		Control: func(network, address string, _ syscall.RawConn) error {
			if allowPrivate {
				return nil
			}
			host, _, err := net.SplitHostPort(address)
			if err != nil {
				return err
			}
			ip := net.ParseIP(host)
			if ip == nil || BlockedIP(ip) {
				return errors.New("target address is not allowed")
			}
			return nil
		},
	}

	return &Checker{
		repo: repo,
		sem:  make(chan struct{}, maxParallel),
		client: &http.Client{
			Timeout: 10 * time.Second,
			Transport: &http.Transport{
				Proxy:                 nil, // never go through a proxy from the environment
				DialContext:           dialer.DialContext,
				TLSHandshakeTimeout:   5 * time.Second,
				ResponseHeaderTimeout: 10 * time.Second,
				DisableKeepAlives:     true,
			},
			CheckRedirect: func(_ *http.Request, via []*http.Request) error {
				if len(via) >= 3 {
					return errors.New("too many redirects")
				}
				return nil
			},
		},
	}
}

// Run looks for due checks every 5 seconds until ctx is cancelled.
func (c *Checker) Run(ctx context.Context) {
	tick := time.NewTicker(5 * time.Second)
	prune := time.NewTicker(time.Hour)
	defer tick.Stop()
	defer prune.Stop()

	for {
		select {
		case <-ctx.Done():
			c.wg.Wait() // let running checks finish and save
			return
		case <-tick.C:
			c.dispatch(ctx)
		case <-prune.C:
			if err := c.repo.PruneResults(ctx); err != nil {
				slog.Error("prune check results", "err", err)
			}
		}
	}
}

func (c *Checker) dispatch(ctx context.Context) {
	free := cap(c.sem) - len(c.sem)
	if free == 0 {
		return
	}
	due, err := c.repo.ClaimDue(ctx, free)
	if err != nil {
		if ctx.Err() == nil {
			slog.Error("claim health checks", "err", err)
		}
		return
	}
	for _, d := range due {
		c.sem <- struct{}{} // never blocks: only this loop adds, and we counted free slots
		c.wg.Add(1)
		go func(d repository.DueCheck) {
			defer func() { <-c.sem; c.wg.Done() }()
			c.probe(d)
		}(d)
	}
}

// probe requests the URL once and saves what happened.
func (c *Checker) probe(d repository.DueCheck) {
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	var (
		ok     bool
		status = "down"
		ms     *int
		code   *int
	)

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, d.URL, nil)
	if err == nil {
		req.Header.Set("User-Agent", "CloudPulse-HealthCheck/1.0")
		start := time.Now()
		resp, err := c.client.Do(req)
		took := int(time.Since(start).Milliseconds())
		if err == nil {
			resp.Body.Close() // only the status matters, so the body is never read
			sc := resp.StatusCode
			code, ms = &sc, &took
			if sc == d.ExpectedStatus {
				ok = true
				status = "up"
				if took >= slowAfterMs {
					status = "slow"
				}
			}
		}
	}

	if err := c.repo.SaveResult(ctx, d.ID, ok, status, ms, code); err != nil {
		slog.Warn("save check result", "check", d.ID, "err", err)
	}
}

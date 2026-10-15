package middleware

import (
	"context"
	"math"
	"net"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"time"

	"golang.org/x/time/rate"
)

type visitor struct {
	lim  *rate.Limiter
	seen time.Time
}

// Limiter gives every key (usually a client IP) its own token bucket.
type Limiter struct {
	mu      sync.Mutex
	clients map[string]*visitor
	rate    rate.Limit
	burst   int
	key     func(*http.Request) string
}

// NewLimiter allows perSecond requests per key, with short bursts up to burst.
// Idle keys are forgotten after 10 minutes, until ctx is cancelled.
func NewLimiter(ctx context.Context, perSecond float64, burst int, key func(*http.Request) string) *Limiter {
	l := &Limiter{
		clients: make(map[string]*visitor),
		rate:    rate.Limit(perSecond),
		burst:   burst,
		key:     key,
	}
	go l.cleanup(ctx)
	return l
}

func (l *Limiter) allow(k string) bool {
	l.mu.Lock()
	defer l.mu.Unlock()

	v, ok := l.clients[k]
	if !ok {
		v = &visitor{lim: rate.NewLimiter(l.rate, l.burst)}
		l.clients[k] = v
	}
	v.seen = time.Now()
	return v.lim.Allow()
}

func (l *Limiter) cleanup(ctx context.Context) {
	tick := time.NewTicker(5 * time.Minute)
	defer tick.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-tick.C:
			l.mu.Lock()
			for k, v := range l.clients {
				if time.Since(v.seen) > 10*time.Minute {
					delete(l.clients, k)
				}
			}
			l.mu.Unlock()
		}
	}
}

// Middleware rejects requests over the limit with 429 and a Retry-After header.
func (l *Limiter) Middleware(next http.Handler) http.Handler {
	retry := strconv.Itoa(int(math.Max(1, math.Ceil(1/float64(l.rate)))))
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if !l.allow(l.key(r)) {
			w.Header().Set("Retry-After", retry)
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusTooManyRequests)
			_, _ = w.Write([]byte(`{"error":"too many requests, slow down"}`))
			return
		}
		next.ServeHTTP(w, r)
	})
}

// ClientIP returns the caller's address. Behind a reverse proxy such as Caddy,
// set trustProxy so the real client address is read from X-Forwarded-For.
// Only turn that on when a proxy is really in front, or clients could fake it.
func ClientIP(trustProxy bool) func(*http.Request) string {
	return func(r *http.Request) string {
		if trustProxy {
			// The proxy adds the real address last, so take the final entry
			parts := strings.Split(r.Header.Get("X-Forwarded-For"), ",")
			if ip := strings.TrimSpace(parts[len(parts)-1]); ip != "" {
				return ip
			}
		}
		host, _, err := net.SplitHostPort(r.RemoteAddr)
		if err != nil {
			return r.RemoteAddr
		}
		return host
	}
}

// AgentKey limits each agent separately, falling back to its IP.
func AgentKey(ip func(*http.Request) string) func(*http.Request) string {
	return func(r *http.Request) string {
		if k := r.Header.Get("X-Agent-Key"); k != "" {
			return "key:" + k
		}
		return "ip:" + ip(r)
	}
}

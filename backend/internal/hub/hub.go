package hub

import (
	"encoding/json"
	"sync"

	"cloudpulse/internal/models"
)

// Snapshot is the JSON the browser receives. The names match the
// MetricSnapshot type in the React app.
type Snapshot struct {
	ServerID  string  `json:"serverId"`
	Timestamp int64   `json:"timestamp"` // unix ms
	CPU       float64 `json:"cpu"`
	Memory    float64 `json:"memory"`
	Disk      float64 `json:"disk"`
	Network   float64 `json:"network"`
	Requests  int     `json:"requests"`  // not collected yet
	ErrorRate float64 `json:"errorRate"` // not collected yet
	P95       float64 `json:"p95"`       // not collected yet
}

// Hub keeps track of which browsers are watching which server.
type Hub struct {
	mu   sync.RWMutex
	subs map[string]map[*Client]struct{} // server id -> watchers
}

func New() *Hub {
	return &Hub{subs: make(map[string]map[*Client]struct{})}
}

func (h *Hub) subscribe(c *Client) {
	h.mu.Lock()
	defer h.mu.Unlock()
	set := h.subs[c.serverID]
	if set == nil {
		set = make(map[*Client]struct{})
		h.subs[c.serverID] = set
	}
	set[c] = struct{}{}
}

func (h *Hub) unsubscribe(c *Client) {
	h.mu.Lock()
	defer h.mu.Unlock()
	set := h.subs[c.serverID]
	if _, ok := set[c]; !ok {
		return // already removed
	}
	delete(set, c)
	if len(set) == 0 {
		delete(h.subs, c.serverID)
	}
	close(c.send)
}

// Publish sends a sample to everyone watching that server. It never blocks:
// a browser that cannot keep up is disconnected instead of slowing the rest.
func (h *Hub) Publish(s models.Sample) {
	h.mu.RLock()
	set := h.subs[s.ServerID]
	if len(set) == 0 {
		h.mu.RUnlock()
		return // nobody is watching, so skip the work
	}
	msg, err := json.Marshal(Snapshot{
		ServerID:  s.ServerID,
		Timestamp: s.At.UnixMilli(),
		CPU:       s.CPU,
		Memory:    s.Memory,
		Disk:      s.Disk,
		Network:   s.Network,
	})
	if err != nil {
		h.mu.RUnlock()
		return
	}

	var slow []*Client
	for c := range set {
		select {
		case c.send <- msg:
		default:
			slow = append(slow, c)
		}
	}
	h.mu.RUnlock()

	for _, c := range slow {
		h.unsubscribe(c)
	}
}

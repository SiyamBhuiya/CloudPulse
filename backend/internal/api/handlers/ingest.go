package handlers

import (
	"errors"
	"log/slog"
	"math"
	"net/http"
	"time"

	"cloudpulse/internal/collector"
	"cloudpulse/internal/hub"
	"cloudpulse/internal/models"
	"cloudpulse/internal/repository"
	"cloudpulse/internal/service"
)

type IngestHandler struct {
	servers *service.ServerService
	col     *collector.Collector
	hub     *hub.Hub
}

func NewIngestHandler(servers *service.ServerService, col *collector.Collector, h *hub.Hub) *IngestHandler {
	return &IngestHandler{servers: servers, col: col, hub: h}
}

type ingestRequest struct {
	CPU     float64 `json:"cpu"`
	Memory  float64 `json:"memory"`
	Disk    float64 `json:"disk"`
	Network float64 `json:"network"`
}

func validPercent(v float64) bool {
	return !math.IsNaN(v) && !math.IsInf(v, 0) && v >= 0 && v <= 100
}

// Ingest receives one sample from an agent. Agents log in with their key
// instead of a user token.
func (h *IngestHandler) Ingest(w http.ResponseWriter, r *http.Request) {
	key := r.Header.Get("X-Agent-Key")
	if key == "" {
		writeError(w, http.StatusUnauthorized, "missing agent key")
		return
	}
	serverID, err := h.servers.AuthenticateAgent(r.Context(), key)
	if errors.Is(err, repository.ErrNotFound) {
		writeError(w, http.StatusUnauthorized, "invalid agent key")
		return
	}
	if err != nil {
		slog.Error("agent auth", "err", err)
		writeError(w, http.StatusInternalServerError, "something went wrong")
		return
	}

	var req ingestRequest
	if !decode(w, r, &req) {
		return
	}
	if !validPercent(req.CPU) || !validPercent(req.Memory) || !validPercent(req.Disk) ||
		math.IsNaN(req.Network) || math.IsInf(req.Network, 0) || req.Network < 0 {
		writeError(w, http.StatusBadRequest, "metric values out of range")
		return
	}

	sample := models.Sample{
		ServerID: serverID,
		At:       time.Now(),
		CPU:      req.CPU,
		Memory:   req.Memory,
		Disk:     req.Disk,
		Network:  req.Network,
	}
	if !h.col.Submit(sample) {
		w.Header().Set("Retry-After", "2")
		writeError(w, http.StatusServiceUnavailable, "busy, try again")
		return
	}
	h.hub.Publish(sample) // live view, without waiting for the database
	w.WriteHeader(http.StatusNoContent)
}

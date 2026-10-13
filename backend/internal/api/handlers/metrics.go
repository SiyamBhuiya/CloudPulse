package handlers

import (
	"errors"
	"log/slog"
	"net/http"

	"github.com/go-chi/chi/v5"

	"cloudpulse/internal/api/middleware"
	"cloudpulse/internal/repository"
	"cloudpulse/internal/service"
)

type MetricHandler struct {
	svc *service.MetricService
}

func NewMetricHandler(svc *service.MetricService) *MetricHandler {
	return &MetricHandler{svc: svc}
}

func fail(w http.ResponseWriter, what string, err error) {
	switch {
	case errors.Is(err, repository.ErrNotFound):
		writeError(w, http.StatusNotFound, "server not found")
	case errors.Is(err, service.ErrBadRange):
		writeError(w, http.StatusBadRequest, err.Error())
	default:
		slog.Error(what, "err", err)
		writeError(w, http.StatusInternalServerError, "something went wrong")
	}
}

func (h *MetricHandler) History(w http.ResponseWriter, r *http.Request) {
	rng := r.URL.Query().Get("range")
	if rng == "" {
		rng = "6h"
	}
	points, err := h.svc.History(r.Context(), middleware.UserID(r.Context()), chi.URLParam(r, "id"), rng)
	if err != nil {
		fail(w, "history", err)
		return
	}
	writeJSON(w, http.StatusOK, points)
}

func (h *MetricHandler) Processes(w http.ResponseWriter, r *http.Request) {
	list, err := h.svc.Processes(r.Context(), middleware.UserID(r.Context()), chi.URLParam(r, "id"))
	if err != nil {
		fail(w, "processes", err)
		return
	}
	writeJSON(w, http.StatusOK, list)
}

// Events stays empty until alerts and health checks exist to create them.
func (h *MetricHandler) Events(w http.ResponseWriter, r *http.Request) {
	if err := h.svc.Exists(r.Context(), middleware.UserID(r.Context()), chi.URLParam(r, "id")); err != nil {
		fail(w, "events", err)
		return
	}
	writeJSON(w, http.StatusOK, []struct{}{})
}

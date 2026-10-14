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

type HealthHandler struct {
	svc *service.HealthService
}

func NewHealthHandler(svc *service.HealthService) *HealthHandler {
	return &HealthHandler{svc: svc}
}

func healthFail(w http.ResponseWriter, what string, err error) {
	var ve service.ValidationError
	switch {
	case errors.As(err, &ve):
		writeError(w, http.StatusBadRequest, ve.Msg)
	case errors.Is(err, repository.ErrNotFound):
		writeError(w, http.StatusNotFound, "health check not found")
	default:
		slog.Error(what, "err", err)
		writeError(w, http.StatusInternalServerError, "something went wrong")
	}
}

func (h *HealthHandler) List(w http.ResponseWriter, r *http.Request) {
	list, err := h.svc.List(r.Context(), middleware.UserID(r.Context()))
	if err != nil {
		healthFail(w, "list health checks", err)
		return
	}
	writeJSON(w, http.StatusOK, list)
}

type createCheckRequest struct {
	Name           string `json:"name"`
	URL            string `json:"url"`
	IntervalSec    int    `json:"intervalSec"`
	ExpectedStatus int    `json:"expectedStatus"`
}

func (h *HealthHandler) Create(w http.ResponseWriter, r *http.Request) {
	var req createCheckRequest
	if !decode(w, r, &req) {
		return
	}
	check, err := h.svc.Create(r.Context(), middleware.UserID(r.Context()),
		req.Name, req.URL, req.IntervalSec, req.ExpectedStatus)
	if err != nil {
		healthFail(w, "create health check", err)
		return
	}
	writeJSON(w, http.StatusCreated, check)
}

func (h *HealthHandler) SetEnabled(w http.ResponseWriter, r *http.Request) {
	var req struct {
		Enabled bool `json:"enabled"`
	}
	if !decode(w, r, &req) {
		return
	}
	err := h.svc.SetEnabled(r.Context(), middleware.UserID(r.Context()), chi.URLParam(r, "id"), req.Enabled)
	if err != nil {
		healthFail(w, "update health check", err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (h *HealthHandler) Delete(w http.ResponseWriter, r *http.Request) {
	err := h.svc.Delete(r.Context(), middleware.UserID(r.Context()), chi.URLParam(r, "id"))
	if err != nil {
		healthFail(w, "delete health check", err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

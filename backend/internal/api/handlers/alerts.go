package handlers

import (
	"errors"
	"log/slog"
	"net/http"

	"github.com/go-chi/chi/v5"

	"cloudpulse/internal/api/middleware"
	"cloudpulse/internal/models"
	"cloudpulse/internal/repository"
	"cloudpulse/internal/service"
)

type AlertHandler struct {
	svc *service.AlertService
}

func NewAlertHandler(svc *service.AlertService) *AlertHandler {
	return &AlertHandler{svc: svc}
}

func alertFail(w http.ResponseWriter, what string, err error) {
	var ve service.ValidationError
	switch {
	case errors.As(err, &ve):
		writeError(w, http.StatusBadRequest, ve.Msg)
	case errors.Is(err, repository.ErrNotFound):
		writeError(w, http.StatusNotFound, "not found")
	default:
		slog.Error(what, "err", err)
		writeError(w, http.StatusInternalServerError, "something went wrong")
	}
}

func (h *AlertHandler) ListRules(w http.ResponseWriter, r *http.Request) {
	list, err := h.svc.ListRules(r.Context(), middleware.UserID(r.Context()))
	if err != nil {
		alertFail(w, "list alert rules", err)
		return
	}
	writeJSON(w, http.StatusOK, list)
}

func (h *AlertHandler) CreateRule(w http.ResponseWriter, r *http.Request) {
	var in models.AlertRule
	if !decode(w, r, &in) {
		return
	}
	rule, err := h.svc.CreateRule(r.Context(), middleware.UserID(r.Context()), in)
	if err != nil {
		alertFail(w, "create alert rule", err)
		return
	}
	writeJSON(w, http.StatusCreated, rule)
}

func (h *AlertHandler) UpdateRule(w http.ResponseWriter, r *http.Request) {
	var in models.AlertRule
	if !decode(w, r, &in) {
		return
	}
	rule, err := h.svc.UpdateRule(r.Context(), middleware.UserID(r.Context()), chi.URLParam(r, "id"), in)
	if err != nil {
		alertFail(w, "update alert rule", err)
		return
	}
	writeJSON(w, http.StatusOK, rule)
}

func (h *AlertHandler) DeleteRule(w http.ResponseWriter, r *http.Request) {
	if err := h.svc.DeleteRule(r.Context(), middleware.UserID(r.Context()), chi.URLParam(r, "id")); err != nil {
		alertFail(w, "delete alert rule", err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

func (h *AlertHandler) History(w http.ResponseWriter, r *http.Request) {
	list, err := h.svc.History(r.Context(), middleware.UserID(r.Context()))
	if err != nil {
		alertFail(w, "alert history", err)
		return
	}
	writeJSON(w, http.StatusOK, list)
}

// ServerEvents feeds "Recent events" on the ServerDetail page.
func (h *AlertHandler) ServerEvents(w http.ResponseWriter, r *http.Request) {
	list, err := h.svc.ServerEvents(r.Context(), middleware.UserID(r.Context()), chi.URLParam(r, "id"))
	if err != nil {
		alertFail(w, "server events", err)
		return
	}
	writeJSON(w, http.StatusOK, list)
}

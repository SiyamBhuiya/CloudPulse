package handlers

import (
	"errors"
	"log/slog"
	"net/http"
	"strings"
	"unicode/utf8"

	"github.com/go-chi/chi/v5"

	"cloudpulse/internal/api/middleware"
	"cloudpulse/internal/repository"
	"cloudpulse/internal/service"
)

type ServerHandler struct {
	svc *service.ServerService
}

func NewServerHandler(svc *service.ServerService) *ServerHandler {
	return &ServerHandler{svc: svc}
}

type createServerRequest struct {
	Name   string `json:"name"`
	Region string `json:"region"`
}

func (h *ServerHandler) List(w http.ResponseWriter, r *http.Request) {
	servers, err := h.svc.List(r.Context(), middleware.UserID(r.Context()))
	if err != nil {
		slog.Error("list servers", "err", err)
		writeError(w, http.StatusInternalServerError, "something went wrong")
		return
	}
	writeJSON(w, http.StatusOK, servers)
}

func (h *ServerHandler) Create(w http.ResponseWriter, r *http.Request) {
	var req createServerRequest
	if !decode(w, r, &req) {
		return
	}
	req.Name = strings.TrimSpace(req.Name)
	req.Region = strings.TrimSpace(req.Region)

	switch {
	case req.Name == "":
		writeError(w, http.StatusBadRequest, "name is required")
		return
	case utf8.RuneCountInString(req.Name) > 60:
		writeError(w, http.StatusBadRequest, "name must be at most 60 characters")
		return
	case utf8.RuneCountInString(req.Region) > 60:
		writeError(w, http.StatusBadRequest, "region must be at most 60 characters")
		return
	}

	server, key, err := h.svc.Create(r.Context(), middleware.UserID(r.Context()), req.Name, req.Region)
	if err != nil {
		slog.Error("create server", "err", err)
		writeError(w, http.StatusInternalServerError, "something went wrong")
		return
	}
	writeJSON(w, http.StatusCreated, map[string]any{"server": server, "agentKey": key})
}

func (h *ServerHandler) Get(w http.ResponseWriter, r *http.Request) {
	server, err := h.svc.Get(r.Context(), middleware.UserID(r.Context()), chi.URLParam(r, "id"))
	if errors.Is(err, repository.ErrNotFound) {
		writeError(w, http.StatusNotFound, "server not found")
		return
	}
	if err != nil {
		slog.Error("get server", "err", err)
		writeError(w, http.StatusInternalServerError, "something went wrong")
		return
	}
	writeJSON(w, http.StatusOK, server)
}

func (h *ServerHandler) Delete(w http.ResponseWriter, r *http.Request) {
	err := h.svc.Delete(r.Context(), middleware.UserID(r.Context()), chi.URLParam(r, "id"))
	if errors.Is(err, repository.ErrNotFound) {
		writeError(w, http.StatusNotFound, "server not found")
		return
	}
	if err != nil {
		slog.Error("delete server", "err", err)
		writeError(w, http.StatusInternalServerError, "something went wrong")
		return
	}
	w.WriteHeader(http.StatusNoContent)
}
package handlers

import (
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"
	"net/mail"
	"strings"

	"cloudpulse/internal/repository"
	"cloudpulse/internal/service"
)

type AuthHandler struct {
	auth *service.AuthService
}

func NewAuthHandler(auth *service.AuthService) *AuthHandler {
	return &AuthHandler{auth: auth}
}

type registerRequest struct {
	Name     string `json:"name"`
	Email    string `json:"email"`
	Password string `json:"password"`
}

type loginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

func validEmail(s string) bool {
	a, err := mail.ParseAddress(s)
	return err == nil && a.Address == s
}

func decode(w http.ResponseWriter, r *http.Request, dst any) bool {
	r.Body = http.MaxBytesReader(w, r.Body, 1<<20) // 1 MB limit
	if err := json.NewDecoder(r.Body).Decode(dst); err != nil {
		writeError(w, http.StatusBadRequest, "invalid request body")
		return false
	}
	return true
}

func (h *AuthHandler) Register(w http.ResponseWriter, r *http.Request) {
	var req registerRequest
	if !decode(w, r, &req) {
		return
	}
	req.Name = strings.TrimSpace(req.Name)
	req.Email = strings.TrimSpace(req.Email)

	switch {
	case req.Name == "":
		writeError(w, http.StatusBadRequest, "name is required")
		return
	case !validEmail(req.Email):
		writeError(w, http.StatusBadRequest, "enter a valid email address")
		return
	case len(req.Password) < 8:
		writeError(w, http.StatusBadRequest, "password must be at least 8 characters")
		return
	case len(req.Password) > 72: // bcrypt only reads the first 72 bytes
		writeError(w, http.StatusBadRequest, "password must be at most 72 characters")
		return
	}

	token, user, err := h.auth.Register(r.Context(), req.Name, req.Email, req.Password)
	if errors.Is(err, repository.ErrEmailTaken) {
		writeError(w, http.StatusConflict, "email already registered")
		return
	}
	if err != nil {
		slog.Error("register failed", "err", err)
		writeError(w, http.StatusInternalServerError, "something went wrong")
		return
	}
	writeJSON(w, http.StatusCreated, map[string]any{"token": token, "user": user})
}

func (h *AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
	var req loginRequest
	if !decode(w, r, &req) {
		return
	}

	token, user, err := h.auth.Login(r.Context(), strings.TrimSpace(req.Email), req.Password)
	if errors.Is(err, service.ErrInvalidCredentials) {
		writeError(w, http.StatusUnauthorized, "invalid email or password")
		return
	}
	if err != nil {
		slog.Error("login failed", "err", err)
		writeError(w, http.StatusInternalServerError, "something went wrong")
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"token": token, "user": user})
}
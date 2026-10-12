package handlers

import (
	"errors"
	"log/slog"
	"net/http"

	"github.com/gorilla/websocket"

	"cloudpulse/internal/hub"
	"cloudpulse/internal/repository"
	"cloudpulse/internal/service"
)

type WSHandler struct {
	auth     *service.AuthService
	servers  *service.ServerService
	hub      *hub.Hub
	upgrader websocket.Upgrader
}

func NewWSHandler(auth *service.AuthService, servers *service.ServerService, h *hub.Hub, allowedOrigin string) *WSHandler {
	return &WSHandler{
		auth:    auth,
		servers: servers,
		hub:     h,
		upgrader: websocket.Upgrader{
			ReadBufferSize:  1024,
			WriteBufferSize: 1024,
			// Only the React app may open a connection from a browser
			CheckOrigin: func(r *http.Request) bool {
				o := r.Header.Get("Origin")
				return o == "" || o == allowedOrigin
			},
		},
	}
}

// pickServer returns the server to stream. With no ?server= (the Dashboard),
// it picks the user's server that reported most recently.
func (h *WSHandler) pickServer(r *http.Request, userID string) (string, error) {
	if id := r.URL.Query().Get("server"); id != "" {
		s, err := h.servers.Get(r.Context(), userID, id)
		if err != nil {
			return "", err
		}
		return s.ID, nil
	}

	list, err := h.servers.List(r.Context(), userID)
	if err != nil || len(list) == 0 {
		return "", err
	}
	best := list[0]
	for _, s := range list[1:] {
		if s.LastSeen > best.LastSeen {
			best = s
		}
	}
	return best.ID, nil
}

func (h *WSHandler) Serve(w http.ResponseWriter, r *http.Request) {
	userID, err := h.auth.ParseToken(r.URL.Query().Get("token"))
	if err != nil {
		writeError(w, http.StatusUnauthorized, "unauthorized")
		return
	}

	serverID, err := h.pickServer(r, userID)
	if errors.Is(err, repository.ErrNotFound) {
		writeError(w, http.StatusNotFound, "server not found")
		return
	}
	if err != nil {
		slog.Error("ws pick server", "err", err)
		writeError(w, http.StatusInternalServerError, "something went wrong")
		return
	}

	conn, err := h.upgrader.Upgrade(w, r, nil)
	if err != nil {
		return // Upgrade has already sent the error reply
	}
	h.hub.Serve(conn, serverID)
}

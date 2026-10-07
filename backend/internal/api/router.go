package api

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	chimw "github.com/go-chi/chi/v5/middleware"

	"cloudpulse/internal/api/handlers"
	"cloudpulse/internal/api/middleware"
	"cloudpulse/internal/config"
	"cloudpulse/internal/service"
)

func NewRouter(cfg *config.Config, auth *service.AuthService) http.Handler {
	r := chi.NewRouter()
	r.Use(chimw.Recoverer, middleware.Logging, middleware.CORS(cfg.AllowedOrigin))

	authHandler := handlers.NewAuthHandler(auth)

	r.Route("/api", func(r chi.Router) {
		r.Get("/health", func(w http.ResponseWriter, _ *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			_, _ = w.Write([]byte(`{"status":"ok"}`))
		})

		// Public
		r.Post("/auth/register", authHandler.Register)
		r.Post("/auth/login", authHandler.Login)

		// Needs a valid token. Servers, metrics and alerts routes go here later.
		r.Group(func(r chi.Router) {
			r.Use(middleware.Auth(auth.ParseToken))
			r.Get("/me", func(w http.ResponseWriter, r *http.Request) {
				w.Header().Set("Content-Type", "application/json")
				_, _ = w.Write([]byte(`{"userId":"` + middleware.UserID(r.Context()) + `"}`))
			})
		})
	})

	return r
}
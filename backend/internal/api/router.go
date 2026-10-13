package api

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	chimw "github.com/go-chi/chi/v5/middleware"

	"cloudpulse/internal/api/handlers"
	"cloudpulse/internal/api/middleware"
	"cloudpulse/internal/collector"
	"cloudpulse/internal/config"
	"cloudpulse/internal/hub"
	"cloudpulse/internal/service"
)

func NewRouter(
	cfg *config.Config,
	auth *service.AuthService,
	servers *service.ServerService,
	metrics *service.MetricService,
	col *collector.Collector,
	hb *hub.Hub,
) http.Handler {
	r := chi.NewRouter()
	r.Use(chimw.Recoverer, middleware.Logging, middleware.CORS(cfg.AllowedOrigin))

	authHandler := handlers.NewAuthHandler(auth)
	serverHandler := handlers.NewServerHandler(servers)
	metricHandler := handlers.NewMetricHandler(metrics)
	ingestHandler := handlers.NewIngestHandler(servers, col, hb)
	wsHandler := handlers.NewWSHandler(auth, servers, hb, cfg.AllowedOrigin)

	// Live stream for the browser. It checks the token itself (from ?token=).
	r.Get("/ws", wsHandler.Serve)

	r.Route("/api", func(r chi.Router) {
		r.Get("/health", func(w http.ResponseWriter, _ *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			_, _ = w.Write([]byte(`{"status":"ok"}`))
		})

		// Public
		r.Post("/auth/register", authHandler.Register)
		r.Post("/auth/login", authHandler.Login)

		// Agents send metrics here, using their agent key instead of a user token
		r.Post("/agent/metrics", ingestHandler.Ingest)

		// Needs a valid user token
		r.Group(func(r chi.Router) {
			r.Use(middleware.Auth(auth.ParseToken))

			r.Get("/me", func(w http.ResponseWriter, r *http.Request) {
				w.Header().Set("Content-Type", "application/json")
				_, _ = w.Write([]byte(`{"userId":"` + middleware.UserID(r.Context()) + `"}`))
			})

			r.Get("/servers", serverHandler.List)
			r.Post("/servers", serverHandler.Create)
			r.Get("/servers/{id}", serverHandler.Get)
			r.Delete("/servers/{id}", serverHandler.Delete)

			r.Get("/servers/{id}/metrics", metricHandler.History)
			r.Get("/servers/{id}/processes", metricHandler.Processes)
			r.Get("/servers/{id}/events", metricHandler.Events)
		})
	})

	return r
}

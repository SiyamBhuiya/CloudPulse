package api

import (
	"context"
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
	ctx context.Context,
	cfg *config.Config,
	auth *service.AuthService,
	servers *service.ServerService,
	metrics *service.MetricService,
	health *service.HealthService,
	alerts *service.AlertService,
	col *collector.Collector,
	hb *hub.Hub,
) http.Handler {
	r := chi.NewRouter()
	r.Use(chimw.Recoverer, middleware.Logging, middleware.CORS(cfg.AllowedOrigin))

	authHandler := handlers.NewAuthHandler(auth)
	serverHandler := handlers.NewServerHandler(servers)
	metricHandler := handlers.NewMetricHandler(metrics)
	healthHandler := handlers.NewHealthHandler(health)
	alertHandler := handlers.NewAlertHandler(alerts)
	ingestHandler := handlers.NewIngestHandler(servers, col, hb)
	wsHandler := handlers.NewWSHandler(auth, servers, hb, cfg.AllowedOrigin)

	ip := middleware.ClientIP(cfg.TrustProxy)
	apiLimit := middleware.NewLimiter(ctx, 20, 60, ip)                       // everything under /api
	authLimit := middleware.NewLimiter(ctx, 10.0/60, 5, ip)                  // login and register: 10 a minute
	agentLimit := middleware.NewLimiter(ctx, 5, 10, middleware.AgentKey(ip)) // per agent

	// Live stream for the browser. It checks the token itself (from ?token=).
	r.Get("/ws", wsHandler.Serve)

	r.Route("/api", func(r chi.Router) {
		r.Use(apiLimit.Middleware)

		r.Get("/health", func(w http.ResponseWriter, _ *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			_, _ = w.Write([]byte(`{"status":"ok"}`))
		})

		// Public, with the strict limit
		r.With(authLimit.Middleware).Post("/auth/register", authHandler.Register)
		r.With(authLimit.Middleware).Post("/auth/login", authHandler.Login)

		// Agents send metrics here, using their agent key instead of a user token
		r.With(agentLimit.Middleware).Post("/agent/metrics", ingestHandler.Ingest)

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
			r.Get("/servers/{id}/events", alertHandler.ServerEvents)

			r.Get("/health-checks", healthHandler.List)
			r.Post("/health-checks", healthHandler.Create)
			r.Patch("/health-checks/{id}", healthHandler.SetEnabled)
			r.Delete("/health-checks/{id}", healthHandler.Delete)

			r.Get("/alerts/rules", alertHandler.ListRules)
			r.Post("/alerts/rules", alertHandler.CreateRule)
			r.Put("/alerts/rules/{id}", alertHandler.UpdateRule)
			r.Delete("/alerts/rules/{id}", alertHandler.DeleteRule)
			r.Get("/alerts/history", alertHandler.History)
		})
	})

	return r
}

package main

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/joho/godotenv"

	"cloudpulse/internal/alerting"
	"cloudpulse/internal/api"
	"cloudpulse/internal/collector"
	"cloudpulse/internal/config"
	"cloudpulse/internal/db"
	"cloudpulse/internal/healthcheck"
	"cloudpulse/internal/hub"
	"cloudpulse/internal/repository"
	"cloudpulse/internal/retention"
	"cloudpulse/internal/service"
	"cloudpulse/pkg/logger"
)

func main() {
	logger.Setup()
	_ = godotenv.Load() // reads backend/.env if it exists

	cfg, err := config.Load()
	if err != nil {
		fatal("config", err)
	}

	if err := db.Migrate(cfg.DatabaseURL); err != nil {
		fatal("migrate", err)
	}
	pool, err := db.Connect(context.Background(), cfg.DatabaseURL)
	if err != nil {
		fatal("database", err)
	}
	defer pool.Close()

	userRepo := repository.NewUserRepo(pool)
	serverRepo := repository.NewServerRepo(pool)
	metricRepo := repository.NewMetricRepo(pool)
	healthRepo := repository.NewHealthRepo(pool)
	alertRepo := repository.NewAlertRepo(pool)

	auth := service.NewAuthService(userRepo, cfg.JWTSecret, cfg.JWTTTL)
	servers := service.NewServerService(serverRepo)
	metrics := service.NewMetricService(serverRepo, metricRepo)
	health := service.NewHealthService(healthRepo, cfg.AllowPrivateTargets)
	alerts := service.NewAlertService(alertRepo, serverRepo)

	if cfg.AllowPrivateTargets {
		slog.Warn("ALLOW_PRIVATE_TARGETS is on: health checks may reach localhost and private networks")
	}

	// Background workers all stop when workCtx is cancelled
	workCtx, stopWork := context.WithCancel(context.Background())

	// 1024 samples can wait in the buffer; 4 goroutines write them to Postgres
	col := collector.New(metricRepo, 1024, 4)
	col.Start(workCtx)

	// Pings the URLs from the Health checks page
	checker := healthcheck.New(healthRepo, cfg.AllowPrivateTargets)
	checkerDone := make(chan struct{})
	go func() {
		checker.Run(workCtx)
		close(checkerDone)
	}()

	// Checks alert rules against the metrics and sends notifications
	evaluator := alerting.NewEvaluator(alertRepo, alerting.NewNotifier(cfg))
	evaluatorDone := make(chan struct{})
	go func() {
		evaluator.Run(workCtx)
		close(evaluatorDone)
	}()

	// Deletes metrics older than the retention period
	cleaner := retention.New(metricRepo, cfg.MetricsRetentionDays)
	cleanerDone := make(chan struct{})
	go func() {
		cleaner.Run(workCtx)
		close(cleanerDone)
	}()

	// Pushes live metrics to browsers over WebSocket
	hb := hub.New()

	srv := &http.Server{
		Addr:              ":" + cfg.Port,
		Handler:           api.NewRouter(workCtx, cfg, auth, servers, metrics, health, alerts, col, hb),
		ReadHeaderTimeout: 5 * time.Second,
	}

	go func() {
		slog.Info("api listening", "addr", srv.Addr, "retentionDays", cfg.MetricsRetentionDays)
		if err := srv.ListenAndServe(); !errors.Is(err, http.ErrServerClosed) {
			fatal("server", err)
		}
	}()

	// Wait for Ctrl+C, then let in-flight requests finish
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt, syscall.SIGTERM)
	<-stop

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	_ = srv.Shutdown(ctx) // no new requests
	stopWork()            // then tell the background workers to finish
	col.Wait()
	<-checkerDone
	<-evaluatorDone
	<-cleanerDone
	slog.Info("stopped")
}

func fatal(what string, err error) {
	slog.Error("startup failed", "step", what, "err", err)
	os.Exit(1)
}

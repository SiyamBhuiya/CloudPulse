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

	"cloudpulse/internal/api"
	"cloudpulse/internal/collector"
	"cloudpulse/internal/config"
	"cloudpulse/internal/db"
	"cloudpulse/internal/hub"
	"cloudpulse/internal/repository"
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

	users := repository.NewUserRepo(pool)
	auth := service.NewAuthService(users, cfg.JWTSecret, cfg.JWTTTL)
	servers := service.NewServerService(repository.NewServerRepo(pool))

	// 1024 samples can wait in the buffer; 4 goroutines write them to Postgres
	workCtx, stopWork := context.WithCancel(context.Background())
	col := collector.New(repository.NewMetricRepo(pool), 1024, 4)
	col.Start(workCtx)

	// Pushes live metrics to browsers over WebSocket
	hb := hub.New()

	srv := &http.Server{
		Addr:              ":" + cfg.Port,
		Handler:           api.NewRouter(cfg, auth, servers, col, hb),
		ReadHeaderTimeout: 5 * time.Second,
	}

	go func() {
		slog.Info("api listening", "addr", srv.Addr)
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
	stopWork()            // then tell the writers to finish what is queued
	col.Wait()
	slog.Info("stopped")
}

func fatal(what string, err error) {
	slog.Error("startup failed", "step", what, "err", err)
	os.Exit(1)
}

package main

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"cloudpulse/internal/api"
	"cloudpulse/internal/collector"
	"cloudpulse/internal/config"
	"cloudpulse/internal/db"
	"cloudpulse/internal/repository"
	"cloudpulse/internal/service"
	"cloudpulse/pkg/logger"
)

func main() {
	logger.Setup()
	if err := run(); err != nil {
		slog.Error("API stopped", "err", err)
		os.Exit(1)
	}
}

func run() error {
	cfg, err := config.Load()
	if err != nil {
		return err
	}
	if err := db.Migrate(cfg.DatabaseURL); err != nil {
		return fmt.Errorf("migrate database: %w", err)
	}

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()
	workerCtx, cancelWorkers := context.WithCancel(ctx)
	defer cancelWorkers()

	pool, err := db.Connect(ctx, cfg.DatabaseURL)
	if err != nil {
		return fmt.Errorf("connect to database: %w", err)
	}
	defer pool.Close()

	users := repository.NewUserRepo(pool)
	servers := repository.NewServerRepo(pool)
	metrics := repository.NewMetricRepo(pool)
	authService := service.NewAuthService(users, cfg.JWTSecret, cfg.JWTTTL)
	serverService := service.NewServerService(servers)

	metricCollector := collector.New(metrics, 1024, 4)
	metricCollector.Start(workerCtx)

	server := &http.Server{
		Addr:              ":" + cfg.Port,
		Handler:           api.NewRouter(cfg, authService, serverService, metricCollector),
		ReadHeaderTimeout: 5 * time.Second,
	}

	slog.Info("API listening", "addr", server.Addr)
	serveErr := make(chan error, 1)
	go func() {
		serveErr <- server.ListenAndServe()
	}()

	var runErr error
	select {
	case <-ctx.Done():
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		runErr = server.Shutdown(shutdownCtx)
		cancel()
		if runErr != nil {
			runErr = fmt.Errorf("shut down HTTP server: %w", runErr)
		}
	case err := <-serveErr:
		if !errors.Is(err, http.ErrServerClosed) {
			runErr = fmt.Errorf("serve HTTP: %w", err)
		}
	}

	cancelWorkers()
	metricCollector.Wait()
	return runErr
}

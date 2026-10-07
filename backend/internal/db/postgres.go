package db

import (
	"context"
	"embed"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/pgx/v5"
	"github.com/golang-migrate/migrate/v4/source/iofs"
	"github.com/jackc/pgx/v5/pgxpool"
)

// The SQL files are compiled into the binary, so nothing extra is needed at runtime.
//
//go:embed migrations/*.sql
var migrationsFS embed.FS

// Connect opens a connection pool and checks that the database answers.
func Connect(ctx context.Context, url string) (*pgxpool.Pool, error) {
	pool, err := pgxpool.New(ctx, url)
	if err != nil {
		return nil, err
	}
	pingCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()
	if err := pool.Ping(pingCtx); err != nil {
		pool.Close()
		return nil, fmt.Errorf("ping database: %w", err)
	}
	return pool, nil
}

// Migrate applies any migration files that have not run yet.
func Migrate(url string) error {
	src, err := iofs.New(migrationsFS, "migrations")
	if err != nil {
		return err
	}

	// golang-migrate picks its driver from the URL scheme
	murl := strings.Replace(url, "postgresql://", "pgx5://", 1)
	murl = strings.Replace(murl, "postgres://", "pgx5://", 1)

	m, err := migrate.NewWithSourceInstance("iofs", src, murl)
	if err != nil {
		return err
	}
	defer m.Close()

	if err := m.Up(); err != nil && !errors.Is(err, migrate.ErrNoChange) {
		return err
	}
	return nil
}
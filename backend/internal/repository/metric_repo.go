package repository

import (
	"context"

	"github.com/jackc/pgx/v5/pgxpool"

	"cloudpulse/internal/models"
)

type MetricRepo struct {
	pool *pgxpool.Pool
}

func NewMetricRepo(pool *pgxpool.Pool) *MetricRepo {
	return &MetricRepo{pool: pool}
}

// Insert stores one sample and marks the server as seen, in a single statement.
func (r *MetricRepo) Insert(ctx context.Context, s models.Sample) error {
	_, err := r.pool.Exec(ctx,
		`WITH ins AS (
		     INSERT INTO metrics (server_id, ts, cpu, memory, disk, network)
		     VALUES ($1, $2, $3, $4, $5, $6)
		 )
		 UPDATE servers SET last_seen = $2 WHERE id = $1`,
		s.ServerID, s.At, s.CPU, s.Memory, s.Disk, s.Network,
	)
	return err
}

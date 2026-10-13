package repository

import (
	"context"
	"encoding/json"

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
// If the sample carries a process list, that is saved too.
func (r *MetricRepo) Insert(ctx context.Context, s models.Sample) error {
	_, err := r.pool.Exec(ctx,
		`WITH ins AS (
		     INSERT INTO metrics (server_id, ts, cpu, memory, disk, network)
		     VALUES ($1, $2, $3, $4, $5, $6)
		 )
		 UPDATE servers SET last_seen = $2 WHERE id = $1`,
		s.ServerID, s.At, s.CPU, s.Memory, s.Disk, s.Network,
	)
	if err != nil || len(s.Processes) == 0 {
		return err
	}

	raw, err := json.Marshal(s.Processes)
	if err != nil {
		return err
	}
	_, err = r.pool.Exec(ctx,
		`UPDATE servers SET processes = $2::jsonb WHERE id = $1`,
		s.ServerID, string(raw),
	)
	return err
}

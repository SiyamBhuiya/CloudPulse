package repository

import (
	"context"
	"encoding/json"
	"time"

	"cloudpulse/internal/models"
)

// History returns averages over fixed time buckets, so a week of data
// becomes about 72 points instead of 600,000 rows.
func (r *MetricRepo) History(ctx context.Context, serverID string, since time.Time, bucketSec int) ([]models.HistoryPoint, error) {
	rows, err := r.pool.Query(ctx,
		`SELECT floor(extract(epoch FROM ts)::float8 / $3::float8) * $3::float8 AS bucket,
		        avg(cpu), avg(memory), avg(disk)
		 FROM metrics
		 WHERE server_id = $1 AND ts >= $2
		 GROUP BY bucket
		 ORDER BY bucket`,
		serverID, since, bucketSec,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	points := make([]models.HistoryPoint, 0) // an empty list, not null, in the JSON
	for rows.Next() {
		var bucket float64
		var p models.HistoryPoint
		if err := rows.Scan(&bucket, &p.CPU, &p.Memory, &p.Disk); err != nil {
			return nil, err
		}
		p.T = int64(bucket * 1000)
		points = append(points, p)
	}
	return points, rows.Err()
}

// Processes returns the agent's most recent process list.
func (r *MetricRepo) Processes(ctx context.Context, serverID string) ([]models.ProcessInfo, error) {
	var raw []byte
	if err := r.pool.QueryRow(ctx,
		`SELECT processes FROM servers WHERE id = $1`, serverID).Scan(&raw); err != nil {
		return nil, err
	}
	list := make([]models.ProcessInfo, 0)
	if err := json.Unmarshal(raw, &list); err != nil {
		return nil, err
	}
	return list, nil
}

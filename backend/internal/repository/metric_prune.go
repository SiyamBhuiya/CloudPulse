package repository

import (
	"context"
	"time"
)

// PruneOlderThan deletes old samples in batches of 10,000, with a short pause
// between batches, so a big cleanup never locks the table for long.
func (r *MetricRepo) PruneOlderThan(ctx context.Context, cutoff time.Time) (int64, error) {
	const batch = 10000
	var total int64

	for {
		tag, err := r.pool.Exec(ctx,
			`DELETE FROM metrics
			 WHERE id IN (SELECT id FROM metrics WHERE ts < $1 ORDER BY ts LIMIT $2)`,
			cutoff, batch)
		if err != nil {
			return total, err
		}
		n := tag.RowsAffected()
		total += n
		if n < batch {
			return total, nil
		}

		select {
		case <-ctx.Done():
			return total, ctx.Err()
		case <-time.After(100 * time.Millisecond):
		}
	}
}

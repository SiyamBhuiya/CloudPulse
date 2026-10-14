package repository

import (
	"context"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"

	"cloudpulse/internal/models"
)

type HealthRepo struct {
	pool *pgxpool.Pool
}

func NewHealthRepo(pool *pgxpool.Pool) *HealthRepo {
	return &HealthRepo{pool: pool}
}

// DueCheck is what the checker needs to run one check.
type DueCheck struct {
	ID             string
	URL            string
	ExpectedStatus int
}

func emptyDays() []string {
	d := make([]string, 30)
	for i := range d {
		d[i] = "none"
	}
	return d
}

func dayStatus(pct float64) string {
	switch {
	case pct >= 99.5:
		return "ok"
	case pct >= 95:
		return "degraded"
	default:
		return "down"
	}
}

func (r *HealthRepo) Create(ctx context.Context, userID, name, url string, intervalSec, expected int) (*models.HealthCheck, error) {
	h := &models.HealthCheck{
		Name: name, URL: url, IntervalSec: intervalSec, ExpectedStatus: expected,
		Status: "pending", Uptime: 100, Enabled: true, Days: emptyDays(),
	}
	var created time.Time
	err := r.pool.QueryRow(ctx,
		`INSERT INTO health_checks (user_id, name, url, interval_sec, expected_status)
		 VALUES ($1, $2, $3, $4, $5)
		 RETURNING id::text, created_at`,
		userID, name, url, intervalSec, expected,
	).Scan(&h.ID, &created)
	if err != nil {
		return nil, err
	}
	h.LastChecked = created.UnixMilli()
	return h, nil
}

func (r *HealthRepo) CountByUser(ctx context.Context, userID string) (int, error) {
	var n int
	err := r.pool.QueryRow(ctx,
		`SELECT count(*) FROM health_checks WHERE user_id = $1`, userID).Scan(&n)
	return n, err
}

// ListByUser returns every check with its 30-day uptime and day-by-day strip.
func (r *HealthRepo) ListByUser(ctx context.Context, userID string) ([]models.HealthCheck, error) {
	rows, err := r.pool.Query(ctx,
		`SELECT c.id::text, c.name, c.url, c.interval_sec, c.expected_status, c.enabled,
		        c.created_at, c.last_checked, c.last_status, c.last_response_ms,
		        COALESCE(u.pct, 100)
		 FROM health_checks c
		 LEFT JOIN LATERAL (
		     SELECT 100.0 * count(*) FILTER (WHERE ok) / NULLIF(count(*), 0) AS pct
		     FROM check_results
		     WHERE check_id = c.id AND ts >= now() - interval '30 days'
		 ) u ON true
		 WHERE c.user_id = $1
		 ORDER BY c.created_at`,
		userID,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	list := make([]models.HealthCheck, 0)
	for rows.Next() {
		var h models.HealthCheck
		var created time.Time
		var lastChecked *time.Time
		var lastStatus *string
		if err := rows.Scan(&h.ID, &h.Name, &h.URL, &h.IntervalSec, &h.ExpectedStatus, &h.Enabled,
			&created, &lastChecked, &lastStatus, &h.ResponseMs, &h.Uptime); err != nil {
			return nil, err
		}
		h.Status = "pending"
		if lastStatus != nil {
			h.Status = *lastStatus
		}
		seen := created
		if lastChecked != nil {
			seen = *lastChecked
		}
		h.LastChecked = seen.UnixMilli()
		h.Days = emptyDays()
		list = append(list, h)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	if len(list) == 0 {
		return list, nil
	}

	// One query for the day strips of all checks
	start := time.Now().UTC().Truncate(24*time.Hour).AddDate(0, 0, -29)
	drows, err := r.pool.Query(ctx,
		`SELECT r.check_id::text,
		        (r.ts AT TIME ZONE 'UTC')::date AS day,
		        100.0 * count(*) FILTER (WHERE r.ok) / count(*)
		 FROM check_results r
		 JOIN health_checks c ON c.id = r.check_id
		 WHERE c.user_id = $1 AND r.ts >= $2
		 GROUP BY 1, 2`,
		userID, start,
	)
	if err != nil {
		return nil, err
	}
	defer drows.Close()

	index := make(map[string]int, len(list))
	for i := range list {
		index[list[i].ID] = i
	}
	for drows.Next() {
		var id string
		var day time.Time
		var pct float64
		if err := drows.Scan(&id, &day, &pct); err != nil {
			return nil, err
		}
		i, ok := index[id]
		pos := int(day.UTC().Sub(start).Hours() / 24)
		if ok && pos >= 0 && pos < 30 {
			list[i].Days[pos] = dayStatus(pct)
		}
	}
	return list, drows.Err()
}

func (r *HealthRepo) SetEnabled(ctx context.Context, userID, id string, enabled bool) error {
	tag, err := r.pool.Exec(ctx,
		`UPDATE health_checks SET enabled = $3 WHERE id = $1 AND user_id = $2`,
		id, userID, enabled)
	if err = mapNotFound(err); err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return ErrNotFound
	}
	return nil
}

func (r *HealthRepo) Delete(ctx context.Context, userID, id string) error {
	tag, err := r.pool.Exec(ctx,
		`DELETE FROM health_checks WHERE id = $1 AND user_id = $2`, id, userID)
	if err = mapNotFound(err); err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return ErrNotFound
	}
	return nil
}

/* ---------- used by the background checker ---------- */

// ClaimDue picks checks whose interval has passed and marks them as taken in
// one step, so two workers can never run the same check twice.
func (r *HealthRepo) ClaimDue(ctx context.Context, limit int) ([]DueCheck, error) {
	rows, err := r.pool.Query(ctx,
		`UPDATE health_checks SET last_checked = now()
		 WHERE id IN (
		     SELECT id FROM health_checks
		     WHERE enabled
		       AND (last_checked IS NULL OR last_checked + interval_sec * interval '1 second' <= now())
		     ORDER BY last_checked NULLS FIRST
		     LIMIT $1
		     FOR UPDATE SKIP LOCKED
		 )
		 RETURNING id::text, url, expected_status`,
		limit,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var due []DueCheck
	for rows.Next() {
		var d DueCheck
		if err := rows.Scan(&d.ID, &d.URL, &d.ExpectedStatus); err != nil {
			return nil, err
		}
		due = append(due, d)
	}
	return due, rows.Err()
}

// SaveResult stores one run and updates the check's current status.
func (r *HealthRepo) SaveResult(ctx context.Context, id string, ok bool, status string, ms, code *int) error {
	_, err := r.pool.Exec(ctx,
		`WITH ins AS (
		     INSERT INTO check_results (check_id, ok, response_ms, status_code)
		     VALUES ($1, $2, $3, $4)
		 )
		 UPDATE health_checks SET last_status = $5, last_response_ms = $3 WHERE id = $1`,
		id, ok, ms, code, status,
	)
	return err
}

// PruneResults removes runs older than 35 days (the strip shows 30).
func (r *HealthRepo) PruneResults(ctx context.Context) error {
	_, err := r.pool.Exec(ctx,
		`DELETE FROM check_results WHERE ts < now() - interval '35 days'`)
	return err
}

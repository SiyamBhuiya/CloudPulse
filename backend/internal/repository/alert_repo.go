package repository

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"cloudpulse/internal/models"
)

type AlertRepo struct {
	pool *pgxpool.Pool
}

func NewAlertRepo(pool *pgxpool.Pool) *AlertRepo {
	return &AlertRepo{pool: pool}
}

const ruleCols = `id::text, name, COALESCE(server_id::text, 'all'), metric, threshold, duration_min, channel, enabled`

func scanRule(row pgx.Row) (*models.AlertRule, error) {
	r := &models.AlertRule{}
	err := row.Scan(&r.ID, &r.Name, &r.ServerID, &r.Metric, &r.Threshold, &r.DurationMin, &r.Channel, &r.Enabled)
	if err != nil {
		return nil, err
	}
	return r, nil
}

// "all" is stored as NULL
func serverParam(id string) *string {
	if id == "all" {
		return nil
	}
	return &id
}

/* ---------- rules ---------- */

func (r *AlertRepo) CountRules(ctx context.Context, userID string) (int, error) {
	var n int
	err := r.pool.QueryRow(ctx, `SELECT count(*) FROM alert_rules WHERE user_id = $1`, userID).Scan(&n)
	return n, err
}

func (r *AlertRepo) ListRules(ctx context.Context, userID string) ([]models.AlertRule, error) {
	rows, err := r.pool.Query(ctx,
		`SELECT `+ruleCols+` FROM alert_rules WHERE user_id = $1 ORDER BY created_at`, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	list := make([]models.AlertRule, 0)
	for rows.Next() {
		rule, err := scanRule(rows)
		if err != nil {
			return nil, err
		}
		list = append(list, *rule)
	}
	return list, rows.Err()
}

func (r *AlertRepo) CreateRule(ctx context.Context, userID string, in models.AlertRule) (*models.AlertRule, error) {
	return scanRule(r.pool.QueryRow(ctx,
		`INSERT INTO alert_rules (user_id, name, server_id, metric, threshold, duration_min, channel, enabled)
		 VALUES ($1, $2, $3::uuid, $4, $5, $6, $7, $8)
		 RETURNING `+ruleCols,
		userID, in.Name, serverParam(in.ServerID), in.Metric, in.Threshold, in.DurationMin, in.Channel, in.Enabled,
	))
}

// UpdateRule also closes the rule's open incidents, because the conditions may have changed.
func (r *AlertRepo) UpdateRule(ctx context.Context, userID, id string, in models.AlertRule) (*models.AlertRule, error) {
	rule, err := scanRule(r.pool.QueryRow(ctx,
		`UPDATE alert_rules
		 SET name = $3, server_id = $4::uuid, metric = $5, threshold = $6,
		     duration_min = $7, channel = $8, enabled = $9
		 WHERE id = $1 AND user_id = $2
		 RETURNING `+ruleCols,
		id, userID, in.Name, serverParam(in.ServerID), in.Metric, in.Threshold, in.DurationMin, in.Channel, in.Enabled,
	))
	if err = mapNotFound(err); err != nil {
		return nil, err
	}
	if _, err := r.pool.Exec(ctx,
		`UPDATE alert_events SET status = 'resolved', resolved_at = now()
		 WHERE rule_id = $1 AND user_id = $2 AND status = 'firing'`, id, userID); err != nil {
		return nil, err
	}
	return rule, nil
}

func (r *AlertRepo) DeleteRule(ctx context.Context, userID, id string) error {
	// Close incidents first: deleting the rule would unlink them
	if _, err := r.pool.Exec(ctx,
		`UPDATE alert_events SET status = 'resolved', resolved_at = now()
		 WHERE rule_id = $1 AND user_id = $2 AND status = 'firing'`, id, userID); err != nil {
		return mapNotFound(err)
	}
	tag, err := r.pool.Exec(ctx, `DELETE FROM alert_rules WHERE id = $1 AND user_id = $2`, id, userID)
	if err = mapNotFound(err); err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return ErrNotFound
	}
	return nil
}

/* ---------- history ---------- */

func (r *AlertRepo) History(ctx context.Context, userID string) ([]models.AlertEvent, error) {
	rows, err := r.pool.Query(ctx,
		`SELECT id::text, rule_name, server_name, fired_at, status
		 FROM alert_events WHERE user_id = $1
		 ORDER BY fired_at DESC LIMIT 100`, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	list := make([]models.AlertEvent, 0)
	for rows.Next() {
		var e models.AlertEvent
		var at time.Time
		if err := rows.Scan(&e.ID, &e.RuleName, &e.ServerName, &at, &e.Status); err != nil {
			return nil, err
		}
		e.At = at.UnixMilli()
		list = append(list, e)
	}
	return list, rows.Err()
}

var metricNames = map[string]string{"cpu": "CPU", "memory": "Memory", "disk": "Disk"}

func (r *AlertRepo) ServerEvents(ctx context.Context, userID, serverID string) ([]models.ServerEvent, error) {
	rows, err := r.pool.Query(ctx,
		`SELECT id::text, metric, threshold, fired_at, status
		 FROM alert_events WHERE user_id = $1 AND server_id = $2
		 ORDER BY fired_at DESC LIMIT 10`, userID, serverID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	list := make([]models.ServerEvent, 0)
	for rows.Next() {
		var e models.ServerEvent
		var metric, status string
		var threshold float64
		var at time.Time
		if err := rows.Scan(&e.ID, &metric, &threshold, &at, &status); err != nil {
			return nil, err
		}
		e.Kind = "alert"
		e.Text = fmt.Sprintf("%s above %.0f%%", metricNames[metric], threshold)
		e.At = at.UnixMilli()
		e.Resolved = status == "resolved"
		list = append(list, e)
	}
	return list, rows.Err()
}

/* ---------- used by the evaluator ---------- */

// Candidate is one (rule, server) pair with what the evaluator needs to judge it.
type Candidate struct {
	RuleID       string
	RuleName     string
	Metric       string
	Threshold    float64
	DurationMin  int
	Channel      string
	Email        string
	ServerID     string
	ServerName   string
	Samples      int64
	MinValue     *float64 // lowest reading inside the window
	CoveredSec   *float64 // how far back the oldest reading in the window goes
	Latest       *float64
	LatestAgeSec *float64
	Firing       bool
}

func (r *AlertRepo) Candidates(ctx context.Context) ([]Candidate, error) {
	rows, err := r.pool.Query(ctx,
		`SELECT r.id::text, r.name, r.metric, r.threshold, r.duration_min, r.channel, u.email,
		        s.id::text, s.name,
		        COALESCE(w.cnt, 0), w.min_val,
		        EXTRACT(EPOCH FROM (now() - w.min_ts))::float8,
		        l.v,
		        EXTRACT(EPOCH FROM (now() - l.ts))::float8,
		        (e.id IS NOT NULL)
		 FROM alert_rules r
		 JOIN users u ON u.id = r.user_id
		 JOIN servers s ON s.user_id = r.user_id AND (r.server_id IS NULL OR r.server_id = s.id)
		 LEFT JOIN LATERAL (
		     SELECT count(*) AS cnt, min(x.v) AS min_val, min(x.ts) AS min_ts
		     FROM (
		         SELECT m.ts,
		                CASE r.metric WHEN 'cpu' THEN m.cpu WHEN 'memory' THEN m.memory ELSE m.disk END AS v
		         FROM metrics m
		         WHERE m.server_id = s.id AND m.ts >= now() - r.duration_min * interval '1 minute'
		     ) x
		 ) w ON true
		 LEFT JOIN LATERAL (
		     SELECT m.ts,
		            CASE r.metric WHEN 'cpu' THEN m.cpu WHEN 'memory' THEN m.memory ELSE m.disk END AS v
		     FROM metrics m WHERE m.server_id = s.id ORDER BY m.ts DESC LIMIT 1
		 ) l ON true
		 LEFT JOIN alert_events e ON e.rule_id = r.id AND e.server_id = s.id AND e.status = 'firing'
		 WHERE r.enabled`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var list []Candidate
	for rows.Next() {
		var c Candidate
		if err := rows.Scan(&c.RuleID, &c.RuleName, &c.Metric, &c.Threshold, &c.DurationMin, &c.Channel, &c.Email,
			&c.ServerID, &c.ServerName, &c.Samples, &c.MinValue, &c.CoveredSec,
			&c.Latest, &c.LatestAgeSec, &c.Firing); err != nil {
			return nil, err
		}
		list = append(list, c)
	}
	return list, rows.Err()
}

// Open records a new incident. It returns false if one is already open.
func (r *AlertRepo) Open(ctx context.Context, c Candidate, value float64) (bool, error) {
	tag, err := r.pool.Exec(ctx,
		`INSERT INTO alert_events (rule_id, user_id, server_id, rule_name, server_name, metric, threshold, value, status)
		 SELECT $1::uuid, user_id, $2::uuid, $3, $4, $5, $6, $7, 'firing'
		 FROM alert_rules WHERE id = $1::uuid
		 ON CONFLICT (rule_id, server_id) WHERE status = 'firing' DO NOTHING`,
		c.RuleID, c.ServerID, c.RuleName, c.ServerName, c.Metric, c.Threshold, value)
	if err != nil {
		return false, err
	}
	return tag.RowsAffected() == 1, nil
}

// Close marks an open incident resolved. It returns false if there was none.
func (r *AlertRepo) Close(ctx context.Context, ruleID, serverID string) (bool, error) {
	tag, err := r.pool.Exec(ctx,
		`UPDATE alert_events SET status = 'resolved', resolved_at = now()
		 WHERE rule_id = $1::uuid AND server_id = $2::uuid AND status = 'firing'`, ruleID, serverID)
	if err != nil {
		return false, err
	}
	return tag.RowsAffected() > 0, nil
}

// ResolveOrphans closes incidents whose server was deleted.
func (r *AlertRepo) ResolveOrphans(ctx context.Context) error {
	_, err := r.pool.Exec(ctx,
		`UPDATE alert_events SET status = 'resolved', resolved_at = now()
		 WHERE status = 'firing' AND (server_id IS NULL OR rule_id IS NULL)`)
	return err
}

func (r *AlertRepo) PruneEvents(ctx context.Context) error {
	_, err := r.pool.Exec(ctx,
		`DELETE FROM alert_events WHERE status = 'resolved' AND fired_at < now() - interval '90 days'`)
	return err
}

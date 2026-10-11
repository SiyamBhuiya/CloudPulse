package repository

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"

	"cloudpulse/internal/models"
)

type ServerRepo struct {
	pool *pgxpool.Pool
}

func NewServerRepo(pool *pgxpool.Pool) *ServerRepo {
	return &ServerRepo{pool: pool}
}

// Every server query returns these 7 columns, newest metric included.
const serverSelect = `
SELECT s.id::text, s.name, s.region, s.last_seen, s.created_at,
       COALESCE(m.cpu, 0), COALESCE(m.memory, 0)
FROM servers s
LEFT JOIN LATERAL (
    SELECT cpu, memory FROM metrics
    WHERE server_id = s.id ORDER BY ts DESC LIMIT 1
) m ON true`

func scanServer(row pgx.Row) (*models.Server, error) {
	s := &models.Server{}
	err := row.Scan(&s.ID, &s.Name, &s.Region, &s.LastSeenAt, &s.CreatedAt, &s.CPU, &s.Memory)
	if err != nil {
		return nil, err
	}
	return s, nil
}

func (r *ServerRepo) Create(ctx context.Context, userID, name, region, keyHash string) (*models.Server, error) {
	return scanServer(r.pool.QueryRow(ctx,
		`INSERT INTO servers (user_id, name, region, agent_key_hash)
		 VALUES ($1, $2, $3, $4)
		 RETURNING id::text, name, region, last_seen, created_at, 0::float8, 0::float8`,
		userID, name, region, keyHash,
	))
}

func (r *ServerRepo) ListByUser(ctx context.Context, userID string) ([]models.Server, error) {
	rows, err := r.pool.Query(ctx,
		serverSelect+` WHERE s.user_id = $1 ORDER BY s.created_at`, userID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	servers := make([]models.Server, 0) // an empty list, not null, in the JSON
	for rows.Next() {
		s, err := scanServer(rows)
		if err != nil {
			return nil, err
		}
		servers = append(servers, *s)
	}
	return servers, rows.Err()
}

func (r *ServerRepo) Get(ctx context.Context, userID, id string) (*models.Server, error) {
	s, err := scanServer(r.pool.QueryRow(ctx,
		serverSelect+` WHERE s.id = $1 AND s.user_id = $2`, id, userID))
	return s, mapNotFound(err)
}

func (r *ServerRepo) Delete(ctx context.Context, userID, id string) error {
	tag, err := r.pool.Exec(ctx,
		`DELETE FROM servers WHERE id = $1 AND user_id = $2`, id, userID)
	if err = mapNotFound(err); err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return ErrNotFound
	}
	return nil
}

// IDByKeyHash finds the server an agent key belongs to.
func (r *ServerRepo) IDByKeyHash(ctx context.Context, hash string) (string, error) {
	var id string
	err := r.pool.QueryRow(ctx,
		`SELECT id::text FROM servers WHERE agent_key_hash = $1`, hash).Scan(&id)
	return id, mapNotFound(err)
}

// A missing row, or an id that is not a valid UUID, both mean "not found".
func mapNotFound(err error) error {
	if errors.Is(err, pgx.ErrNoRows) {
		return ErrNotFound
	}
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) && pgErr.Code == "22P02" { // invalid UUID text
		return ErrNotFound
	}
	return err
}

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

const serverCols = `id::text, name, region, last_seen, created_at`

func scanServer(row pgx.Row) (*models.Server, error) {
	s := &models.Server{}
	if err := row.Scan(&s.ID, &s.Name, &s.Region, &s.LastSeenAt, &s.CreatedAt); err != nil {
		return nil, err
	}
	return s, nil
}

func (r *ServerRepo) Create(ctx context.Context, userID, name, region, keyHash string) (*models.Server, error) {
	return scanServer(r.pool.QueryRow(ctx,
		`INSERT INTO servers (user_id, name, region, agent_key_hash)
		 VALUES ($1, $2, $3, $4)
		 RETURNING `+serverCols,
		userID, name, region, keyHash,
	))
}

func (r *ServerRepo) ListByUser(ctx context.Context, userID string) ([]models.Server, error) {
	rows, err := r.pool.Query(ctx,
		`SELECT `+serverCols+` FROM servers WHERE user_id = $1 ORDER BY created_at`,
		userID,
	)
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
		`SELECT `+serverCols+` FROM servers WHERE id = $1 AND user_id = $2`,
		id, userID,
	))
	return s, mapNotFound(err)
}

func (r *ServerRepo) Delete(ctx context.Context, userID, id string) error {
	tag, err := r.pool.Exec(ctx,
		`DELETE FROM servers WHERE id = $1 AND user_id = $2`,
		id, userID,
	)
	if err = mapNotFound(err); err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return ErrNotFound
	}
	return nil
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
package repository

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"

	"cloudpulse/internal/models"
)

var (
	ErrEmailTaken = errors.New("email already registered")
	ErrNotFound   = errors.New("not found")
)

type UserRepo struct {
	pool *pgxpool.Pool
}

func NewUserRepo(pool *pgxpool.Pool) *UserRepo {
	return &UserRepo{pool: pool}
}

func (r *UserRepo) Create(ctx context.Context, name, email, passwordHash string) (*models.User, error) {
	u := &models.User{}
	err := r.pool.QueryRow(ctx,
		`INSERT INTO users (name, email, password_hash)
		 VALUES ($1, $2, $3)
		 RETURNING id::text, name, email, password_hash, created_at`,
		name, email, passwordHash,
	).Scan(&u.ID, &u.Name, &u.Email, &u.PasswordHash, &u.CreatedAt)

	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) && pgErr.Code == "23505" { // unique violation
		return nil, ErrEmailTaken
	}
	if err != nil {
		return nil, err
	}
	return u, nil
}

func (r *UserRepo) GetByEmail(ctx context.Context, email string) (*models.User, error) {
	u := &models.User{}
	err := r.pool.QueryRow(ctx,
		`SELECT id::text, name, email, password_hash, created_at
		 FROM users WHERE lower(email) = lower($1)`,
		email,
	).Scan(&u.ID, &u.Name, &u.Email, &u.PasswordHash, &u.CreatedAt)

	if errors.Is(err, pgx.ErrNoRows) {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, err
	}
	return u, nil
}
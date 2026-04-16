package repository

import (
	"context"
	"errors"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

var ErrNotFound = errors.New("user not found")

type User struct {
	ID        string    `json:"id"`
	Email     string    `json:"email"`
	FullName  string    `json:"full_name"`
	Phone     string    `json:"phone"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type Repository interface {
	Migrate(ctx context.Context) error
	Create(ctx context.Context, user User) error
	GetByID(ctx context.Context, id string) (User, error)
	Update(ctx context.Context, user User) error
	Delete(ctx context.Context, id string) error
	List(ctx context.Context, limit, offset int) ([]User, error)
}

type PostgresRepository struct {
	pool *pgxpool.Pool
}

func NewPostgresRepository(pool *pgxpool.Pool) *PostgresRepository {
	return &PostgresRepository{pool: pool}
}

func (r *PostgresRepository) Migrate(ctx context.Context) error {
	query := `
CREATE TABLE IF NOT EXISTS users (
    id UUID PRIMARY KEY,
    email TEXT UNIQUE NOT NULL,
    full_name TEXT NOT NULL,
    phone TEXT NOT NULL,
    created_at TIMESTAMPTZ NOT NULL,
    updated_at TIMESTAMPTZ NOT NULL
);
`
	_, err := r.pool.Exec(ctx, query)
	return err
}

func (r *PostgresRepository) Create(ctx context.Context, user User) error {
	_, err := r.pool.Exec(ctx,
		`INSERT INTO users (id,email,full_name,phone,created_at,updated_at) VALUES ($1,$2,$3,$4,$5,$6)`,
		user.ID, user.Email, user.FullName, user.Phone, user.CreatedAt, user.UpdatedAt,
	)
	return err
}

func (r *PostgresRepository) GetByID(ctx context.Context, id string) (User, error) {
	var user User
	err := r.pool.QueryRow(ctx,
		`SELECT id,email,full_name,phone,created_at,updated_at FROM users WHERE id=$1`, id).
		Scan(&user.ID, &user.Email, &user.FullName, &user.Phone, &user.CreatedAt, &user.UpdatedAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return User{}, ErrNotFound
		}
		return User{}, err
	}
	return user, nil
}

func (r *PostgresRepository) Update(ctx context.Context, user User) error {
	tag, err := r.pool.Exec(ctx,
		`UPDATE users SET full_name=$2, phone=$3, updated_at=$4 WHERE id=$1`,
		user.ID, user.FullName, user.Phone, user.UpdatedAt,
	)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return ErrNotFound
	}
	return nil
}

func (r *PostgresRepository) Delete(ctx context.Context, id string) error {
	tag, err := r.pool.Exec(ctx, `DELETE FROM users WHERE id=$1`, id)
	if err != nil {
		return err
	}
	if tag.RowsAffected() == 0 {
		return ErrNotFound
	}
	return nil
}

func (r *PostgresRepository) List(ctx context.Context, limit, offset int) ([]User, error) {
	rows, err := r.pool.Query(ctx,
		`SELECT id,email,full_name,phone,created_at,updated_at FROM users ORDER BY created_at DESC LIMIT $1 OFFSET $2`,
		limit, offset,
	)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	users := make([]User, 0, limit)
	for rows.Next() {
		var user User
		if err = rows.Scan(&user.ID, &user.Email, &user.FullName, &user.Phone, &user.CreatedAt, &user.UpdatedAt); err != nil {
			return nil, err
		}
		users = append(users, user)
	}
	return users, rows.Err()
}

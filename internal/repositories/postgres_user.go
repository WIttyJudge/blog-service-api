package repositories

import (
	"context"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/wittyjudge/blog-service-api/internal/domains"
)

type PostgresUser struct {
	ctx    context.Context
	pgPool *pgxpool.Pool
}

func NewPostgresUser(ctx context.Context, pgPool *pgxpool.Pool) *PostgresUser {
	return &PostgresUser{
		ctx:    ctx,
		pgPool: pgPool,
	}
}

func (p *PostgresUser) GetByEmail(email string) (*domains.User, error) {
	sql := `
	  SELECT id, first_name, last_name, email, password, created_at, updated_at
		FROM users
		WHERE email = @email
	`

	args := pgx.NamedArgs{"email": email}

	user := &domains.User{}
	_ = p.pgPool.QueryRow(p.ctx, sql, args).Scan(
		&user.ID,
		&user.FirstName,
		&user.LastName,
		&user.Email,
		&user.Password,
		&user.CreatedAt,
		&user.UpdatedAt,
	)

	return user, nil
}

func (p *PostgresUser) Create(user *domains.User) error {
	sql := `
		INSERT INTO users (first_name, last_name, email, password)
		VALUES (@firstName, @lastName, @email, @password)
		RETURNING id, email
	`

	args := pgx.NamedArgs{
		"firstName": user.FirstName,
		"lastName":  user.LastName,
		"email":     user.Email,
		"password":  user.Password,
	}

	return p.pgPool.QueryRow(p.ctx, sql, args).Scan(
		&user.ID,
		&user.Email,
	)
}

func (p *PostgresUser) CheckIfExistsByEmail(email string) bool {
	sql := "SELECT EXISTS(SELECT 1 FROM users where email = @email)"
	args := pgx.NamedArgs{"email": email}

	var resp bool
	_ = p.pgPool.QueryRow(p.ctx, sql, args).Scan(&resp)

	return resp
}

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

func (p *PostgresUser) GetByUsername(username string) (*domains.User, error) {
	return nil, nil
}

func (p *PostgresUser) GetByEmail(email string) (*domains.User, error) {
	return nil, nil
}

func (p *PostgresUser) Create(user *domains.User) error {
	sql := `
	INSERT INTO users (first_name, last_name, username, email, password)
	VALUES (@firstName, @lastName, @username, @email, @password)
	`

	args := pgx.NamedArgs{
		"firstName": user.FirstName,
		"lastName":  user.LastName,
		"username":  user.Username,
		"email":     user.Email,
		"password":  user.Password,
	}

	_, err := p.pgPool.Exec(p.ctx, sql, args)
	if err != nil {
		return err
	}

	return nil
}

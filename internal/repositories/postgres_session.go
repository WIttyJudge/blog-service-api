package repositories

import (
	"context"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/wittyjudge/blog-service-api/internal/domains"
)

type PostgresSession struct {
	ctx    context.Context
	pgPool *pgxpool.Pool
}

func NewPostgresSession(ctx context.Context, pgPool *pgxpool.Pool) *PostgresSession {
	return &PostgresSession{
		ctx:    ctx,
		pgPool: pgPool,
	}
}

func (p *PostgresSession) CreateOrUpdate(session *domains.Session) error {
	sql := `
		INSERT INTO sessions (user_id, refresh_token, expires_at)
		VALUES (@userID, @refreshToken, @expiresAt)
		ON CONFLICT(user_id) DO
		  UPDATE SET
			  refresh_token = @refreshToken,
				expires_at = @expiresAt,
				updated_at = @updatedAt
	`

	args := pgx.NamedArgs{
		"userID":       session.UserID,
		"refreshToken": session.RefreshToken,
		"expiresAt":    session.ExpiresAt,
		"updatedAt":    time.Now(),
	}

	_, err := p.pgPool.Exec(p.ctx, sql, args)
	return err
}

func (p *PostgresSession) Delete(userID int) error {
	sql := `DELETE FROM sessions WHERE id = @userID`
	args := pgx.NamedArgs{"userID": userID}

	_, err := p.pgPool.Exec(p.ctx, sql, args)
	return err
}

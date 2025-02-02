package database

import (
	"context"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/wittyjudge/blog-service-api/internal/config"
)

func NewPostgreSQLPool(ctx context.Context, config config.PostgreSQL) (*pgxpool.Pool, error) {
	pool, err := pgxpool.New(ctx, config.ConnectionURL())
	if err != nil {
		return nil, err
	}

	if err := pool.Ping(ctx); err != nil {
		return nil, err
	}

	return pool, nil
}

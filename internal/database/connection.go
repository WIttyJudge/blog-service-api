package database

import (
	"context"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/redis/go-redis/v9"
	"github.com/wittyjudge/blog-service-api/internal/config"
)

func NewPostgreSQLPool(ctx context.Context, config config.Postgres) (*pgxpool.Pool, error) {
	pool, err := pgxpool.New(ctx, config.ConnectionDSN())
	if err != nil {
		return nil, err
	}

	if err := pool.Ping(ctx); err != nil {
		return nil, err
	}

	return pool, nil
}

func NewRedisClient(ctx context.Context, config config.Redis) (*redis.Client, error) {
	opts := &redis.Options{
		Addr:     config.HostPort(),
		Password: config.Password,
		DB:       0,
		PoolSize: config.PoolMaxConns,
	}

	client := redis.NewClient(opts)

	if err := client.Ping(ctx).Err(); err != nil {
		return nil, err
	}

	return client, nil
}

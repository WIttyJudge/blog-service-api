package app

import (
	"context"

	"github.com/wittyjudge/blog-service-api/internal/api"
	"github.com/wittyjudge/blog-service-api/internal/config"
	"github.com/wittyjudge/blog-service-api/internal/database"
	"github.com/wittyjudge/blog-service-api/pkg/logger"
	"go.uber.org/zap"
)

func StartAPI(ctx context.Context, config *config.Config) error {
	logger := logger.New(config.Environment)
	defer logger.Sync()

	pgPool, err := database.NewPostgreSQLPool(ctx, config.Databases.Postgres)
	if err != nil {
		return err
	}
	defer pgPool.Close()

	redisClient, err := database.NewRedisClient(ctx, config.Databases.Redis)
	if err != nil {
		return err
	}
	defer redisClient.Close()

	api := api.NewAPI(ctx, config, logger, pgPool, redisClient)

	go func() {
		if err := api.Start(); err != nil {
			logger.Error("failed to start HTTP server", zap.Error(err))
		}
	}()

	<-ctx.Done()

	if err := api.Stop(); err != nil {
		return err
	}

	return nil
}

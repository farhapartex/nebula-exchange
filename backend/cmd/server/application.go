package main

import (
	"context"
	"log/slog"

	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"

	"github.com/farhapartex/nebula-exchange/backend/internal/health"
	"github.com/farhapartex/nebula-exchange/backend/internal/platform/config"
	"github.com/farhapartex/nebula-exchange/backend/internal/platform/httpserver"
)

func buildRouter(appConfig config.Config, appLogger *slog.Logger, gormDatabase *gorm.DB, redisClient *redis.Client) (*gin.Engine, error) {
	healthService := health.NewService(appLogger,
		health.NewProbe("database", func(ctx context.Context) error {
			sqlDatabase, err := gormDatabase.DB()
			if err != nil {
				return err
			}
			return sqlDatabase.PingContext(ctx)
		}),
		health.NewProbe("redis", func(ctx context.Context) error {
			return redisClient.Ping(ctx).Err()
		}),
	)

	return httpserver.NewRouter(
		httpserver.RouterOptions{
			Logger:         appLogger,
			IsProduction:   appConfig.IsProduction(),
			AllowedOrigins: appConfig.AllowedOrigins,
			TrustedProxies: appConfig.TrustedProxies,
		},
		health.NewHandler(healthService),
	)
}

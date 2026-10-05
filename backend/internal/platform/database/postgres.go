package database

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	gormlogger "gorm.io/gorm/logger"

	"github.com/farhapartex/nebula-exchange/backend/internal/platform/config"
)

const connectTimeout = 5 * time.Second

func Open(ctx context.Context, databaseConfig config.DatabaseConfig, logger *slog.Logger) (*gorm.DB, error) {
	database, err := gorm.Open(postgres.Open(databaseConfig.URL), &gorm.Config{
		Logger:                 newQueryLogger(logger),
		NowFunc:                func() time.Time { return time.Now().UTC() },
		TranslateError:         true,
		SkipDefaultTransaction: true,
	})
	if err != nil {
		return nil, fmt.Errorf("open postgres: %w", err)
	}

	sqlDatabase, err := database.DB()
	if err != nil {
		return nil, fmt.Errorf("access connection pool: %w", err)
	}
	sqlDatabase.SetMaxOpenConns(databaseConfig.MaxOpenConnections)
	sqlDatabase.SetMaxIdleConns(databaseConfig.MaxIdleConnections)
	sqlDatabase.SetConnMaxLifetime(databaseConfig.ConnectionMaxLifetime)

	pingContext, cancelPing := context.WithTimeout(ctx, connectTimeout)
	defer cancelPing()
	if err := sqlDatabase.PingContext(pingContext); err != nil {
		_ = sqlDatabase.Close()
		return nil, fmt.Errorf("ping postgres: %w", err)
	}
	return database, nil
}

func Close(database *gorm.DB) error {
	sqlDatabase, err := database.DB()
	if err != nil {
		return err
	}
	return sqlDatabase.Close()
}

func newQueryLogger(logger *slog.Logger) gormlogger.Interface {
	return gormlogger.NewSlogLogger(logger, gormlogger.Config{
		SlowThreshold:             200 * time.Millisecond,
		LogLevel:                  gormlogger.Warn,
		IgnoreRecordNotFoundError: true,
		ParameterizedQueries:      true,
	})
}

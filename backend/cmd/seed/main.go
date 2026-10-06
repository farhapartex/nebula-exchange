package main

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"syscall"

	"github.com/farhapartex/nebula-exchange/backend/internal/platform/config"
	"github.com/farhapartex/nebula-exchange/backend/internal/platform/database"
	"github.com/farhapartex/nebula-exchange/backend/internal/platform/logger"
	"github.com/farhapartex/nebula-exchange/backend/internal/platform/objectstorage"
	"github.com/farhapartex/nebula-exchange/backend/internal/story/seeding"
)

func main() {
	appLogger := logger.New(slog.LevelInfo, false)
	if err := run(os.Args[1:], appLogger); err != nil {
		appLogger.Error("seed failed", slog.Any("error", err))
		os.Exit(1)
	}
}

func run(levelPackageDirectories []string, appLogger *slog.Logger) error {
	if len(levelPackageDirectories) == 0 {
		return errors.New("usage: seed <level package directory>...")
	}
	databaseConfig, err := config.LoadDatabaseConfig()
	if err != nil {
		return err
	}
	storageConfig, err := config.LoadStorageConfig()
	if err != nil {
		return err
	}

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	gormDatabase, err := database.Open(ctx, databaseConfig, appLogger)
	if err != nil {
		return fmt.Errorf("connect database: %w", err)
	}
	defer database.Close(gormDatabase)

	assetStorage, err := objectstorage.NewMinioStorage(storageConfig)
	if err != nil {
		return err
	}

	seeder := seeding.NewSeeder(gormDatabase, assetStorage, appLogger)
	for _, levelPackageDirectory := range levelPackageDirectories {
		levelPackage, err := seeding.LoadLevelPackage(levelPackageDirectory)
		if err != nil {
			return fmt.Errorf("load %s: %w", levelPackageDirectory, err)
		}
		if err := seeder.SeedLevel(ctx, levelPackage); err != nil {
			return fmt.Errorf("seed %s: %w", levelPackageDirectory, err)
		}
	}
	return nil
}

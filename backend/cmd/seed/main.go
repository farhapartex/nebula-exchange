package main

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"os"
	"os/signal"
	"path/filepath"
	"sort"
	"syscall"

	"gorm.io/gorm"

	paymentseeding "github.com/farhapartex/nebula-exchange/backend/internal/payment/seeding"
	"github.com/farhapartex/nebula-exchange/backend/internal/platform/config"
	"github.com/farhapartex/nebula-exchange/backend/internal/platform/database"
	"github.com/farhapartex/nebula-exchange/backend/internal/platform/logger"
	"github.com/farhapartex/nebula-exchange/backend/internal/platform/objectstorage"
	progressseeding "github.com/farhapartex/nebula-exchange/backend/internal/progress/seeding"
	storyseeding "github.com/farhapartex/nebula-exchange/backend/internal/story/seeding"
)

const (
	plansFileName         = "plans/plans.json"
	fightersDirectoryName = "fighters"
	storyDirectoryName    = "story"
)

func main() {
	appLogger := logger.New(slog.LevelInfo, false)
	if err := run(os.Args[1:], appLogger); err != nil {
		appLogger.Error("seed failed", slog.Any("error", err))
		os.Exit(1)
	}
}

func run(arguments []string, appLogger *slog.Logger) error {
	if len(arguments) != 1 {
		return errors.New("usage: seed <seeds directory>")
	}
	seedsDirectory := arguments[0]

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

	plans, err := paymentseeding.LoadPlans(filepath.Join(seedsDirectory, plansFileName))
	if err != nil {
		return err
	}
	if err := paymentseeding.SeedPlans(ctx, gormDatabase, plans, appLogger); err != nil {
		return err
	}
	if err := seedFighterTemplates(ctx, gormDatabase, filepath.Join(seedsDirectory, fightersDirectoryName), appLogger); err != nil {
		return err
	}
	return seedStoryLevels(ctx, storyseeding.NewSeeder(gormDatabase, assetStorage, appLogger), filepath.Join(seedsDirectory, storyDirectoryName))
}

func seedFighterTemplates(ctx context.Context, gormDatabase *gorm.DB, fightersDirectory string, appLogger *slog.Logger) error {
	templateFiles, err := filepath.Glob(filepath.Join(fightersDirectory, "*.json"))
	if err != nil {
		return err
	}
	sort.Strings(templateFiles)
	for _, templateFile := range templateFiles {
		template, err := progressseeding.LoadFighterTemplate(templateFile)
		if err != nil {
			return fmt.Errorf("load %s: %w", templateFile, err)
		}
		if err := progressseeding.SeedFighterTemplate(ctx, gormDatabase, template, appLogger); err != nil {
			return err
		}
	}
	return nil
}

func seedStoryLevels(ctx context.Context, seeder *storyseeding.Seeder, storyDirectory string) error {
	levelFiles, err := filepath.Glob(filepath.Join(storyDirectory, "*", "level.json"))
	if err != nil {
		return err
	}
	sort.Strings(levelFiles)
	for _, levelFile := range levelFiles {
		levelPackageDirectory := filepath.Dir(levelFile)
		levelPackage, err := storyseeding.LoadLevelPackage(levelPackageDirectory)
		if err != nil {
			return fmt.Errorf("load %s: %w", levelPackageDirectory, err)
		}
		if err := seeder.SeedLevel(ctx, levelPackage); err != nil {
			return fmt.Errorf("seed %s: %w", levelPackageDirectory, err)
		}
	}
	return nil
}

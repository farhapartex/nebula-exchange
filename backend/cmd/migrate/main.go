package main

import (
	"errors"
	"fmt"
	"log/slog"
	"os"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"

	"github.com/farhapartex/nebula-exchange/backend/internal/platform/config"
	"github.com/farhapartex/nebula-exchange/backend/internal/platform/logger"
)

const (
	commandUp      = "up"
	commandDown    = "down"
	commandVersion = "version"
)

func main() {
	appLogger := logger.New(slog.LevelInfo, false)
	command := commandUp
	if len(os.Args) > 1 {
		command = os.Args[1]
	}
	if err := run(command, appLogger); err != nil {
		appLogger.Error("migration failed", slog.String("command", command), slog.Any("error", err))
		os.Exit(1)
	}
}

func run(command string, appLogger *slog.Logger) error {
	migrationConfig, err := config.LoadMigrationConfig()
	if err != nil {
		return fmt.Errorf("load config: %w", err)
	}

	migrator, err := migrate.New(migrationConfig.MigrationsSource, migrationConfig.DatabaseURL)
	if err != nil {
		return fmt.Errorf("open migrations: %w", err)
	}
	defer migrator.Close()

	switch command {
	case commandUp:
		err = migrator.Up()
	case commandDown:
		err = migrator.Steps(-1)
	case commandVersion:
	default:
		return fmt.Errorf("unknown command %q, use up, down or version", command)
	}
	if err != nil && !isNothingToMigrate(err) {
		return err
	}

	return logVersion(migrator, appLogger)
}

func isNothingToMigrate(err error) bool {
	return errors.Is(err, migrate.ErrNoChange) || errors.Is(err, os.ErrNotExist)
}

func logVersion(migrator *migrate.Migrate, appLogger *slog.Logger) error {
	version, isDirty, err := migrator.Version()
	if errors.Is(err, migrate.ErrNilVersion) {
		appLogger.Info("database has no migrations applied")
		return nil
	}
	if err != nil {
		return fmt.Errorf("read version: %w", err)
	}
	appLogger.Info("database migration version", slog.Uint64("version", uint64(version)), slog.Bool("dirty", isDirty))
	return nil
}

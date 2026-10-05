package databasetest

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net/url"
	"os"
	"path/filepath"
	"runtime"
	"testing"
	"time"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"gorm.io/gorm"

	"github.com/farhapartex/nebula-exchange/backend/internal/platform/config"
	"github.com/farhapartex/nebula-exchange/backend/internal/platform/database"
)

const (
	adminURLEnvironmentKey        = "TEST_DATABASE_ADMIN_URL"
	requireDatabaseEnvironmentKey = "REQUIRE_DATABASE_TESTS"
	defaultAdminURL               = "postgres://nebula:nebula@localhost:5432/postgres?sslmode=disable"
	testDatabaseNamePrefix        = "street_test_"
)

func Open(t *testing.T) *gorm.DB {
	t.Helper()
	adminURL := readAdminURL()
	adminDatabase, err := openQuietly(adminURL)
	if err != nil {
		skipOrFail(t, fmt.Errorf("connect to test postgres: %w", err))
		return nil
	}

	testDatabaseName := testDatabaseNamePrefix + randomSuffix(t)
	if err := adminDatabase.Exec("CREATE DATABASE " + testDatabaseName).Error; err != nil {
		t.Fatalf("create test database: %v", err)
	}
	t.Cleanup(func() {
		if err := adminDatabase.Exec("DROP DATABASE IF EXISTS " + testDatabaseName + " WITH (FORCE)").Error; err != nil {
			t.Errorf("drop test database: %v", err)
		}
		_ = database.Close(adminDatabase)
	})

	testDatabaseURL, err := replaceDatabaseName(adminURL, testDatabaseName)
	if err != nil {
		t.Fatalf("build test database url: %v", err)
	}
	if err := applyMigrations(testDatabaseURL); err != nil {
		t.Fatalf("apply migrations: %v", err)
	}

	testDatabase, err := openQuietly(testDatabaseURL)
	if err != nil {
		t.Fatalf("open test database: %v", err)
	}
	t.Cleanup(func() { _ = database.Close(testDatabase) })
	return testDatabase
}

func openQuietly(databaseURL string) (*gorm.DB, error) {
	quietLogger := slog.New(slog.NewTextHandler(io.Discard, nil))
	return database.Open(context.Background(), config.DatabaseConfig{
		URL:                   databaseURL,
		MaxOpenConnections:    5,
		MaxIdleConnections:    1,
		ConnectionMaxLifetime: time.Minute,
	}, quietLogger)
}

func readAdminURL() string {
	if adminURL := os.Getenv(adminURLEnvironmentKey); adminURL != "" {
		return adminURL
	}
	return defaultAdminURL
}

func skipOrFail(t *testing.T, reason error) {
	t.Helper()
	if os.Getenv(requireDatabaseEnvironmentKey) == "true" {
		t.Fatal(reason)
	}
	t.Skipf("skipping database test, start postgres with make docker-up: %v", reason)
}

func randomSuffix(t *testing.T) string {
	t.Helper()
	randomBytes := make([]byte, 6)
	if _, err := rand.Read(randomBytes); err != nil {
		t.Fatalf("generate database name: %v", err)
	}
	return hex.EncodeToString(randomBytes)
}

func replaceDatabaseName(connectionURL, databaseName string) (string, error) {
	parsedURL, err := url.Parse(connectionURL)
	if err != nil {
		return "", err
	}
	parsedURL.Path = "/" + databaseName
	return parsedURL.String(), nil
}

func applyMigrations(databaseURL string) error {
	migrator, err := migrate.New("file://"+migrationsDirectory(), databaseURL)
	if err != nil {
		return err
	}
	defer migrator.Close()
	err = migrator.Up()
	if errors.Is(err, migrate.ErrNoChange) || errors.Is(err, os.ErrNotExist) {
		return nil
	}
	return err
}

func migrationsDirectory() string {
	_, currentFile, _, _ := runtime.Caller(0)
	return filepath.Join(filepath.Dir(currentFile), "..", "..", "..", "..", "migrations")
}

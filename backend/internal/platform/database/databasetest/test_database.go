package databasetest

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"net/url"
	"os"
	"path/filepath"
	"runtime"
	"testing"

	"github.com/golang-migrate/migrate/v4"
	_ "github.com/golang-migrate/migrate/v4/database/pgx/v5"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"nebula-exchange/backend/internal/platform/database"
)

const (
	adminURLEnvironmentKey        = "TEST_DATABASE_ADMIN_URL"
	requireDatabaseEnvironmentKey = "REQUIRE_DATABASE_TESTS"
	defaultAdminURL               = "postgres://nebula:nebula@localhost:5432/postgres?sslmode=disable"
	testDatabaseNamePrefix        = "nebula_test_"
)

func NewPool(t *testing.T) *pgxpool.Pool {
	t.Helper()
	ctx := context.Background()

	adminURL := readAdminURL()
	adminConnection, err := pgx.Connect(ctx, adminURL)
	if err != nil {
		skipOrFail(t, fmt.Errorf("connect to test postgres at %s: %w", adminURL, err))
		return nil
	}
	defer adminConnection.Close(ctx)

	testDatabaseName := testDatabaseNamePrefix + randomSuffix(t)
	if _, err := adminConnection.Exec(ctx, "CREATE DATABASE "+pgx.Identifier{testDatabaseName}.Sanitize()); err != nil {
		t.Fatalf("create test database: %v", err)
	}
	t.Cleanup(func() { dropDatabase(t, adminURL, testDatabaseName) })

	testDatabaseURL, err := replaceDatabaseName(adminURL, testDatabaseName)
	if err != nil {
		t.Fatalf("build test database url: %v", err)
	}
	if err := applyMigrations(testDatabaseURL); err != nil {
		t.Fatalf("apply migrations: %v", err)
	}

	poolConfig, err := pgxpool.ParseConfig(testDatabaseURL)
	if err != nil {
		t.Fatalf("parse test database url: %v", err)
	}
	pool, err := database.OpenPool(ctx, poolConfig)
	if err != nil {
		t.Fatalf("open test database pool: %v", err)
	}
	t.Cleanup(pool.Close)
	return pool
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
	migrationsSource := "file://" + migrationsDirectory()
	migrationDatabaseURL, err := withScheme(databaseURL, "pgx5")
	if err != nil {
		return err
	}

	migrator, err := migrate.New(migrationsSource, migrationDatabaseURL)
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

func withScheme(connectionURL, scheme string) (string, error) {
	parsedURL, err := url.Parse(connectionURL)
	if err != nil {
		return "", err
	}
	parsedURL.Scheme = scheme
	return parsedURL.String(), nil
}

func migrationsDirectory() string {
	_, currentFile, _, _ := runtime.Caller(0)
	return filepath.Join(filepath.Dir(currentFile), "..", "..", "..", "..", "migrations")
}

func dropDatabase(t *testing.T, adminURL, databaseName string) {
	ctx := context.Background()
	adminConnection, err := pgx.Connect(ctx, adminURL)
	if err != nil {
		t.Errorf("connect to drop test database: %v", err)
		return
	}
	defer adminConnection.Close(ctx)

	dropStatement := "DROP DATABASE IF EXISTS " + pgx.Identifier{databaseName}.Sanitize() + " WITH (FORCE)"
	if _, err := adminConnection.Exec(ctx, dropStatement); err != nil {
		t.Errorf("drop test database: %v", err)
	}
}

package database_test

import (
	"context"
	"errors"
	"testing"

	"github.com/jackc/pgx/v5"

	"nebula-exchange/backend/internal/platform/database"
	"nebula-exchange/backend/internal/platform/database/databasetest"
)

func TestWithTransactionCommitsOnSuccessAndRollsBackOnError(t *testing.T) {
	pool := databasetest.NewPool(t)
	ctx := context.Background()

	if _, err := pool.Exec(ctx, "CREATE TABLE transaction_probe (label TEXT NOT NULL)"); err != nil {
		t.Fatalf("create probe table: %v", err)
	}

	err := database.WithTransaction(ctx, pool, func(transaction pgx.Tx) error {
		_, err := transaction.Exec(ctx, "INSERT INTO transaction_probe (label) VALUES ('committed')")
		return err
	})
	if err != nil {
		t.Fatalf("committed transaction failed: %v", err)
	}

	failedWork := errors.New("work failed")
	err = database.WithTransaction(ctx, pool, func(transaction pgx.Tx) error {
		if _, err := transaction.Exec(ctx, "INSERT INTO transaction_probe (label) VALUES ('rolled_back')"); err != nil {
			return err
		}
		return failedWork
	})
	if !errors.Is(err, failedWork) {
		t.Fatalf("got error %v, want %v", err, failedWork)
	}

	rows, err := pool.Query(ctx, "SELECT label FROM transaction_probe ORDER BY label")
	if err != nil {
		t.Fatalf("query probe table: %v", err)
	}
	storedLabels, err := pgx.CollectRows(rows, pgx.RowTo[string])
	if err != nil {
		t.Fatalf("collect probe rows: %v", err)
	}

	if len(storedLabels) != 1 || storedLabels[0] != "committed" {
		t.Fatalf("got labels %v, want only committed", storedLabels)
	}
}

func TestTestDatabasesAreIsolated(t *testing.T) {
	firstPool := databasetest.NewPool(t)
	secondPool := databasetest.NewPool(t)
	ctx := context.Background()

	if _, err := firstPool.Exec(ctx, "CREATE TABLE isolation_probe (id INT)"); err != nil {
		t.Fatalf("create table in first database: %v", err)
	}

	var tableCount int
	err := secondPool.QueryRow(ctx, "SELECT count(*) FROM information_schema.tables WHERE table_name = 'isolation_probe'").Scan(&tableCount)
	if err != nil {
		t.Fatalf("query second database: %v", err)
	}
	if tableCount != 0 {
		t.Fatal("expected second test database to be empty")
	}
}

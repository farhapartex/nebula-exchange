package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"os"
	"strings"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"nebula-exchange/backend/internal/ledger"
	"nebula-exchange/backend/internal/platform/money"
)

const usage = `usage:
  devtool credit-nc --email EMAIL --amount NC [--bucket card|crypto|earned_pending|earned]
  devtool grant-item --email EMAIL --item ITEM_ID --quantity N
  devtool complete-payment --id PAYMENT_ID`

func main() {
	if err := run(os.Args[1:]); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}

func run(arguments []string) error {
	if len(arguments) == 0 {
		return errors.New(usage)
	}
	databaseURL := os.Getenv("DATABASE_URL")
	if databaseURL == "" {
		return errors.New("DATABASE_URL is required")
	}
	ctx := context.Background()
	pool, err := pgxpool.New(ctx, databaseURL)
	if err != nil {
		return fmt.Errorf("connect database: %w", err)
	}
	defer pool.Close()

	switch arguments[0] {
	case "credit-nc":
		return creditNC(ctx, pool, arguments[1:])
	case "grant-item":
		return grantItem(ctx, pool, arguments[1:])
	case "complete-payment":
		return completePayment(ctx, pool, arguments[1:])
	default:
		return errors.New(usage)
	}
}

func creditNC(ctx context.Context, pool *pgxpool.Pool, arguments []string) error {
	flags := flag.NewFlagSet("credit-nc", flag.ContinueOnError)
	email := flags.String("email", "", "player email")
	amountText := flags.String("amount", "", "NC amount, up to 6 decimals")
	bucket := flags.String("bucket", string(ledger.BucketCard), "NC bucket")
	if err := flags.Parse(arguments); err != nil {
		return err
	}
	amount, err := money.ParseNC(*amountText)
	if err != nil || amount <= 0 {
		return fmt.Errorf("amount %q must be a positive NC value with up to 6 decimals", *amountText)
	}
	return creditPlayer(ctx, pool, *email, func(userID uuid.UUID) ledger.AccountKey {
		return ledger.PlayerNC(userID, ledger.Bucket(*bucket))
	}, int64(amount), fmt.Sprintf("%s NC to %s", *amountText, *bucket))
}

func grantItem(ctx context.Context, pool *pgxpool.Pool, arguments []string) error {
	flags := flag.NewFlagSet("grant-item", flag.ContinueOnError)
	email := flags.String("email", "", "player email")
	itemID := flags.Int("item", 0, "item id")
	quantity := flags.Int64("quantity", 1, "quantity")
	if err := flags.Parse(arguments); err != nil {
		return err
	}
	return creditPlayer(ctx, pool, *email, func(userID uuid.UUID) ledger.AccountKey {
		return ledger.PlayerItem(userID, *itemID)
	}, *quantity, fmt.Sprintf("%d of item %d", *quantity, *itemID))
}

func creditPlayer(ctx context.Context, pool *pgxpool.Pool, email string, accountFor func(uuid.UUID) ledger.AccountKey, amount int64, description string) error {
	return pgx.BeginFunc(ctx, pool, func(tx pgx.Tx) error {
		var userID uuid.UUID
		if err := tx.QueryRow(ctx, "SELECT id FROM users WHERE lower(email) = lower($1)", strings.TrimSpace(email)).Scan(&userID); err != nil {
			return fmt.Errorf("find player %q: %w", email, err)
		}
		journalID, err := ledger.DevCredit(ctx, tx, accountFor(userID), amount)
		if err != nil {
			return err
		}
		fmt.Printf("credited %s to %s (journal %s)\n", description, email, journalID)
		return nil
	})
}

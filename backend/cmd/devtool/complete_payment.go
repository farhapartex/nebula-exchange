package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"log/slog"
	"os"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"nebula-exchange/backend/internal/catalog"
	"nebula-exchange/backend/internal/payments"
	"nebula-exchange/backend/internal/platform/money"
	"nebula-exchange/backend/internal/purpose"
	"nebula-exchange/backend/internal/users"
)

func completePayment(ctx context.Context, pool *pgxpool.Pool, arguments []string) error {
	flags := flag.NewFlagSet("complete-payment", flag.ContinueOnError)
	paymentIDText := flags.String("id", "", "payment id")
	if err := flags.Parse(arguments); err != nil {
		return err
	}
	paymentID, err := uuid.Parse(*paymentIDText)
	if err != nil {
		return fmt.Errorf("payment id %q is not a UUID", *paymentIDText)
	}

	var amountMicro int64
	if err := pool.QueryRow(ctx, "SELECT amount_micro FROM payments WHERE id = $1", paymentID).Scan(&amountMicro); errors.Is(err, pgx.ErrNoRows) {
		return fmt.Errorf("payment %s does not exist", paymentID)
	} else if err != nil {
		return err
	}

	catalogService := catalog.NewService(catalog.NewLoader(pool), time.Minute, time.Now)
	toolLogger := slog.New(slog.NewTextHandler(os.Stderr, nil))
	settler := payments.NewSettler(pool, purpose.NewDefaultRunner(users.NewRepository(), catalogService, time.Now), toolLogger, time.Now)
	outcome, err := settler.Settle(ctx, payments.SettlementRequest{
		PaymentID: paymentID,
		Credited:  money.Micro(amountMicro),
		Event: payments.ExternalEvent{
			ID:       "development:" + paymentID.String(),
			Provider: "development",
			Type:     "checkout.completed",
			Payload:  map[string]any{"payment_id": paymentID.String()},
		},
	})
	if err != nil {
		return err
	}
	fmt.Printf("payment %s: %+v\n", paymentID, outcome)
	return nil
}

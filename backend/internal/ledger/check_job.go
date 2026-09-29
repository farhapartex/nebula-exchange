package ledger

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"

	"nebula-exchange/backend/internal/ledger/ledgerstore"
)

type CheckReport struct {
	UnbalancedJournals       []ledgerstore.FindUnbalancedJournalsRow
	BalancesNotMatchingEntry []ledgerstore.FindBalancesNotMatchingEntriesRow
	HeldNotMatchingHolds     []ledgerstore.FindHeldNotMatchingHoldsRow
}

func (report CheckReport) MismatchCount() int {
	return len(report.UnbalancedJournals) + len(report.BalancesNotMatchingEntry) + len(report.HeldNotMatchingHolds)
}

func RunIntegrityCheck(ctx context.Context, pool *pgxpool.Pool) (CheckReport, error) {
	queries := ledgerstore.New(pool)
	unbalancedJournals, err := queries.FindUnbalancedJournals(ctx)
	if err != nil {
		return CheckReport{}, fmt.Errorf("check journal sums: %w", err)
	}
	balancesNotMatchingEntries, err := queries.FindBalancesNotMatchingEntries(ctx)
	if err != nil {
		return CheckReport{}, fmt.Errorf("check balances against entries: %w", err)
	}
	heldNotMatchingHolds, err := queries.FindHeldNotMatchingHolds(ctx)
	if err != nil {
		return CheckReport{}, fmt.Errorf("check held against holds: %w", err)
	}
	return CheckReport{
		UnbalancedJournals:       unbalancedJournals,
		BalancesNotMatchingEntry: balancesNotMatchingEntries,
		HeldNotMatchingHolds:     heldNotMatchingHolds,
	}, nil
}

type CheckJob struct {
	pool   *pgxpool.Pool
	logger *slog.Logger
}

func NewCheckJob(pool *pgxpool.Pool, logger *slog.Logger) *CheckJob {
	return &CheckJob{pool: pool, logger: logger}
}

func (job *CheckJob) Name() string            { return "ledger_check" }
func (job *CheckJob) Interval() time.Duration { return time.Hour }

func (job *CheckJob) Run(ctx context.Context) error {
	report, err := RunIntegrityCheck(ctx, job.pool)
	if err != nil {
		return err
	}
	if report.MismatchCount() == 0 {
		return nil
	}
	job.logger.Error("ledger integrity mismatch",
		slog.Bool("alert", true),
		slog.Any("unbalanced_journals", report.UnbalancedJournals),
		slog.Any("balances_not_matching_entries", report.BalancesNotMatchingEntry),
		slog.Any("held_not_matching_holds", report.HeldNotMatchingHolds),
	)
	return fmt.Errorf("ledger check found %d mismatches", report.MismatchCount())
}

package ledger

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"slices"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"

	"nebula-exchange/backend/internal/ledger/ledgerstore"
)

type HoldCapture struct {
	HoldID uuid.UUID
	Amount int64
}

type balanceChange struct {
	availableChange int64
	heldTaken       int64
}

func Post(ctx context.Context, tx pgx.Tx, journal Journal, captures ...HoldCapture) (uuid.UUID, error) {
	queries := ledgerstore.New(tx)

	entries, err := resolveLegs(ctx, queries, journal.Legs)
	if err != nil {
		return uuid.Nil, err
	}
	captureEntries, err := resolveCaptures(ctx, queries, captures)
	if err != nil {
		return uuid.Nil, err
	}
	entries = append(entries, captureEntries...)
	if len(entries) < 2 {
		return uuid.Nil, fmt.Errorf("%w: a journal needs at least two entries", ErrUnbalancedJournal)
	}

	assetsByAccount, err := loadAccountAssets(ctx, queries, entries)
	if err != nil {
		return uuid.Nil, err
	}
	if err := ensureBalancedPerAsset(entries, assetsByAccount); err != nil {
		return uuid.Nil, err
	}

	journalID, err := insertJournal(ctx, queries, journal)
	if err != nil {
		return uuid.Nil, err
	}
	if err := reduceCapturedHolds(ctx, queries, captures); err != nil {
		return uuid.Nil, err
	}
	if err := applyBalanceChanges(ctx, queries, entries, assetsByAccount, journal.Type == JournalDisputeDebit); err != nil {
		return uuid.Nil, err
	}

	entryRows := make([]ledgerstore.InsertEntriesParams, 0, len(entries))
	for _, entry := range entries {
		entryRows = append(entryRows, ledgerstore.InsertEntriesParams{JournalID: journalID, AccountID: entry.accountID, Amount: entry.amount})
	}
	if _, err := queries.InsertEntries(ctx, entryRows); err != nil {
		return uuid.Nil, fmt.Errorf("insert ledger entries: %w", err)
	}
	return journalID, nil
}

type resolvedEntry struct {
	accountID    int64
	amount       int64
	isHeldCharge bool
}

func resolveLegs(ctx context.Context, queries *ledgerstore.Queries, legs []Leg) ([]resolvedEntry, error) {
	entries := make([]resolvedEntry, 0, len(legs))
	for _, leg := range legs {
		if leg.Amount == 0 {
			return nil, fmt.Errorf("%w: zero-amount leg on %+v", ErrInvalidAmount, leg.Account)
		}
		accountID, err := resolveAccount(ctx, queries, leg.Account)
		if err != nil {
			return nil, err
		}
		entries = append(entries, resolvedEntry{accountID: accountID, amount: leg.Amount})
	}
	return entries, nil
}

func resolveCaptures(ctx context.Context, queries *ledgerstore.Queries, captures []HoldCapture) ([]resolvedEntry, error) {
	entries := make([]resolvedEntry, 0, len(captures))
	for _, capture := range captures {
		if capture.Amount <= 0 {
			return nil, ErrInvalidAmount
		}
		hold, err := queries.GetHold(ctx, capture.HoldID)
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrHoldNotAvailable
		}
		if err != nil {
			return nil, fmt.Errorf("load hold: %w", err)
		}
		entries = append(entries, resolvedEntry{accountID: hold.AccountID, amount: -capture.Amount, isHeldCharge: true})
	}
	return entries, nil
}

func loadAccountAssets(ctx context.Context, queries *ledgerstore.Queries, entries []resolvedEntry) (map[int64]Asset, error) {
	accountIDs := make([]int64, 0, len(entries))
	for _, entry := range entries {
		accountIDs = append(accountIDs, entry.accountID)
	}
	accountRows, err := queries.ListAccountsByID(ctx, accountIDs)
	if err != nil {
		return nil, fmt.Errorf("load ledger accounts: %w", err)
	}
	assetsByAccount := make(map[int64]Asset, len(accountRows))
	for _, accountRow := range accountRows {
		asset := NC
		if accountRow.ItemID != nil {
			asset = Item(int(*accountRow.ItemID))
		}
		assetsByAccount[accountRow.ID] = asset
	}
	return assetsByAccount, nil
}

func ensureBalancedPerAsset(entries []resolvedEntry, assetsByAccount map[int64]Asset) error {
	totalsByAsset := map[Asset]int64{}
	for _, entry := range entries {
		totalsByAsset[assetsByAccount[entry.accountID]] += entry.amount
	}
	for asset, total := range totalsByAsset {
		if total != 0 {
			return fmt.Errorf("%w: asset %d is off by %d", ErrUnbalancedJournal, asset, total)
		}
	}
	return nil
}

func insertJournal(ctx context.Context, queries *ledgerstore.Queries, journal Journal) (uuid.UUID, error) {
	journalID, err := uuid.NewV7()
	if err != nil {
		return uuid.Nil, fmt.Errorf("generate journal id: %w", err)
	}
	metadata := journal.Metadata
	if metadata == nil {
		metadata = map[string]any{}
	}
	encodedMetadata, err := json.Marshal(metadata)
	if err != nil {
		return uuid.Nil, fmt.Errorf("encode journal metadata: %w", err)
	}
	insertedID, err := queries.InsertJournal(ctx, ledgerstore.InsertJournalParams{
		ID:       journalID,
		Type:     string(journal.Type),
		RefType:  journal.Reference.Type,
		RefID:    journal.Reference.ID,
		Metadata: encodedMetadata,
	})
	if errors.Is(err, pgx.ErrNoRows) {
		return uuid.Nil, ErrDuplicateJournal
	}
	if err != nil {
		return uuid.Nil, fmt.Errorf("insert journal: %w", err)
	}
	return insertedID, nil
}

func reduceCapturedHolds(ctx context.Context, queries *ledgerstore.Queries, captures []HoldCapture) error {
	sortedCaptures := slices.Clone(captures)
	slices.SortFunc(sortedCaptures, func(first, second HoldCapture) int {
		return slices.Compare(first.HoldID[:], second.HoldID[:])
	})
	for _, capture := range sortedCaptures {
		_, err := queries.ReduceHold(ctx, ledgerstore.ReduceHoldParams{ID: capture.HoldID, Amount: capture.Amount, FinalStatus: "CAPTURED"})
		if errors.Is(err, pgx.ErrNoRows) {
			return ErrHoldNotAvailable
		}
		if err != nil {
			return fmt.Errorf("capture hold: %w", err)
		}
	}
	return nil
}

func applyBalanceChanges(ctx context.Context, queries *ledgerstore.Queries, entries []resolvedEntry, assetsByAccount map[int64]Asset, allowsDisputeOverdraft bool) error {
	changesByAccount := map[int64]*balanceChange{}
	for _, entry := range entries {
		change, exists := changesByAccount[entry.accountID]
		if !exists {
			change = &balanceChange{}
			changesByAccount[entry.accountID] = change
		}
		if entry.isHeldCharge {
			change.heldTaken += -entry.amount
		} else {
			change.availableChange += entry.amount
		}
	}

	accountIDs := make([]int64, 0, len(changesByAccount))
	for accountID := range changesByAccount {
		accountIDs = append(accountIDs, accountID)
	}
	slices.Sort(accountIDs)

	for _, accountID := range accountIDs {
		change := changesByAccount[accountID]
		if change.heldTaken > 0 {
			_, err := queries.TakeFromHeld(ctx, ledgerstore.TakeFromHeldParams{AccountID: accountID, Amount: change.heldTaken})
			if errors.Is(err, pgx.ErrNoRows) {
				return ErrHoldNotAvailable
			}
			if err != nil {
				return fmt.Errorf("take held balance: %w", err)
			}
		}
		if change.availableChange == 0 {
			continue
		}
		_, err := queries.ApplyAvailableChange(ctx, ledgerstore.ApplyAvailableChangeParams{
			AccountID:              accountID,
			Change:                 change.availableChange,
			AllowsDisputeOverdraft: allowsDisputeOverdraft,
		})
		if errors.Is(err, pgx.ErrNoRows) {
			return &InsufficientBalanceError{Asset: assetsByAccount[accountID]}
		}
		if err != nil {
			return fmt.Errorf("apply balance change: %w", err)
		}
	}
	return nil
}

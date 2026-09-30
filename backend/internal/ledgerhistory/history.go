package ledgerhistory

import (
	"context"
	"fmt"
	"strconv"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"

	"nebula-exchange/backend/internal/ledger/ledgerstore"
	"nebula-exchange/backend/internal/platform/pagination"
)

type JournalReference struct {
	Type string `json:"type"`
	ID   string `json:"id"`
}

type JournalEntry struct {
	ItemID *int    `json:"item_id"`
	Bucket *string `json:"bucket"`
	Amount string  `json:"amount"`
}

type JournalRecord struct {
	ID        uuid.UUID        `json:"id"`
	Type      string           `json:"type"`
	Reference JournalReference `json:"reference"`
	CreatedAt time.Time        `json:"created_at"`
	Entries   []JournalEntry   `json:"entries"`
}

type Cursor struct {
	BeforeJournalID uuid.UUID `json:"before_journal_id"`
}

type Filter struct {
	JournalType string
}

type Reader struct {
	pool *pgxpool.Pool
}

func NewReader(pool *pgxpool.Pool) *Reader {
	return &Reader{pool: pool}
}

func (reader *Reader) List(ctx context.Context, userID uuid.UUID, filter Filter, pageRequest pagination.Request, cursor *Cursor) (pagination.Page[JournalRecord], error) {
	queries := ledgerstore.New(reader.pool)
	parameters := ledgerstore.ListPlayerJournalsParams{UserID: &userID, RowLimit: int32(pageRequest.FetchLimit())}
	if cursor != nil {
		parameters.BeforeJournalID = &cursor.BeforeJournalID
	}
	if filter.JournalType != "" {
		parameters.JournalType = &filter.JournalType
	}
	journalRows, err := queries.ListPlayerJournals(ctx, parameters)
	if err != nil {
		return pagination.Page[JournalRecord]{}, fmt.Errorf("list player journals: %w", err)
	}

	records := make([]JournalRecord, 0, len(journalRows))
	recordIndexByJournal := make(map[uuid.UUID]int, len(journalRows))
	journalIDs := make([]uuid.UUID, 0, len(journalRows))
	for _, journalRow := range journalRows {
		recordIndexByJournal[journalRow.ID] = len(records)
		journalIDs = append(journalIDs, journalRow.ID)
		records = append(records, JournalRecord{
			ID:        journalRow.ID,
			Type:      journalRow.Type,
			Reference: JournalReference{Type: journalRow.RefType, ID: journalRow.RefID},
			CreatedAt: journalRow.CreatedAt,
			Entries:   []JournalEntry{},
		})
	}

	if len(journalIDs) > 0 {
		entryRows, err := queries.ListPlayerJournalEntries(ctx, ledgerstore.ListPlayerJournalEntriesParams{JournalIds: journalIDs, UserID: &userID})
		if err != nil {
			return pagination.Page[JournalRecord]{}, fmt.Errorf("list player journal entries: %w", err)
		}
		for _, entryRow := range entryRows {
			entry := JournalEntry{Bucket: entryRow.Bucket, Amount: strconv.FormatInt(entryRow.Amount, 10)}
			if entryRow.ItemID != nil {
				itemID := int(*entryRow.ItemID)
				entry.ItemID = &itemID
			}
			recordIndex := recordIndexByJournal[entryRow.JournalID]
			records[recordIndex].Entries = append(records[recordIndex].Entries, entry)
		}
	}

	return pagination.BuildPage(records, pageRequest, func(record JournalRecord) Cursor {
		return Cursor{BeforeJournalID: record.ID}
	})
}

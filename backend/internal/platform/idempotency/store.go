package idempotency

import (
	"context"
	"errors"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"

	"nebula-exchange/backend/internal/platform/idempotency/idempotencystore"
)

type StoredResponse struct {
	StatusCode  int
	ContentType string
	Body        []byte
}

type Record struct {
	RequestHash []byte
	IsCompleted bool
	Response    StoredResponse
	CreatedAt   time.Time
}

type Store interface {
	Begin(ctx context.Context, scope, key string, requestHash []byte) (bool, error)
	Find(ctx context.Context, scope, key string) (Record, bool, error)
	Complete(ctx context.Context, scope, key string, response StoredResponse) error
	Release(ctx context.Context, scope, key string) error
	DeleteIfCreatedBefore(ctx context.Context, scope, key string, cutoff time.Time) error
}

type PostgresStore struct {
	queries idempotencystore.Querier
}

func NewPostgresStore(database idempotencystore.DBTX) *PostgresStore {
	return &PostgresStore{queries: idempotencystore.New(database)}
}

func (store *PostgresStore) Begin(ctx context.Context, scope, key string, requestHash []byte) (bool, error) {
	insertedRows, err := store.queries.InsertInProgressKey(ctx, idempotencystore.InsertInProgressKeyParams{
		Scope:          scope,
		IdempotencyKey: key,
		RequestHash:    requestHash,
	})
	return insertedRows == 1, err
}

func (store *PostgresStore) Find(ctx context.Context, scope, key string) (Record, bool, error) {
	storedKey, err := store.queries.FindKey(ctx, idempotencystore.FindKeyParams{Scope: scope, IdempotencyKey: key})
	if errors.Is(err, pgx.ErrNoRows) {
		return Record{}, false, nil
	}
	if err != nil {
		return Record{}, false, err
	}
	return Record{
		RequestHash: storedKey.RequestHash,
		IsCompleted: storedKey.Status == "completed",
		Response: StoredResponse{
			StatusCode:  int(storedKey.ResponseStatusCode.Int32),
			ContentType: storedKey.ResponseContentType.String,
			Body:        storedKey.ResponseBody,
		},
		CreatedAt: storedKey.CreatedAt.Time,
	}, true, nil
}

func (store *PostgresStore) Complete(ctx context.Context, scope, key string, response StoredResponse) error {
	return store.queries.CompleteKey(ctx, idempotencystore.CompleteKeyParams{
		Scope:               scope,
		IdempotencyKey:      key,
		ResponseStatusCode:  pgtype.Int4{Int32: int32(response.StatusCode), Valid: true},
		ResponseContentType: pgtype.Text{String: response.ContentType, Valid: response.ContentType != ""},
		ResponseBody:        response.Body,
	})
}

func (store *PostgresStore) Release(ctx context.Context, scope, key string) error {
	return store.queries.DeleteKey(ctx, idempotencystore.DeleteKeyParams{Scope: scope, IdempotencyKey: key})
}

func (store *PostgresStore) DeleteIfCreatedBefore(ctx context.Context, scope, key string, cutoff time.Time) error {
	_, err := store.queries.DeleteKeyCreatedBefore(ctx, idempotencystore.DeleteKeyCreatedBeforeParams{
		Scope:          scope,
		IdempotencyKey: key,
		CreatedAt:      pgtype.Timestamptz{Time: cutoff, Valid: true},
	})
	return err
}

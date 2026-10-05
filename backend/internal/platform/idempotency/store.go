package idempotency

import (
	"context"
	"errors"
	"time"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
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

type GormStore struct {
	database *gorm.DB
}

func NewGormStore(database *gorm.DB) *GormStore {
	return &GormStore{database: database}
}

func (store *GormStore) Begin(ctx context.Context, scope, key string, requestHash []byte) (bool, error) {
	result := store.database.WithContext(ctx).
		Clauses(clause.OnConflict{DoNothing: true}).
		Create(&idempotencyKeyRecord{Scope: scope, IdempotencyKey: key, RequestHash: requestHash, Status: statusInProgress})
	return result.RowsAffected == 1, result.Error
}

func (store *GormStore) Find(ctx context.Context, scope, key string) (Record, bool, error) {
	var storedKey idempotencyKeyRecord
	err := store.database.WithContext(ctx).
		Where(map[string]any{"scope": scope, "idempotency_key": key}).
		Take(&storedKey).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return Record{}, false, nil
	}
	if err != nil {
		return Record{}, false, err
	}
	return storedKey.toRecord(), true, nil
}

func (store *GormStore) Complete(ctx context.Context, scope, key string, response StoredResponse) error {
	completedAt := time.Now().UTC()
	statusCode := response.StatusCode
	return store.database.WithContext(ctx).
		Model(&idempotencyKeyRecord{}).
		Where(map[string]any{"scope": scope, "idempotency_key": key}).
		Updates(idempotencyKeyRecord{
			Status:              statusCompleted,
			ResponseStatusCode:  &statusCode,
			ResponseContentType: response.ContentType,
			ResponseBody:        response.Body,
			CompletedAt:         &completedAt,
		}).Error
}

func (store *GormStore) Release(ctx context.Context, scope, key string) error {
	return store.database.WithContext(ctx).
		Where(map[string]any{"scope": scope, "idempotency_key": key}).
		Delete(&idempotencyKeyRecord{}).Error
}

func (store *GormStore) DeleteIfCreatedBefore(ctx context.Context, scope, key string, cutoff time.Time) error {
	return store.database.WithContext(ctx).
		Where(map[string]any{"scope": scope, "idempotency_key": key}).
		Where("created_at < ?", cutoff).
		Delete(&idempotencyKeyRecord{}).Error
}

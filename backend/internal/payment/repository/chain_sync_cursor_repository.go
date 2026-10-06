package repository

import (
	"context"
	"errors"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	"github.com/farhapartex/nebula-exchange/backend/internal/payment/models"
	"github.com/farhapartex/nebula-exchange/backend/internal/platform/database"
)

var ErrChainSyncCursorNotFound = errors.New("chain sync cursor not found")

type ChainSyncCursorRepository interface {
	Find(ctx context.Context, name string) (models.ChainSyncCursor, error)
	Save(ctx context.Context, cursor *models.ChainSyncCursor) error
}

type GormChainSyncCursorRepository struct {
	database *gorm.DB
}

func NewChainSyncCursorRepository(database *gorm.DB) *GormChainSyncCursorRepository {
	return &GormChainSyncCursorRepository{database: database}
}

func (repository *GormChainSyncCursorRepository) Find(ctx context.Context, name string) (models.ChainSyncCursor, error) {
	var cursor models.ChainSyncCursor
	err := database.Session(ctx, repository.database).Where(map[string]any{"name": name}).Take(&cursor).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return models.ChainSyncCursor{}, ErrChainSyncCursorNotFound
	}
	return cursor, err
}

func (repository *GormChainSyncCursorRepository) Save(ctx context.Context, cursor *models.ChainSyncCursor) error {
	return database.Session(ctx, repository.database).
		Clauses(clause.OnConflict{
			Columns:   []clause.Column{{Name: "name"}},
			DoUpdates: clause.AssignmentColumns([]string{"chain_id", "block_number", "block_hash", "updated_at"}),
		}).
		Create(cursor).Error
}

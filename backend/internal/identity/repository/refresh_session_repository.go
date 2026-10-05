package repository

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	"github.com/farhapartex/nebula-exchange/backend/internal/identity/models"
	"github.com/farhapartex/nebula-exchange/backend/internal/platform/database"
)

type RefreshSessionRepository interface {
	Create(ctx context.Context, refreshSession *models.RefreshSession) error
	FindByHash(ctx context.Context, tokenHash []byte) (models.RefreshSession, bool, error)
	Consume(ctx context.Context, tokenHash []byte, now time.Time) (models.RefreshSession, bool, error)
	LinkReplacement(ctx context.Context, sessionID, replacementID uuid.UUID) error
	RevokeByHash(ctx context.Context, tokenHash []byte, now time.Time) error
	RevokeFamily(ctx context.Context, familyID uuid.UUID, now time.Time) (int64, error)
}

type GormRefreshSessionRepository struct {
	database *gorm.DB
}

func NewRefreshSessionRepository(database *gorm.DB) *GormRefreshSessionRepository {
	return &GormRefreshSessionRepository{database: database}
}

func (repository *GormRefreshSessionRepository) Create(ctx context.Context, refreshSession *models.RefreshSession) error {
	return database.Session(ctx, repository.database).Create(refreshSession).Error
}

func (repository *GormRefreshSessionRepository) FindByHash(ctx context.Context, tokenHash []byte) (models.RefreshSession, bool, error) {
	var storedSession models.RefreshSession
	err := database.Session(ctx, repository.database).Where("token_hash = ?", tokenHash).Take(&storedSession).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return models.RefreshSession{}, false, nil
	}
	if err != nil {
		return models.RefreshSession{}, false, err
	}
	return storedSession, true, nil
}

func (repository *GormRefreshSessionRepository) Consume(ctx context.Context, tokenHash []byte, now time.Time) (models.RefreshSession, bool, error) {
	var consumedSessions []models.RefreshSession
	err := database.Session(ctx, repository.database).
		Model(&consumedSessions).
		Clauses(clause.Returning{}).
		Where("token_hash = ?", tokenHash).
		Where("revoked_at IS NULL AND expires_at > ?", now).
		Update("revoked_at", now).Error
	if err != nil || len(consumedSessions) == 0 {
		return models.RefreshSession{}, false, err
	}
	return consumedSessions[0], true, nil
}

func (repository *GormRefreshSessionRepository) LinkReplacement(ctx context.Context, sessionID, replacementID uuid.UUID) error {
	return database.Session(ctx, repository.database).
		Model(&models.RefreshSession{}).
		Where(map[string]any{"id": sessionID}).
		Update("replaced_by_id", replacementID).Error
}

func (repository *GormRefreshSessionRepository) RevokeByHash(ctx context.Context, tokenHash []byte, now time.Time) error {
	return database.Session(ctx, repository.database).
		Model(&models.RefreshSession{}).
		Where("token_hash = ?", tokenHash).
		Where("revoked_at IS NULL").
		Update("revoked_at", now).Error
}

func (repository *GormRefreshSessionRepository) RevokeFamily(ctx context.Context, familyID uuid.UUID, now time.Time) (int64, error) {
	result := database.Session(ctx, repository.database).
		Model(&models.RefreshSession{}).
		Where(map[string]any{"family_id": familyID}).
		Where("revoked_at IS NULL").
		Update("revoked_at", now)
	return result.RowsAffected, result.Error
}

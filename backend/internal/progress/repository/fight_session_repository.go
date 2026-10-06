package repository

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"github.com/farhapartex/nebula-exchange/backend/internal/platform/database"
	"github.com/farhapartex/nebula-exchange/backend/internal/progress/models"
)

const oneOpenSessionPerLevelIndex = "fight_sessions_one_open_per_level"

var ErrFightAlreadyStarting = errors.New("another fight for this level started at the same time")

type FightSessionRepository interface {
	AbandonOpen(ctx context.Context, userID uuid.UUID, levelID string, abandonedAt time.Time) (int64, error)
	Create(ctx context.Context, fightSession *models.FightSession) error
}

type GormFightSessionRepository struct {
	database *gorm.DB
}

func NewFightSessionRepository(database *gorm.DB) *GormFightSessionRepository {
	return &GormFightSessionRepository{database: database}
}

func (repository *GormFightSessionRepository) AbandonOpen(ctx context.Context, userID uuid.UUID, levelID string, abandonedAt time.Time) (int64, error) {
	result := database.Session(ctx, repository.database).
		Model(&models.FightSession{}).
		Where(map[string]any{"user_id": userID, "level_id": levelID, "status": models.FightStatusStarted}).
		Updates(map[string]any{"status": models.FightStatusAbandoned, "finished_at": abandonedAt, "updated_at": abandonedAt})
	return result.RowsAffected, result.Error
}

func (repository *GormFightSessionRepository) Create(ctx context.Context, fightSession *models.FightSession) error {
	err := database.Session(ctx, repository.database).Create(fightSession).Error
	if constraintName, isUniqueViolation := database.ViolatedUniqueConstraint(err); isUniqueViolation && constraintName == oneOpenSessionPerLevelIndex {
		return ErrFightAlreadyStarting
	}
	return err
}

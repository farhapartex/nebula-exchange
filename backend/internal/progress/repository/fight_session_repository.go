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

type FinishedFight struct {
	Outcome     models.FightOutcome
	Stars       int
	DurationMS  int
	DamageDealt int
	DamageTaken int
	FinishedAt  time.Time
}

type WonFightPosition struct {
	FinishedAt time.Time
	ID         uuid.UUID
}

type FightSessionRepository interface {
	AbandonOpen(ctx context.Context, userID uuid.UUID, levelID string, abandonedAt time.Time) (int64, error)
	Create(ctx context.Context, fightSession *models.FightSession) error
	FindForUser(ctx context.Context, fightSessionID, userID uuid.UUID) (models.FightSession, bool, error)
	Finish(ctx context.Context, fightSessionID uuid.UUID, finishedFight FinishedFight) (bool, error)
	Reject(ctx context.Context, fightSessionID uuid.UUID, reason string, rejectedAt time.Time) (bool, error)
	CountOutcomes(ctx context.Context, userID uuid.UUID, levelIDs []string) (int, int, error)
	ListWonInLevels(ctx context.Context, userID uuid.UUID, levelIDs []string, after *WonFightPosition, limit int) ([]models.FightSession, error)
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

func (repository *GormFightSessionRepository) FindForUser(ctx context.Context, fightSessionID, userID uuid.UUID) (models.FightSession, bool, error) {
	var fightSession models.FightSession
	err := database.Session(ctx, repository.database).
		Where(map[string]any{"id": fightSessionID, "user_id": userID}).
		Take(&fightSession).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return models.FightSession{}, false, nil
	}
	if err != nil {
		return models.FightSession{}, false, err
	}
	return fightSession, true, nil
}

func (repository *GormFightSessionRepository) Finish(ctx context.Context, fightSessionID uuid.UUID, finishedFight FinishedFight) (bool, error) {
	noReward := 0
	result := database.Session(ctx, repository.database).
		Model(&models.FightSession{}).
		Where(map[string]any{"id": fightSessionID, "status": models.FightStatusStarted}).
		Updates(map[string]any{
			"status":            models.FightStatusFinished,
			"outcome":           finishedFight.Outcome,
			"stars":             finishedFight.Stars,
			"duration_ms":       finishedFight.DurationMS,
			"damage_dealt":      finishedFight.DamageDealt,
			"damage_taken":      finishedFight.DamageTaken,
			"reward_coins":      int64(noReward),
			"reward_experience": noReward,
			"finished_at":       finishedFight.FinishedAt,
			"updated_at":        finishedFight.FinishedAt,
		})
	return result.RowsAffected == 1, result.Error
}

func (repository *GormFightSessionRepository) Reject(ctx context.Context, fightSessionID uuid.UUID, reason string, rejectedAt time.Time) (bool, error) {
	result := database.Session(ctx, repository.database).
		Model(&models.FightSession{}).
		Where(map[string]any{"id": fightSessionID, "status": models.FightStatusStarted}).
		Updates(map[string]any{"status": models.FightStatusRejected, "rejection_reason": reason, "finished_at": rejectedAt, "updated_at": rejectedAt})
	return result.RowsAffected == 1, result.Error
}

func (repository *GormFightSessionRepository) CountOutcomes(ctx context.Context, userID uuid.UUID, levelIDs []string) (int, int, error) {
	if len(levelIDs) == 0 {
		return 0, 0, nil
	}
	var outcomeCounts []struct {
		Outcome models.FightOutcome
		Total   int
	}
	err := database.Session(ctx, repository.database).
		Model(&models.FightSession{}).
		Select("outcome, count(*) AS total").
		Where(map[string]any{"user_id": userID, "status": models.FightStatusFinished, "level_id": levelIDs}).
		Group("outcome").
		Scan(&outcomeCounts).Error
	if err != nil {
		return 0, 0, err
	}
	wins, losses := 0, 0
	for _, outcomeCount := range outcomeCounts {
		switch outcomeCount.Outcome {
		case models.FightOutcomeWon:
			wins = outcomeCount.Total
		case models.FightOutcomeLost:
			losses = outcomeCount.Total
		}
	}
	return wins, losses, nil
}

func (repository *GormFightSessionRepository) ListWonInLevels(ctx context.Context, userID uuid.UUID, levelIDs []string, after *WonFightPosition, limit int) ([]models.FightSession, error) {
	if len(levelIDs) == 0 || limit < 1 {
		return nil, nil
	}
	query := database.Session(ctx, repository.database).
		Where(map[string]any{"user_id": userID, "status": models.FightStatusFinished, "outcome": models.FightOutcomeWon, "level_id": levelIDs})
	if after != nil {
		query = query.Where("(finished_at, id) < (?, ?)", after.FinishedAt, after.ID)
	}
	var wonFights []models.FightSession
	err := query.Order("finished_at DESC").Order("id DESC").Limit(limit).Find(&wonFights).Error
	return wonFights, err
}

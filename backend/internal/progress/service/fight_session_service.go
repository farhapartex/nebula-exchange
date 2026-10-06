package service

import (
	"context"
	"crypto/rand"
	"encoding/binary"
	"errors"
	"log/slog"
	"net/http"
	"time"

	"github.com/google/uuid"

	"github.com/farhapartex/nebula-exchange/backend/internal/platform/apierror"
	"github.com/farhapartex/nebula-exchange/backend/internal/platform/database"
	"github.com/farhapartex/nebula-exchange/backend/internal/progress/models"
	"github.com/farhapartex/nebula-exchange/backend/internal/progress/repository"
)

var (
	ErrLevelLocked          = apierror.New(http.StatusForbidden, apierror.CodeLevelLocked, "Win the previous level to unlock this one")
	ErrFightAlreadyStarting = apierror.Conflict("This level is already being started. Try again in a moment.")
)

type StartedFight struct {
	ID        uuid.UUID
	LevelID   string
	Status    models.FightStatus
	Seed      int64
	StartedAt time.Time
}

type FightSessionService interface {
	Start(ctx context.Context, userID uuid.UUID, levelID string) (StartedFight, error)
}

type FightSessionDependencies struct {
	Levels        LevelCatalog
	LevelProgress repository.LevelProgressRepository
	FightSessions repository.FightSessionRepository
	Transactions  TransactionRunner
	Logger        *slog.Logger
	Now           Clock
}

type fightSessionService struct {
	dependencies FightSessionDependencies
}

func NewFightSessionService(dependencies FightSessionDependencies) FightSessionService {
	return &fightSessionService{dependencies: dependencies}
}

func (fights *fightSessionService) Start(ctx context.Context, userID uuid.UUID, levelID string) (StartedFight, error) {
	placement, err := fights.dependencies.Levels.PlayableLevel(ctx, levelID)
	if err != nil {
		return StartedFight{}, err
	}
	if placement.PreviousLevelID != nil {
		isPreviousLevelWon, err := fights.dependencies.LevelProgress.IsCompleted(ctx, userID, *placement.PreviousLevelID)
		if err != nil {
			return StartedFight{}, err
		}
		if !isPreviousLevelWon {
			return StartedFight{}, ErrLevelLocked
		}
	}

	newFight, err := fights.newFightSession(userID, levelID)
	if err != nil {
		return StartedFight{}, err
	}
	err = fights.dependencies.Transactions.WithinTransaction(ctx, func(ctx context.Context) error {
		abandonedCount, err := fights.dependencies.FightSessions.AbandonOpen(ctx, userID, levelID, newFight.StartedAt)
		if err != nil {
			return err
		}
		if abandonedCount > 0 {
			fights.dependencies.Logger.InfoContext(ctx, "unfinished fight abandoned", slog.String("user_id", userID.String()), slog.String("level_id", levelID))
		}
		if err := fights.dependencies.FightSessions.Create(ctx, &newFight); err != nil {
			return err
		}
		return fights.dependencies.LevelProgress.RecordStart(ctx, userID, levelID, newFight.StartedAt)
	})
	if errors.Is(err, repository.ErrFightAlreadyStarting) {
		return StartedFight{}, ErrFightAlreadyStarting
	}
	if err != nil {
		return StartedFight{}, err
	}
	return StartedFight{ID: newFight.ID, LevelID: levelID, Status: newFight.Status, Seed: newFight.Seed, StartedAt: newFight.StartedAt}, nil
}

func (fights *fightSessionService) newFightSession(userID uuid.UUID, levelID string) (models.FightSession, error) {
	fightID, err := uuid.NewV7()
	if err != nil {
		return models.FightSession{}, err
	}
	fightSeed, err := randomSeed()
	if err != nil {
		return models.FightSession{}, err
	}
	startedAt := fights.dependencies.Now().UTC()
	return models.FightSession{
		ID:        fightID,
		UserID:    userID,
		LevelID:   levelID,
		Status:    models.FightStatusStarted,
		Seed:      fightSeed,
		Loadout:   database.JSONDocument(`{}`),
		StartedAt: startedAt,
		CreatedAt: startedAt,
		UpdatedAt: startedAt,
	}, nil
}

func randomSeed() (int64, error) {
	var seedBytes [8]byte
	if _, err := rand.Read(seedBytes[:]); err != nil {
		return 0, err
	}
	return int64(binary.BigEndian.Uint64(seedBytes[:]) >> 1), nil
}

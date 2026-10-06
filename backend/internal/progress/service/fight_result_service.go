package service

import (
	"context"
	"log/slog"
	"net/http"
	"time"

	"github.com/google/uuid"

	"github.com/farhapartex/nebula-exchange/backend/internal/platform/apierror"
	"github.com/farhapartex/nebula-exchange/backend/internal/progress/models"
	"github.com/farhapartex/nebula-exchange/backend/internal/progress/repository"
)

var (
	ErrFightNotFound     = apierror.NotFound("This fight does not exist")
	ErrFightAlreadyEnded = apierror.Conflict("This fight has already ended")
)

type FightResult struct {
	FightSessionID   uuid.UUID
	LevelID          string
	Status           models.FightStatus
	Outcome          models.FightOutcome
	Stars            int
	DurationMS       int
	DamageDealt      int
	DamageTaken      int
	IsLevelCompleted bool
	FinishedAt       time.Time
}

type FightResultService interface {
	Submit(ctx context.Context, userID, fightSessionID uuid.UUID, report FightReport) (FightResult, error)
}

type FightResultDependencies struct {
	Levels        LevelCatalog
	FightSessions repository.FightSessionRepository
	Fighters      repository.FighterRepository
	LevelProgress repository.LevelProgressRepository
	Transactions  TransactionRunner
	Logger        *slog.Logger
	Now           Clock
}

type fightResultService struct {
	dependencies FightResultDependencies
}

func NewFightResultService(dependencies FightResultDependencies) FightResultService {
	return &fightResultService{dependencies: dependencies}
}

func (results *fightResultService) Submit(ctx context.Context, userID, fightSessionID uuid.UUID, report FightReport) (FightResult, error) {
	fightSession, isFound, err := results.dependencies.FightSessions.FindForUser(ctx, fightSessionID, userID)
	if err != nil {
		return FightResult{}, err
	}
	if !isFound {
		return FightResult{}, ErrFightNotFound
	}
	if fightSession.Status != models.FightStatusStarted {
		return FightResult{}, ErrFightAlreadyEnded
	}

	fightContent, err := results.dependencies.Levels.FightContent(ctx, fightSession.LevelID)
	if err != nil {
		return FightResult{}, err
	}
	reportContext, err := results.buildReportContext(ctx, userID, fightSession, fightContent.TimeLimitSeconds, fightContent.Waves[0].Enemy.Stats)
	if err != nil {
		return FightResult{}, err
	}
	if problem := findReportProblem(report, reportContext); problem != "" {
		return FightResult{}, results.reject(ctx, fightSession, problem, reportContext.ReportedAt)
	}

	hasWon := report.Outcome == models.FightOutcomeWon
	stars := 0
	if hasWon {
		stars = fightContent.StarRules.StarsForWin(float64(report.DurationMS) / 1000)
	}
	finishedFight := repository.FinishedFight{
		Outcome:     report.Outcome,
		Stars:       stars,
		DurationMS:  report.DurationMS,
		DamageDealt: report.DamageDealt,
		DamageTaken: report.DamageTaken,
		FinishedAt:  reportContext.ReportedAt,
	}
	err = results.dependencies.Transactions.WithinTransaction(ctx, func(ctx context.Context) error {
		isFinished, err := results.dependencies.FightSessions.Finish(ctx, fightSession.ID, finishedFight)
		if err != nil {
			return err
		}
		if !isFinished {
			return ErrFightAlreadyEnded
		}
		if err := results.dependencies.Fighters.RecordResult(ctx, userID, hasWon, finishedFight.FinishedAt); err != nil {
			return err
		}
		if !hasWon {
			return nil
		}
		return results.dependencies.LevelProgress.RecordWin(ctx, userID, fightSession.LevelID, stars, finishedFight.FinishedAt)
	})
	if err != nil {
		return FightResult{}, err
	}
	return FightResult{
		FightSessionID:   fightSession.ID,
		LevelID:          fightSession.LevelID,
		Status:           models.FightStatusFinished,
		Outcome:          report.Outcome,
		Stars:            stars,
		DurationMS:       report.DurationMS,
		DamageDealt:      report.DamageDealt,
		DamageTaken:      report.DamageTaken,
		IsLevelCompleted: hasWon,
		FinishedAt:       finishedFight.FinishedAt,
	}, nil
}

func (results *fightResultService) buildReportContext(ctx context.Context, userID uuid.UUID, fightSession models.FightSession, timeLimitSeconds int, enemyStatsDocument []byte) (fightReportContext, error) {
	profile, hasProfile, err := results.dependencies.Fighters.FindProfile(ctx, userID)
	if err != nil {
		return fightReportContext{}, err
	}
	if !hasProfile {
		return fightReportContext{}, ErrGameDataMissing
	}
	playerStats, err := parseCombatStats(profile.Stats)
	if err != nil {
		return fightReportContext{}, err
	}
	enemyStats, err := parseCombatStats(enemyStatsDocument)
	if err != nil {
		return fightReportContext{}, err
	}
	return fightReportContext{
		TimeLimitMS: timeLimitSeconds * 1000,
		Player:      playerStats,
		Enemy:       enemyStats,
		StartedAt:   fightSession.StartedAt,
		ReportedAt:  results.dependencies.Now().UTC(),
	}, nil
}

func (results *fightResultService) reject(ctx context.Context, fightSession models.FightSession, problem string, rejectedAt time.Time) error {
	isRejected, err := results.dependencies.FightSessions.Reject(ctx, fightSession.ID, problem, rejectedAt)
	if err != nil {
		return err
	}
	if !isRejected {
		return ErrFightAlreadyEnded
	}
	results.dependencies.Logger.WarnContext(ctx, "fight result rejected",
		slog.String("fight_session_id", fightSession.ID.String()),
		slog.String("user_id", fightSession.UserID.String()),
		slog.String("reason", problem),
	)
	return apierror.New(http.StatusUnprocessableEntity, apierror.CodeFightResultRejected, "This fight result could not be accepted").
		WithDetails(map[string]string{"reason": problem})
}

package service

import (
	"context"

	"github.com/google/uuid"

	"github.com/farhapartex/nebula-exchange/backend/internal/progress/models"
	"github.com/farhapartex/nebula-exchange/backend/internal/progress/repository"
	storyservice "github.com/farhapartex/nebula-exchange/backend/internal/story/service"
)

type NextLevelStatus string

const (
	NextLevelAvailable  NextLevelStatus = "AVAILABLE"
	NextLevelLocked     NextLevelStatus = "LOCKED"
	NextLevelComingSoon NextLevelStatus = "COMING_SOON"
)

const firstChapterNumber = 1

type NextLevel struct {
	Status        NextLevelStatus
	ChapterNumber int
	ChapterTitle  *string
	ChapterPrice  *int64
	IsChapterPaid bool
	Level         *storyservice.LevelPlacement
	BestStars     *int
	Attempts      int
}

type PlayerProgress struct {
	CurrentLevel       int
	CurrentLevelPrice  *int64
	IsCurrentLevelPaid bool
	StoryLevel         int
	TotalWins          int
	TotalLosses        int
	CurrentLevelWins   int
	CurrentLevelLosses int
	NextLevel          NextLevel
}

type PlayerProgressService interface {
	PlayerProgress(ctx context.Context, userID uuid.UUID) (PlayerProgress, error)
}

type playerProgressService struct {
	levels         LevelCatalog
	levelProgress  repository.LevelProgressRepository
	fightSessions  repository.FightSessionRepository
	fighters       repository.FighterRepository
	chapterUnlocks repository.ChapterUnlockRepository
}

type PlayerProgressDependencies struct {
	Levels         LevelCatalog
	LevelProgress  repository.LevelProgressRepository
	FightSessions  repository.FightSessionRepository
	Fighters       repository.FighterRepository
	ChapterUnlocks repository.ChapterUnlockRepository
}

func NewPlayerProgressService(dependencies PlayerProgressDependencies) PlayerProgressService {
	return &playerProgressService{
		levels:         dependencies.Levels,
		levelProgress:  dependencies.LevelProgress,
		fightSessions:  dependencies.FightSessions,
		fighters:       dependencies.Fighters,
		chapterUnlocks: dependencies.ChapterUnlocks,
	}
}

func (progress *playerProgressService) PlayerProgress(ctx context.Context, userID uuid.UUID) (PlayerProgress, error) {
	orderedPlacements, err := progress.levels.OrderedPlacements(ctx)
	if err != nil {
		return PlayerProgress{}, err
	}
	progressByLevel, err := progress.progressByLevel(ctx, userID)
	if err != nil {
		return PlayerProgress{}, err
	}

	owned, err := loadOwnedChapters(ctx, progress.chapterUnlocks, userID)
	if err != nil {
		return PlayerProgress{}, err
	}
	nextLevel := findNextLevel(orderedPlacements, progressByLevel, owned)
	currentChapterLevelIDs, storyLevel := chapterSummary(orderedPlacements, progressByLevel, nextLevel.ChapterNumber)
	currentLevelWins, currentLevelLosses, err := progress.fightSessions.CountOutcomes(ctx, userID, currentChapterLevelIDs)
	if err != nil {
		return PlayerProgress{}, err
	}
	totalWins, totalLosses, err := progress.totals(ctx, userID)
	if err != nil {
		return PlayerProgress{}, err
	}
	return PlayerProgress{
		CurrentLevel:       nextLevel.ChapterNumber,
		CurrentLevelPrice:  nextLevel.ChapterPrice,
		IsCurrentLevelPaid: nextLevel.IsChapterPaid,
		StoryLevel:         storyLevel,
		TotalWins:          totalWins,
		TotalLosses:        totalLosses,
		CurrentLevelWins:   currentLevelWins,
		CurrentLevelLosses: currentLevelLosses,
		NextLevel:          nextLevel,
	}, nil
}

func (progress *playerProgressService) progressByLevel(ctx context.Context, userID uuid.UUID) (map[string]models.LevelProgress, error) {
	levelProgressRows, err := progress.levelProgress.ListForUser(ctx, userID)
	if err != nil {
		return nil, err
	}
	progressByLevel := make(map[string]models.LevelProgress, len(levelProgressRows))
	for _, levelProgress := range levelProgressRows {
		progressByLevel[levelProgress.LevelID] = levelProgress
	}
	return progressByLevel, nil
}

func (progress *playerProgressService) totals(ctx context.Context, userID uuid.UUID) (int, int, error) {
	profile, hasProfile, err := progress.fighters.FindProfile(ctx, userID)
	if err != nil || !hasProfile {
		return 0, 0, err
	}
	return profile.Wins, profile.Losses, nil
}

func findNextLevel(orderedPlacements []storyservice.LevelPlacement, progressByLevel map[string]models.LevelProgress, owned ownedChapters) NextLevel {
	for placementIndex := range orderedPlacements {
		placement := orderedPlacements[placementIndex]
		levelProgress, hasProgress := progressByLevel[placement.LevelID]
		if hasProgress && levelProgress.Status == models.ProgressStatusCompleted {
			continue
		}
		status := NextLevelAvailable
		if !owned.canPlay(placement) {
			status = NextLevelLocked
		}
		var chapterPrice *int64
		if !placement.IsChapterFree {
			chapterPrice = placement.ChapterPrice
		}
		chapterTitle := placement.ChapterTitle
		return NextLevel{
			Status:        status,
			ChapterNumber: placement.ChapterNumber,
			ChapterTitle:  &chapterTitle,
			ChapterPrice:  chapterPrice,
			IsChapterPaid: owned.hasPaidFor(placement),
			Level:         &placement,
			BestStars:     levelProgress.BestStars,
			Attempts:      levelProgress.Attempts,
		}
	}
	nextChapterNumber := firstChapterNumber
	if len(orderedPlacements) > 0 {
		nextChapterNumber = orderedPlacements[len(orderedPlacements)-1].ChapterNumber + 1
	}
	return NextLevel{Status: NextLevelComingSoon, ChapterNumber: nextChapterNumber}
}

func chapterSummary(orderedPlacements []storyservice.LevelPlacement, progressByLevel map[string]models.LevelProgress, chapterNumber int) ([]string, int) {
	var chapterLevelIDs []string
	highestWonLevel := 0
	for _, placement := range orderedPlacements {
		if placement.ChapterNumber != chapterNumber {
			continue
		}
		chapterLevelIDs = append(chapterLevelIDs, placement.LevelID)
		if levelProgress, hasProgress := progressByLevel[placement.LevelID]; hasProgress && levelProgress.Status == models.ProgressStatusCompleted {
			highestWonLevel = max(highestWonLevel, placement.LevelNumber)
		}
	}
	return chapterLevelIDs, highestWonLevel
}

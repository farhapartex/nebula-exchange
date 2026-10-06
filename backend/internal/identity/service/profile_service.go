package service

import (
	"context"

	"github.com/google/uuid"

	"github.com/farhapartex/nebula-exchange/backend/internal/identity/models"
	"github.com/farhapartex/nebula-exchange/backend/internal/identity/repository"
)

type PlayerSummary struct {
	Name               string
	Email              string
	CurrentLevel       int
	StoryLevel         int
	TotalWins          int
	TotalLosses        int
	CurrentLevelWins   int
	CurrentLevelLosses int
}

type PlayerProgress struct {
	CurrentLevel       int
	StoryLevel         int
	TotalWins          int
	TotalLosses        int
	CurrentLevelWins   int
	CurrentLevelLosses int
}

type PlayerProgressReader interface {
	PlayerProgressOf(ctx context.Context, userID uuid.UUID) (PlayerProgress, error)
}

type ProfileService interface {
	CurrentPlayer(ctx context.Context, userID uuid.UUID) (PlayerSummary, error)
}

type profileService struct {
	users          repository.UserRepository
	playerProgress PlayerProgressReader
}

func NewProfileService(users repository.UserRepository, playerProgress PlayerProgressReader) ProfileService {
	return &profileService{users: users, playerProgress: playerProgress}
}

func (profile *profileService) CurrentPlayer(ctx context.Context, userID uuid.UUID) (PlayerSummary, error) {
	currentUser, isFound, err := profile.users.FindByID(ctx, userID)
	if err != nil {
		return PlayerSummary{}, err
	}
	if !isFound || currentUser.Status == models.UserStatusBanned {
		return PlayerSummary{}, ErrNotLoggedIn
	}
	playerProgress, err := profile.playerProgress.PlayerProgressOf(ctx, userID)
	if err != nil {
		return PlayerSummary{}, err
	}
	return PlayerSummary{
		Name:               currentUser.Username,
		Email:              currentUser.Email,
		CurrentLevel:       playerProgress.CurrentLevel,
		StoryLevel:         playerProgress.StoryLevel,
		TotalWins:          playerProgress.TotalWins,
		TotalLosses:        playerProgress.TotalLosses,
		CurrentLevelWins:   playerProgress.CurrentLevelWins,
		CurrentLevelLosses: playerProgress.CurrentLevelLosses,
	}, nil
}

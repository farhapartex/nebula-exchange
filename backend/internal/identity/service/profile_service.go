package service

import (
	"context"

	"github.com/google/uuid"

	"github.com/farhapartex/nebula-exchange/backend/internal/identity/models"
	"github.com/farhapartex/nebula-exchange/backend/internal/identity/repository"
)

type PlayerSummary struct {
	Name         string
	Email        string
	CurrentLevel int
	StoryLevel   *string
	TotalWins    int
	TotalLosses  int
}

type StoryProgressReader interface {
	FurthestStartedLevel(ctx context.Context, userID uuid.UUID) (*string, error)
}

type ProfileService interface {
	CurrentPlayer(ctx context.Context, userID uuid.UUID) (PlayerSummary, error)
}

type profileService struct {
	users         repository.UserRepository
	storyProgress StoryProgressReader
}

func NewProfileService(users repository.UserRepository, storyProgress StoryProgressReader) ProfileService {
	return &profileService{users: users, storyProgress: storyProgress}
}

func (profile *profileService) CurrentPlayer(ctx context.Context, userID uuid.UUID) (PlayerSummary, error) {
	currentUser, isFound, err := profile.users.FindByID(ctx, userID)
	if err != nil {
		return PlayerSummary{}, err
	}
	if !isFound || currentUser.Status == models.UserStatusBanned {
		return PlayerSummary{}, ErrNotLoggedIn
	}
	storyLevel, err := profile.storyProgress.FurthestStartedLevel(ctx, userID)
	if err != nil {
		return PlayerSummary{}, err
	}
	return PlayerSummary{Name: currentUser.Username, Email: currentUser.Email, StoryLevel: storyLevel}, nil
}

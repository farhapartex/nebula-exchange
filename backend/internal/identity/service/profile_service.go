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
	TotalWins    int
	TotalLosses  int
}

type ProfileService interface {
	CurrentPlayer(ctx context.Context, userID uuid.UUID) (PlayerSummary, error)
}

type profileService struct {
	users repository.UserRepository
}

func NewProfileService(users repository.UserRepository) ProfileService {
	return &profileService{users: users}
}

func (profile *profileService) CurrentPlayer(ctx context.Context, userID uuid.UUID) (PlayerSummary, error) {
	currentUser, isFound, err := profile.users.FindByID(ctx, userID)
	if err != nil {
		return PlayerSummary{}, err
	}
	if !isFound || currentUser.Status == models.UserStatusBanned {
		return PlayerSummary{}, ErrNotLoggedIn
	}
	return PlayerSummary{Name: currentUser.Username, Email: currentUser.Email}, nil
}

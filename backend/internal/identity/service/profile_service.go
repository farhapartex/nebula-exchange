package service

import (
	"context"

	"github.com/google/uuid"

	"github.com/farhapartex/nebula-exchange/backend/internal/identity/models"
	"github.com/farhapartex/nebula-exchange/backend/internal/identity/repository"
)

type ProfileService interface {
	CurrentUser(ctx context.Context, userID uuid.UUID) (models.User, error)
}

type profileService struct {
	users repository.UserRepository
}

func NewProfileService(users repository.UserRepository) ProfileService {
	return &profileService{users: users}
}

func (profile *profileService) CurrentUser(ctx context.Context, userID uuid.UUID) (models.User, error) {
	currentUser, isFound, err := profile.users.FindByID(ctx, userID)
	if err != nil {
		return models.User{}, err
	}
	if !isFound || currentUser.Status == models.UserStatusBanned {
		return models.User{}, ErrNotLoggedIn
	}
	return currentUser, nil
}

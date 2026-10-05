package repository

import (
	"context"
	"errors"
	"strings"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	"github.com/farhapartex/nebula-exchange/backend/internal/identity/models"
	"github.com/farhapartex/nebula-exchange/backend/internal/platform/database"
)

const (
	emailUniqueConstraint    = "users_email_unique"
	usernameUniqueConstraint = "users_username_unique"
)

type TakenIdentifiers struct {
	IsEmailTaken    bool
	IsUsernameTaken bool
}

func (takenIdentifiers TakenIdentifiers) Any() bool {
	return takenIdentifiers.IsEmailTaken || takenIdentifiers.IsUsernameTaken
}

type UserRepository interface {
	Create(ctx context.Context, user *models.User) error
	FindTakenIdentifiers(ctx context.Context, email, username string) (TakenIdentifiers, error)
	FindByEmail(ctx context.Context, email string) (models.User, bool, error)
	FindByID(ctx context.Context, userID uuid.UUID) (models.User, bool, error)
	Activate(ctx context.Context, userID uuid.UUID, activatedAt time.Time) (models.User, bool, error)
	RecordLogin(ctx context.Context, userID uuid.UUID, loggedInAt time.Time) error
}

type GormUserRepository struct {
	database *gorm.DB
}

func NewUserRepository(database *gorm.DB) *GormUserRepository {
	return &GormUserRepository{database: database}
}

func (repository *GormUserRepository) Create(ctx context.Context, user *models.User) error {
	err := database.Session(ctx, repository.database).Create(user).Error
	if constraintName, isUniqueViolation := database.ViolatedUniqueConstraint(err); isUniqueViolation {
		switch constraintName {
		case emailUniqueConstraint:
			return ErrEmailTaken
		case usernameUniqueConstraint:
			return ErrUsernameTaken
		}
	}
	return err
}

func (repository *GormUserRepository) FindTakenIdentifiers(ctx context.Context, email, username string) (TakenIdentifiers, error) {
	var matchingUsers []models.User
	err := database.Session(ctx, repository.database).
		Select("email", "username").
		Where(map[string]any{"email": email}).
		Or(map[string]any{"username": username}).
		Find(&matchingUsers).Error
	if err != nil {
		return TakenIdentifiers{}, err
	}
	var takenIdentifiers TakenIdentifiers
	for _, matchingUser := range matchingUsers {
		takenIdentifiers.IsEmailTaken = takenIdentifiers.IsEmailTaken || strings.EqualFold(matchingUser.Email, email)
		takenIdentifiers.IsUsernameTaken = takenIdentifiers.IsUsernameTaken || strings.EqualFold(matchingUser.Username, username)
	}
	return takenIdentifiers, nil
}

func (repository *GormUserRepository) FindByEmail(ctx context.Context, email string) (models.User, bool, error) {
	return repository.findOne(ctx, map[string]any{"email": email})
}

func (repository *GormUserRepository) FindByID(ctx context.Context, userID uuid.UUID) (models.User, bool, error) {
	return repository.findOne(ctx, map[string]any{"id": userID})
}

func (repository *GormUserRepository) findOne(ctx context.Context, conditions map[string]any) (models.User, bool, error) {
	var foundUser models.User
	err := database.Session(ctx, repository.database).Where(conditions).Take(&foundUser).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return models.User{}, false, nil
	}
	if err != nil {
		return models.User{}, false, err
	}
	return foundUser, true, nil
}

func (repository *GormUserRepository) Activate(ctx context.Context, userID uuid.UUID, activatedAt time.Time) (models.User, bool, error) {
	var activatedUsers []models.User
	result := database.Session(ctx, repository.database).
		Model(&activatedUsers).
		Clauses(clause.Returning{}).
		Where(map[string]any{"id": userID, "is_active": false}).
		Updates(map[string]any{
			"is_active":    true,
			"activated_at": activatedAt,
			"status":       models.UserStatusActive,
			"updated_at":   activatedAt,
		})
	if result.Error != nil || len(activatedUsers) == 0 {
		return models.User{}, false, result.Error
	}
	return activatedUsers[0], true, nil
}

func (repository *GormUserRepository) RecordLogin(ctx context.Context, userID uuid.UUID, loggedInAt time.Time) error {
	return database.Session(ctx, repository.database).
		Model(&models.User{}).
		Where(map[string]any{"id": userID}).
		Updates(map[string]any{"last_login_at": loggedInAt, "updated_at": loggedInAt}).Error
}

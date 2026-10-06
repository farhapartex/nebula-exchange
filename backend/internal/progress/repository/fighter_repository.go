package repository

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	"github.com/farhapartex/nebula-exchange/backend/internal/platform/database"
	"github.com/farhapartex/nebula-exchange/backend/internal/progress/models"
)

type FighterRepository interface {
	FindDefaultTemplate(ctx context.Context) (models.FighterTemplate, bool, error)
	FindProfile(ctx context.Context, userID uuid.UUID) (models.FighterProfile, bool, error)
	CreateProfileIfMissing(ctx context.Context, profile *models.FighterProfile) error
}

type GormFighterRepository struct {
	database *gorm.DB
}

func NewFighterRepository(database *gorm.DB) *GormFighterRepository {
	return &GormFighterRepository{database: database}
}

func (repository *GormFighterRepository) FindDefaultTemplate(ctx context.Context) (models.FighterTemplate, bool, error) {
	var template models.FighterTemplate
	err := database.Session(ctx, repository.database).Where(map[string]any{"is_default": true}).Take(&template).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return models.FighterTemplate{}, false, nil
	}
	if err != nil {
		return models.FighterTemplate{}, false, err
	}
	return template, true, nil
}

func (repository *GormFighterRepository) FindProfile(ctx context.Context, userID uuid.UUID) (models.FighterProfile, bool, error) {
	var profile models.FighterProfile
	err := database.Session(ctx, repository.database).Preload("Template").Where(map[string]any{"user_id": userID}).Take(&profile).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return models.FighterProfile{}, false, nil
	}
	if err != nil {
		return models.FighterProfile{}, false, err
	}
	return profile, true, nil
}

func (repository *GormFighterRepository) CreateProfileIfMissing(ctx context.Context, profile *models.FighterProfile) error {
	return database.Session(ctx, repository.database).
		Omit("Template").
		Clauses(clause.OnConflict{Columns: []clause.Column{{Name: "user_id"}}, DoNothing: true}).
		Create(profile).Error
}

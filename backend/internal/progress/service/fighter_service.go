package service

import (
	"context"
	"time"

	"github.com/google/uuid"

	"github.com/farhapartex/nebula-exchange/backend/internal/platform/apierror"
	"github.com/farhapartex/nebula-exchange/backend/internal/platform/database"
	"github.com/farhapartex/nebula-exchange/backend/internal/progress/models"
	"github.com/farhapartex/nebula-exchange/backend/internal/progress/repository"
)

var ErrGameDataMissing = apierror.ServiceUnavailable("The game data is not loaded yet. Try again later.")

type Fighter struct {
	ID    string
	Name  string
	Title string
	Level int
	Wins  int
	Loss  int
	Stats database.JSONDocument
	Look  database.JSONDocument
}

type fighterResolver struct {
	fighters repository.FighterRepository
}

func (resolver fighterResolver) currentFighter(ctx context.Context, userID uuid.UUID) (Fighter, error) {
	profile, hasProfile, err := resolver.fighters.FindProfile(ctx, userID)
	if err != nil {
		return Fighter{}, err
	}
	if hasProfile {
		return Fighter{
			ID:    profile.Template.ID,
			Name:  profile.Template.Name,
			Title: profile.Template.Title,
			Level: profile.FighterLevel,
			Wins:  profile.Wins,
			Loss:  profile.Losses,
			Stats: profile.Stats,
			Look:  profile.Template.Look,
		}, nil
	}
	template, hasTemplate, err := resolver.fighters.FindDefaultTemplate(ctx)
	if err != nil {
		return Fighter{}, err
	}
	if !hasTemplate {
		return Fighter{}, ErrGameDataMissing
	}
	return Fighter{ID: template.ID, Name: template.Name, Title: template.Title, Level: template.StartingLevel, Stats: template.Stats, Look: template.Look}, nil
}

func (resolver fighterResolver) ensureProfile(ctx context.Context, userID uuid.UUID, createdAt time.Time) error {
	template, hasTemplate, err := resolver.fighters.FindDefaultTemplate(ctx)
	if err != nil {
		return err
	}
	if !hasTemplate {
		return ErrGameDataMissing
	}
	return resolver.fighters.CreateProfileIfMissing(ctx, &models.FighterProfile{
		UserID:       userID,
		TemplateID:   template.ID,
		FighterLevel: template.StartingLevel,
		Stats:        template.Stats,
		CreatedAt:    createdAt,
		UpdatedAt:    createdAt,
	})
}

package seeding

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"os"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	"github.com/farhapartex/nebula-exchange/backend/internal/platform/database"
	"github.com/farhapartex/nebula-exchange/backend/internal/progress/models"
)

type FighterTemplateContent struct {
	ID            string                `json:"id"`
	Name          string                `json:"name"`
	Title         string                `json:"title"`
	StartingLevel int                   `json:"starting_level"`
	IsDefault     bool                  `json:"is_default"`
	Stats         database.JSONDocument `json:"stats"`
	Look          database.JSONDocument `json:"look"`
}

func LoadFighterTemplate(filePath string) (FighterTemplateContent, error) {
	templateFile, err := os.ReadFile(filePath)
	if err != nil {
		return FighterTemplateContent{}, err
	}
	var template FighterTemplateContent
	if err := json.Unmarshal(templateFile, &template); err != nil {
		return FighterTemplateContent{}, fmt.Errorf("parse %s: %w", filePath, err)
	}
	if template.ID == "" || template.Name == "" || template.StartingLevel < 1 || len(template.Stats) == 0 || len(template.Look) == 0 {
		return FighterTemplateContent{}, errors.New("a fighter template needs an id, a name, a starting level of 1 or more, stats and a look")
	}
	return template, nil
}

func SeedFighterTemplate(ctx context.Context, gormDatabase *gorm.DB, template FighterTemplateContent, logger *slog.Logger) error {
	return gormDatabase.WithContext(ctx).Transaction(func(transaction *gorm.DB) error {
		if template.IsDefault {
			err := transaction.Model(&models.FighterTemplate{}).
				Where(map[string]any{"is_default": true}).
				Where("id <> ?", template.ID).
				Update("is_default", false).Error
			if err != nil {
				return err
			}
		}
		err := transaction.Clauses(clause.OnConflict{
			Columns:   []clause.Column{{Name: "id"}},
			DoUpdates: clause.AssignmentColumns([]string{"name", "title", "starting_level", "stats", "look", "is_default", "updated_at"}),
		}).Create(&models.FighterTemplate{
			ID:            template.ID,
			Name:          template.Name,
			Title:         template.Title,
			StartingLevel: template.StartingLevel,
			Stats:         template.Stats,
			Look:          template.Look,
			IsDefault:     template.IsDefault,
		}).Error
		if err != nil {
			return fmt.Errorf("save fighter template %s: %w", template.ID, err)
		}
		logger.InfoContext(ctx, "fighter template seeded", slog.String("template_id", template.ID), slog.Bool("is_default", template.IsDefault))
		return nil
	})
}

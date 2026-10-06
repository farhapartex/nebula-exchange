package seeding

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"os"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	"github.com/farhapartex/nebula-exchange/backend/internal/billing/models"
	"github.com/farhapartex/nebula-exchange/backend/internal/billing/service"
	"github.com/farhapartex/nebula-exchange/backend/internal/platform/database"
)

type PlanContent struct {
	ID            string                `json:"id"`
	Kind          models.PlanKind       `json:"kind"`
	Name          string                `json:"name"`
	Description   string                `json:"description"`
	DiscountTiers database.JSONDocument `json:"discount_tiers"`
	SortOrder     int                   `json:"sort_order"`
	IsActive      bool                  `json:"is_active"`
}

func LoadPlans(filePath string) ([]PlanContent, error) {
	plansFile, err := os.ReadFile(filePath)
	if err != nil {
		return nil, err
	}
	var plans []PlanContent
	if err := json.Unmarshal(plansFile, &plans); err != nil {
		return nil, fmt.Errorf("parse %s: %w", filePath, err)
	}
	for _, plan := range plans {
		if plan.ID == "" || plan.Name == "" {
			return nil, fmt.Errorf("every plan needs an id and a name")
		}
		switch plan.Kind {
		case models.PlanKindSingleChapter, models.PlanKindChapterBundle, models.PlanKindAllChapters:
		default:
			return nil, fmt.Errorf("plan %s has an unknown kind %q", plan.ID, plan.Kind)
		}
		if _, err := service.ParseDiscountTiers(plan.DiscountTiers); err != nil {
			return nil, fmt.Errorf("plan %s: %w", plan.ID, err)
		}
	}
	return plans, nil
}

func SeedPlans(ctx context.Context, gormDatabase *gorm.DB, plans []PlanContent, logger *slog.Logger) error {
	return gormDatabase.WithContext(ctx).Transaction(func(transaction *gorm.DB) error {
		for _, plan := range plans {
			discountTiers := plan.DiscountTiers
			if len(discountTiers) == 0 {
				discountTiers = database.JSONDocument(`[]`)
			}
			err := transaction.Clauses(clause.OnConflict{
				Columns:   []clause.Column{{Name: "id"}},
				DoUpdates: clause.AssignmentColumns([]string{"kind", "name", "description", "discount_tiers", "sort_order", "is_active", "updated_at"}),
			}).Create(&models.Plan{
				ID:            plan.ID,
				Kind:          plan.Kind,
				Name:          plan.Name,
				Description:   plan.Description,
				DiscountTiers: discountTiers,
				SortOrder:     plan.SortOrder,
				IsActive:      plan.IsActive,
			}).Error
			if err != nil {
				return fmt.Errorf("save plan %s: %w", plan.ID, err)
			}
		}
		logger.InfoContext(ctx, "plans seeded", slog.Int("plans", len(plans)))
		return nil
	})
}

package repository

import (
	"context"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"github.com/farhapartex/nebula-exchange/backend/internal/platform/database"
	"github.com/farhapartex/nebula-exchange/backend/internal/tools/models"
)

type ToolRepository interface {
	CountOwnedByType(ctx context.Context, ownerUserID uuid.UUID, toolTypeIDs []string) (map[string]int, error)
}

type GormToolRepository struct {
	database *gorm.DB
}

func NewToolRepository(database *gorm.DB) *GormToolRepository {
	return &GormToolRepository{database: database}
}

type ownedToolCount struct {
	ToolTypeID string
	OwnedCount int
}

func (repository *GormToolRepository) CountOwnedByType(ctx context.Context, ownerUserID uuid.UUID, toolTypeIDs []string) (map[string]int, error) {
	ownedCounts := make(map[string]int, len(toolTypeIDs))
	if len(toolTypeIDs) == 0 {
		return ownedCounts, nil
	}
	var countRows []ownedToolCount
	err := database.Session(ctx, repository.database).
		Model(&models.Tool{}).
		Select("tool_type_id, count(*) AS owned_count").
		Where(map[string]any{"owner_user_id": ownerUserID, "tool_type_id": toolTypeIDs}).
		Group("tool_type_id").
		Scan(&countRows).Error
	if err != nil {
		return nil, err
	}
	for _, countRow := range countRows {
		ownedCounts[countRow.ToolTypeID] = countRow.OwnedCount
	}
	return ownedCounts, nil
}

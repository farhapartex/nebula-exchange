package repository

import (
	"context"
	"strings"

	"gorm.io/gorm"

	"github.com/farhapartex/nebula-exchange/backend/internal/platform/database"
	"github.com/farhapartex/nebula-exchange/backend/internal/tools/models"
)

type ShopFilter struct {
	Search   string
	Category *models.ToolCategory
	Rarity   *models.ToolRarity
}

type ShopPosition struct {
	MinimumFighterLevel int16
	ID                  string
}

type ToolTypeRepository interface {
	ListInShop(ctx context.Context, filter ShopFilter, after *ShopPosition, limit int) ([]models.ToolType, error)
}

type GormToolTypeRepository struct {
	database *gorm.DB
}

func NewToolTypeRepository(database *gorm.DB) *GormToolTypeRepository {
	return &GormToolTypeRepository{database: database}
}

var likePatternEscaper = strings.NewReplacer(`\`, `\\`, `%`, `\%`, `_`, `\_`)

func (repository *GormToolTypeRepository) ListInShop(ctx context.Context, filter ShopFilter, after *ShopPosition, limit int) ([]models.ToolType, error) {
	query := database.Session(ctx, repository.database).Where("shop_price_coins IS NOT NULL")
	if filter.Search != "" {
		query = query.Where(`name ILIKE ? ESCAPE '\'`, "%"+likePatternEscaper.Replace(filter.Search)+"%")
	}
	if filter.Category != nil {
		query = query.Where(map[string]any{"category": *filter.Category})
	}
	if filter.Rarity != nil {
		query = query.Where(map[string]any{"rarity": *filter.Rarity})
	}
	if after != nil {
		query = query.Where("(minimum_fighter_level, id) > (?, ?)", after.MinimumFighterLevel, after.ID)
	}
	var toolTypes []models.ToolType
	err := query.Order("minimum_fighter_level").Order("id").Limit(limit).Find(&toolTypes).Error
	return toolTypes, err
}

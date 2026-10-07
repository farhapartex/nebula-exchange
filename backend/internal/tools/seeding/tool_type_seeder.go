package seeding

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	"github.com/farhapartex/nebula-exchange/backend/internal/platform/database"
	"github.com/farhapartex/nebula-exchange/backend/internal/tools/models"
)

type ToolTypeContent struct {
	ID                  string                `json:"id"`
	Name                string                `json:"name"`
	Description         string                `json:"description"`
	Category            models.ToolCategory   `json:"category"`
	Rarity              models.ToolRarity     `json:"rarity"`
	BaseStats           database.JSONDocument `json:"base_stats"`
	MasteryCurve        database.JSONDocument `json:"mastery_curve"`
	ShopPriceCoins      *int64                `json:"shop_price_coins"`
	IsTradeable         bool                  `json:"is_tradeable"`
	MaxSupply           *int32                `json:"max_supply"`
	IntroducedInLevelID *string               `json:"introduced_in_level_id"`
	MinimumFighterLevel int16                 `json:"minimum_fighter_level"`
	ImageSVGFile        string                `json:"image_svg_file"`
	ImageSVG            string                `json:"-"`
}

type masteryCurve struct {
	PointsToReachLevel []int32                    `json:"points_to_reach_level"`
	StatGainPerLevel   map[string]json.RawMessage `json:"stat_gain_per_level"`
}

const maximumMasteryLevel = 10

var (
	knownCategories = map[models.ToolCategory]bool{models.ToolCategoryWeapon: true, models.ToolCategoryGuard: true}
	knownRarities   = map[models.ToolRarity]bool{
		models.ToolRarityCommon: true, models.ToolRarityUncommon: true, models.ToolRarityRare: true,
		models.ToolRarityEpic: true, models.ToolRarityLegendary: true,
	}
)

func LoadToolType(filePath string) (ToolTypeContent, error) {
	toolTypeFile, err := os.ReadFile(filePath)
	if err != nil {
		return ToolTypeContent{}, err
	}
	var content ToolTypeContent
	if err := json.Unmarshal(toolTypeFile, &content); err != nil {
		return ToolTypeContent{}, fmt.Errorf("parse %s: %w", filePath, err)
	}
	if content.ImageSVGFile == "" || filepath.Base(content.ImageSVGFile) != content.ImageSVGFile {
		return ToolTypeContent{}, fmt.Errorf("%s: image_svg_file must name an SVG file next to the tool file", filePath)
	}
	svgMarkup, err := os.ReadFile(filepath.Join(filepath.Dir(filePath), content.ImageSVGFile))
	if err != nil {
		return ToolTypeContent{}, fmt.Errorf("%s: read image: %w", filePath, err)
	}
	content.ImageSVG = string(svgMarkup)
	if err := validateToolType(content); err != nil {
		return ToolTypeContent{}, fmt.Errorf("%s: %w", filePath, err)
	}
	return content, nil
}

func validateToolType(content ToolTypeContent) error {
	if content.ID == "" || content.Name == "" || content.Description == "" {
		return errors.New("a tool needs an id, a name and a description")
	}
	if !knownCategories[content.Category] {
		return fmt.Errorf("unknown category %q", content.Category)
	}
	if !knownRarities[content.Rarity] {
		return fmt.Errorf("unknown rarity %q", content.Rarity)
	}
	var baseStats map[string]json.RawMessage
	if err := json.Unmarshal(content.BaseStats, &baseStats); err != nil || len(baseStats) == 0 {
		return errors.New("base_stats must be an object with at least one stat")
	}
	if content.ShopPriceCoins != nil && *content.ShopPriceCoins <= 0 {
		return errors.New("shop_price_coins must be above zero or left out")
	}
	if content.MaxSupply != nil && *content.MaxSupply <= 0 {
		return errors.New("max_supply must be above zero or left out")
	}
	if content.MinimumFighterLevel < 1 {
		return errors.New("minimum_fighter_level must be 1 or more")
	}
	if err := validateSVGImage(content.ImageSVG); err != nil {
		return err
	}
	_, err := masteryLevelsOf(content.MasteryCurve)
	return err
}

func masteryLevelsOf(curveDocument database.JSONDocument) (int16, error) {
	var curve masteryCurve
	if err := json.Unmarshal(curveDocument, &curve); err != nil {
		return 0, errors.New("mastery_curve must be an object")
	}
	levelCount := len(curve.PointsToReachLevel)
	if levelCount < 1 || levelCount > maximumMasteryLevel || curve.PointsToReachLevel[0] != 0 {
		return 0, fmt.Errorf("points_to_reach_level needs 1 to %d levels and starts at 0", maximumMasteryLevel)
	}
	for levelIndex := 1; levelIndex < levelCount; levelIndex++ {
		if curve.PointsToReachLevel[levelIndex] <= curve.PointsToReachLevel[levelIndex-1] {
			return 0, errors.New("points_to_reach_level must go up level by level")
		}
	}
	if len(curve.StatGainPerLevel) == 0 {
		return 0, errors.New("stat_gain_per_level needs at least one stat")
	}
	return int16(levelCount), nil
}

func SeedToolType(ctx context.Context, gormDatabase *gorm.DB, content ToolTypeContent, logger *slog.Logger) error {
	maxMasteryLevel, err := masteryLevelsOf(content.MasteryCurve)
	if err != nil {
		return err
	}
	toolType := models.ToolType{
		ID:                  content.ID,
		Name:                content.Name,
		Description:         content.Description,
		Category:            content.Category,
		Rarity:              content.Rarity,
		BaseStats:           content.BaseStats,
		MasteryCurve:        content.MasteryCurve,
		MaxMasteryLevel:     maxMasteryLevel,
		ShopPriceCoins:      content.ShopPriceCoins,
		IsTradeable:         content.IsTradeable,
		MaxSupply:           content.MaxSupply,
		IntroducedInLevelID: content.IntroducedInLevelID,
		MinimumFighterLevel: content.MinimumFighterLevel,
		ImageSVG:            &content.ImageSVG,
	}
	err = gormDatabase.WithContext(ctx).
		Clauses(clause.OnConflict{
			Columns: []clause.Column{{Name: "id"}},
			DoUpdates: clause.AssignmentColumns([]string{
				"name", "description", "category", "rarity", "base_stats", "mastery_curve", "max_mastery_level",
				"shop_price_coins", "is_tradeable", "max_supply", "introduced_in_level_id", "minimum_fighter_level", "image_svg", "updated_at",
			}),
		}).
		Create(&toolType).Error
	if err != nil {
		return fmt.Errorf("seed tool %s: %w", content.ID, err)
	}
	logger.InfoContext(ctx, "tool seeded", slog.String("tool_type_id", content.ID), slog.String("category", string(content.Category)))
	return nil
}

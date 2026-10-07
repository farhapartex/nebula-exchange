package service

import (
	"context"

	"github.com/google/uuid"

	"github.com/farhapartex/nebula-exchange/backend/internal/platform/pagination"
	"github.com/farhapartex/nebula-exchange/backend/internal/tools/models"
	"github.com/farhapartex/nebula-exchange/backend/internal/tools/repository"
)

type UnlockLevel struct {
	ID            string
	ChapterNumber int
	Number        int
	Title         string
}

type ShopItem struct {
	ToolType    models.ToolType
	PriceCoins  int64
	IsUnlocked  bool
	UnlockLevel *UnlockLevel
	IsUsable    bool
	OwnedCount  int
}

type ShopItemCursor struct {
	MinimumFighterLevel int16  `json:"minimum_fighter_level"`
	ID                  string `json:"id"`
}

type ShopItemService interface {
	List(ctx context.Context, userID uuid.UUID, filter repository.ShopFilter, after *ShopItemCursor, pageRequest pagination.Request) (pagination.Page[ShopItem], error)
}

type shopItemService struct {
	toolTypes repository.ToolTypeRepository
	tools     repository.ToolRepository
	standing  PlayerStanding
	levels    LevelDirectory
}

func NewShopItemService(toolTypes repository.ToolTypeRepository, tools repository.ToolRepository, standing PlayerStanding, levels LevelDirectory) ShopItemService {
	return &shopItemService{toolTypes: toolTypes, tools: tools, standing: standing, levels: levels}
}

func (shop *shopItemService) List(ctx context.Context, userID uuid.UUID, filter repository.ShopFilter, after *ShopItemCursor, pageRequest pagination.Request) (pagination.Page[ShopItem], error) {
	var afterPosition *repository.ShopPosition
	if after != nil {
		afterPosition = &repository.ShopPosition{MinimumFighterLevel: after.MinimumFighterLevel, ID: after.ID}
	}
	toolTypes, err := shop.toolTypes.ListInShop(ctx, filter, afterPosition, pageRequest.FetchLimit())
	if err != nil {
		return pagination.Page[ShopItem]{}, err
	}
	fighterLevel, err := shop.standing.FighterLevel(ctx, userID)
	if err != nil {
		return pagination.Page[ShopItem]{}, err
	}
	wonLevelIDs, err := shop.standing.WonLevelIDs(ctx, userID)
	if err != nil {
		return pagination.Page[ShopItem]{}, err
	}
	placements, err := shop.levels.Placements(ctx)
	if err != nil {
		return pagination.Page[ShopItem]{}, err
	}
	toolTypeIDs := make([]string, 0, len(toolTypes))
	for _, toolType := range toolTypes {
		toolTypeIDs = append(toolTypeIDs, toolType.ID)
	}
	ownedCounts, err := shop.tools.CountOwnedByType(ctx, userID, toolTypeIDs)
	if err != nil {
		return pagination.Page[ShopItem]{}, err
	}

	shopItems := make([]ShopItem, 0, len(toolTypes))
	for _, toolType := range toolTypes {
		shopItem := ShopItem{
			ToolType:   toolType,
			PriceCoins: *toolType.ShopPriceCoins,
			IsUnlocked: toolType.IntroducedInLevelID == nil || wonLevelIDs[*toolType.IntroducedInLevelID],
			IsUsable:   fighterLevel >= int(toolType.MinimumFighterLevel),
			OwnedCount: ownedCounts[toolType.ID],
		}
		if toolType.IntroducedInLevelID != nil {
			if placement, isKnownLevel := placements[*toolType.IntroducedInLevelID]; isKnownLevel {
				shopItem.UnlockLevel = &UnlockLevel{ID: placement.LevelID, ChapterNumber: placement.ChapterNumber, Number: placement.LevelNumber, Title: placement.Title}
			}
		}
		shopItems = append(shopItems, shopItem)
	}
	return pagination.BuildPage(shopItems, pageRequest, func(shopItem ShopItem) ShopItemCursor {
		return ShopItemCursor{MinimumFighterLevel: shopItem.ToolType.MinimumFighterLevel, ID: shopItem.ToolType.ID}
	})
}

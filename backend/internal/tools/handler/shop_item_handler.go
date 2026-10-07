package handler

import (
	"encoding/json"
	"net/http"
	"strconv"
	"strings"
	"unicode/utf8"

	"github.com/gin-gonic/gin"

	"github.com/farhapartex/nebula-exchange/backend/internal/platform/apierror"
	"github.com/farhapartex/nebula-exchange/backend/internal/platform/httpserver/authentication"
	"github.com/farhapartex/nebula-exchange/backend/internal/platform/httpserver/request"
	"github.com/farhapartex/nebula-exchange/backend/internal/platform/httpserver/response"
	"github.com/farhapartex/nebula-exchange/backend/internal/platform/pagination"
	"github.com/farhapartex/nebula-exchange/backend/internal/tools/models"
	"github.com/farhapartex/nebula-exchange/backend/internal/tools/repository"
	"github.com/farhapartex/nebula-exchange/backend/internal/tools/service"
)

const maximumSearchLength = 100

var (
	knownCategories = map[models.ToolCategory]bool{models.ToolCategoryWeapon: true, models.ToolCategoryGuard: true}
	knownRarities   = map[models.ToolRarity]bool{
		models.ToolRarityCommon: true, models.ToolRarityUncommon: true, models.ToolRarityRare: true,
		models.ToolRarityEpic: true, models.ToolRarityLegendary: true,
	}
)

type ToolTypeResponse struct {
	ID                  string              `json:"id"`
	Name                string              `json:"name"`
	Description         string              `json:"description"`
	Category            models.ToolCategory `json:"category"`
	Rarity              models.ToolRarity   `json:"rarity"`
	BaseStats           json.RawMessage     `json:"base_stats"`
	MaxMasteryLevel     int16               `json:"max_mastery_level"`
	MinimumFighterLevel int16               `json:"minimum_fighter_level"`
	ImageSVG            *string             `json:"image_svg"`
}

type UnlockLevelResponse struct {
	ID            string `json:"id"`
	ChapterNumber int    `json:"chapter_number"`
	Number        int    `json:"number"`
	Title         string `json:"title"`
}

type ShopItemResponse struct {
	ToolType    ToolTypeResponse     `json:"tool_type"`
	PriceCoins  string               `json:"price_coins"`
	IsUnlocked  bool                 `json:"is_unlocked"`
	UnlockLevel *UnlockLevelResponse `json:"unlock_level"`
	IsUsable    bool                 `json:"is_usable"`
	OwnedCount  int                  `json:"owned_count"`
}

type ShopItemHandler struct {
	shopItemService service.ShopItemService
}

func NewShopItemHandler(shopItemService service.ShopItemService) *ShopItemHandler {
	return &ShopItemHandler{shopItemService: shopItemService}
}

func (handler *ShopItemHandler) RegisterRoutes(router gin.IRouter) {
	router.GET("/shop-items", authentication.RequireUser(), handler.listShopItems)
}

func (handler *ShopItemHandler) listShopItems(context *gin.Context) {
	pageRequest, err := request.PaginationFromQuery(context)
	if err != nil {
		response.WriteError(context, err)
		return
	}
	filter, err := shopFilterFromQuery(context)
	if err != nil {
		response.WriteError(context, err)
		return
	}
	var after *service.ShopItemCursor
	if pageRequest.HasCursor() {
		cursor, err := request.DecodeCursorPosition[service.ShopItemCursor](pageRequest)
		if err != nil {
			response.WriteError(context, err)
			return
		}
		after = &cursor
	}
	userID, _ := authentication.UserIDFrom(context)
	shopItemPage, err := handler.shopItemService.List(context.Request.Context(), userID, filter, after, pageRequest)
	if err != nil {
		response.WriteError(context, err)
		return
	}
	shopItemResponses := make([]ShopItemResponse, 0, len(shopItemPage.Items))
	for _, shopItem := range shopItemPage.Items {
		shopItemResponses = append(shopItemResponses, toShopItemResponse(shopItem))
	}
	response.WriteList(context, http.StatusOK, pagination.Page[ShopItemResponse]{Items: shopItemResponses, Info: shopItemPage.Info})
}

func shopFilterFromQuery(context *gin.Context) (repository.ShopFilter, error) {
	search := strings.TrimSpace(context.Query("q"))
	if utf8.RuneCountInString(search) > maximumSearchLength {
		return repository.ShopFilter{}, apierror.ValidationFailed(map[string]string{"q": "must be at most " + strconv.Itoa(maximumSearchLength) + " characters"})
	}
	filter := repository.ShopFilter{Search: search}
	if rawCategory := context.Query("category"); rawCategory != "" {
		category := models.ToolCategory(rawCategory)
		if !knownCategories[category] {
			return repository.ShopFilter{}, apierror.ValidationFailed(map[string]string{"category": "must be WEAPON or GUARD"})
		}
		filter.Category = &category
	}
	if rawRarity := context.Query("rarity"); rawRarity != "" {
		rarity := models.ToolRarity(rawRarity)
		if !knownRarities[rarity] {
			return repository.ShopFilter{}, apierror.ValidationFailed(map[string]string{"rarity": "must be COMMON, UNCOMMON, RARE, EPIC or LEGENDARY"})
		}
		filter.Rarity = &rarity
	}
	return filter, nil
}

func toShopItemResponse(shopItem service.ShopItem) ShopItemResponse {
	shopItemResponse := ShopItemResponse{
		ToolType: ToolTypeResponse{
			ID:                  shopItem.ToolType.ID,
			Name:                shopItem.ToolType.Name,
			Description:         shopItem.ToolType.Description,
			Category:            shopItem.ToolType.Category,
			Rarity:              shopItem.ToolType.Rarity,
			BaseStats:           json.RawMessage(shopItem.ToolType.BaseStats),
			MaxMasteryLevel:     shopItem.ToolType.MaxMasteryLevel,
			MinimumFighterLevel: shopItem.ToolType.MinimumFighterLevel,
			ImageSVG:            shopItem.ToolType.ImageSVG,
		},
		PriceCoins: strconv.FormatInt(shopItem.PriceCoins, 10),
		IsUnlocked: shopItem.IsUnlocked,
		IsUsable:   shopItem.IsUsable,
		OwnedCount: shopItem.OwnedCount,
	}
	if shopItem.UnlockLevel != nil {
		shopItemResponse.UnlockLevel = &UnlockLevelResponse{
			ID:            shopItem.UnlockLevel.ID,
			ChapterNumber: shopItem.UnlockLevel.ChapterNumber,
			Number:        shopItem.UnlockLevel.Number,
			Title:         shopItem.UnlockLevel.Title,
		}
	}
	return shopItemResponse
}

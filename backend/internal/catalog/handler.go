package catalog

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"

	"nebula-exchange/backend/internal/auth/authentication"
	"nebula-exchange/backend/internal/platform/apierror"
	"nebula-exchange/backend/internal/platform/httpserver/request"
	"nebula-exchange/backend/internal/platform/httpserver/response"
	"nebula-exchange/backend/internal/platform/pagination"
)

type Handler struct {
	service          *Service
	zoneAvailability ZoneAvailability
}

func NewHandler(service *Service, zoneAvailability ZoneAvailability) *Handler {
	return &Handler{service: service, zoneAvailability: zoneAvailability}
}

func (handler *Handler) RegisterRoutes(router gin.IRouter) {
	router.GET("/items", handler.listItems)
	router.GET("/items/:itemID", handler.getItem)
	router.GET("/recipes", handler.listRecipes)
	router.GET("/upgrades", handler.listUpgrades)
	router.GET("/zones", handler.listZones)
	router.GET("/shop-items", handler.listShopItems)
}

func (handler *Handler) listItems(context *gin.Context) {
	writeCatalogList(context, handler.service, func(snapshot Snapshot) []Item { return snapshot.Items }, func(item Item) string {
		return strconv.Itoa(item.ID)
	})
}

func (handler *Handler) getItem(context *gin.Context) {
	itemID, err := strconv.Atoi(context.Param("itemID"))
	if err != nil {
		response.WriteError(context, apierror.NotFound("This item does not exist"))
		return
	}
	snapshot, err := handler.service.Snapshot(context.Request.Context())
	if err != nil {
		response.WriteError(context, err)
		return
	}
	item, isFound := snapshot.ItemByID(itemID)
	if !isFound {
		response.WriteError(context, apierror.NotFound("This item does not exist"))
		return
	}
	response.WriteData(context, http.StatusOK, ItemDetail{Item: item, Usage: snapshot.UsageOf(itemID)})
}

func (handler *Handler) listRecipes(context *gin.Context) {
	writeCatalogList(context, handler.service, func(snapshot Snapshot) []Recipe { return snapshot.Recipes }, func(recipe Recipe) string {
		return recipe.ID
	})
}

func (handler *Handler) listUpgrades(context *gin.Context) {
	writeCatalogList(context, handler.service, func(snapshot Snapshot) []Upgrade { return snapshot.Upgrades }, func(upgrade Upgrade) string {
		return upgrade.ID
	})
}

func (handler *Handler) listZones(context *gin.Context) {
	pageRequest, err := request.PaginationFromQuery(context)
	if err != nil {
		response.WriteError(context, err)
		return
	}
	snapshot, err := handler.service.Snapshot(context.Request.Context())
	if err != nil {
		response.WriteError(context, err)
		return
	}
	zonePage, err := paginateInOrder(snapshot.Zones, pageRequest, func(zone Zone) string { return zone.ID })
	if err != nil {
		response.WriteError(context, err)
		return
	}

	var unlocks map[string]ZoneUnlock
	if userID, isSignedIn := authentication.UserIDFrom(context); isSignedIn && handler.zoneAvailability != nil {
		if unlocks, err = handler.zoneAvailability.ForPlayer(context.Request.Context(), userID, zonePage.Items); err != nil {
			response.WriteError(context, err)
			return
		}
	}
	playerZones := make([]PlayerZone, 0, len(zonePage.Items))
	for _, zone := range zonePage.Items {
		playerZone := PlayerZone{Zone: zone}
		if unlock, isKnown := unlocks[zone.ID]; isKnown {
			playerZone.Unlock = &unlock
		}
		playerZones = append(playerZones, playerZone)
	}
	response.WriteList(context, http.StatusOK, pagination.Page[PlayerZone]{Items: playerZones, Info: zonePage.Info})
}

func (handler *Handler) listShopItems(context *gin.Context) {
	writeCatalogList(context, handler.service, func(snapshot Snapshot) []ShopItem { return snapshot.ShopItems }, func(shopItem ShopItem) string {
		return shopItem.SKU
	})
}

func writeCatalogList[Entry any](context *gin.Context, service *Service, entriesOf func(Snapshot) []Entry, keyOf func(Entry) string) {
	pageRequest, err := request.PaginationFromQuery(context)
	if err != nil {
		response.WriteError(context, err)
		return
	}
	snapshot, err := service.Snapshot(context.Request.Context())
	if err != nil {
		response.WriteError(context, err)
		return
	}
	entryPage, err := paginateInOrder(entriesOf(snapshot), pageRequest, keyOf)
	if err != nil {
		response.WriteError(context, err)
		return
	}
	response.WriteList[Entry](context, http.StatusOK, entryPage)
}

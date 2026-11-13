package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/farhapartex/nebula-exchange/backend/internal/payment/service"
	"github.com/farhapartex/nebula-exchange/backend/internal/platform/httpserver/authentication"
	"github.com/farhapartex/nebula-exchange/backend/internal/platform/httpserver/request"
	"github.com/farhapartex/nebula-exchange/backend/internal/platform/httpserver/response"
	"github.com/farhapartex/nebula-exchange/backend/internal/platform/pagination"
)

type CoinPackResponse struct {
	ID         string `json:"id"`
	Name       string `json:"name"`
	Coins      string `json:"coins"`
	PriceCents string `json:"price_cents"`
}

type CoinPackHandler struct {
	catalogService service.CoinPackCatalogService
}

func NewCoinPackHandler(catalogService service.CoinPackCatalogService) *CoinPackHandler {
	return &CoinPackHandler{catalogService: catalogService}
}

func (handler *CoinPackHandler) RegisterRoutes(router gin.IRouter) {
	router.GET("/coin-packs", authentication.RequireUser(), handler.listCoinPacks)
}

func (handler *CoinPackHandler) listCoinPacks(context *gin.Context) {
	pageRequest, err := request.PaginationFromQuery(context)
	if err != nil {
		response.WriteError(context, err)
		return
	}
	var after service.CoinPackCursor
	if pageRequest.HasCursor() {
		if after, err = request.DecodeCursorPosition[service.CoinPackCursor](pageRequest); err != nil {
			response.WriteError(context, err)
			return
		}
	}
	coinPackPage, err := handler.catalogService.List(context.Request.Context(), after, pageRequest)
	if err != nil {
		response.WriteError(context, err)
		return
	}
	coinPackResponses := make([]CoinPackResponse, 0, len(coinPackPage.Items))
	for _, coinPack := range coinPackPage.Items {
		coinPackResponses = append(coinPackResponses, CoinPackResponse{
			ID:         coinPack.ID,
			Name:       coinPack.Name,
			Coins:      formatCents(coinPack.Coins),
			PriceCents: formatCents(coinPack.PriceCents),
		})
	}
	response.WriteList(context, http.StatusOK, pagination.Page[CoinPackResponse]{Items: coinPackResponses, Info: coinPackPage.Info})
}

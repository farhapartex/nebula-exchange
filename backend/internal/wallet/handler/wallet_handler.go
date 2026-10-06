package handler

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/farhapartex/nebula-exchange/backend/internal/platform/httpserver/authentication"
	"github.com/farhapartex/nebula-exchange/backend/internal/platform/httpserver/request"
	"github.com/farhapartex/nebula-exchange/backend/internal/platform/httpserver/response"
	"github.com/farhapartex/nebula-exchange/backend/internal/platform/pagination"
	"github.com/farhapartex/nebula-exchange/backend/internal/wallet/models"
	"github.com/farhapartex/nebula-exchange/backend/internal/wallet/service"
)

type walletChallengeRequest struct {
	Address string `json:"address" binding:"required,eth_addr"`
	ChainID int64  `json:"chain_id" binding:"required,min=1"`
}

type walletLinkRequest struct {
	Message   string `json:"message" binding:"required,max=2048"`
	Signature string `json:"signature" binding:"required,max=200"`
}

type WalletChallengeResponse struct {
	Nonce     string    `json:"nonce"`
	ExpiresAt time.Time `json:"expires_at"`
}

type WalletResponse struct {
	Address  string    `json:"address"`
	ChainID  int64     `json:"chain_id"`
	LinkedAt time.Time `json:"linked_at"`
}

type WalletHandler struct {
	walletLinkService service.WalletLinkService
}

func NewWalletHandler(walletLinkService service.WalletLinkService) *WalletHandler {
	return &WalletHandler{walletLinkService: walletLinkService}
}

func (handler *WalletHandler) RegisterRoutes(router gin.IRouter) {
	router.POST("/wallet-challenges", authentication.RequireUser(), handler.postWalletChallenge)
	router.POST("/wallets", authentication.RequireUser(), handler.postWallet)
	router.GET("/wallets", authentication.RequireUser(), handler.listWallets)
}

func (handler *WalletHandler) postWalletChallenge(context *gin.Context) {
	var body walletChallengeRequest
	if err := request.BindJSON(context, &body); err != nil {
		response.WriteError(context, err)
		return
	}
	userID, _ := authentication.UserIDFrom(context)
	issuedChallenge, err := handler.walletLinkService.IssueChallenge(context.Request.Context(), userID, body.Address, body.ChainID)
	if err != nil {
		response.WriteError(context, err)
		return
	}
	response.WriteData(context, http.StatusCreated, WalletChallengeResponse{Nonce: issuedChallenge.Nonce, ExpiresAt: issuedChallenge.ExpiresAt})
}

func (handler *WalletHandler) postWallet(context *gin.Context) {
	var body walletLinkRequest
	if err := request.BindJSON(context, &body); err != nil {
		response.WriteError(context, err)
		return
	}
	userID, _ := authentication.UserIDFrom(context)
	linkedWallet, err := handler.walletLinkService.Link(context.Request.Context(), userID, body.Message, body.Signature)
	if err != nil {
		response.WriteError(context, err)
		return
	}
	statusCode := http.StatusOK
	if linkedWallet.IsNewly {
		statusCode = http.StatusCreated
	}
	response.WriteData(context, statusCode, toWalletResponse(linkedWallet.Wallet))
}

func (handler *WalletHandler) listWallets(context *gin.Context) {
	pageRequest, err := request.PaginationFromQuery(context)
	if err != nil {
		response.WriteError(context, err)
		return
	}
	var after *service.WalletCursor
	if pageRequest.HasCursor() {
		cursor, err := request.DecodeCursorPosition[service.WalletCursor](pageRequest)
		if err != nil {
			response.WriteError(context, err)
			return
		}
		after = &cursor
	}
	userID, _ := authentication.UserIDFrom(context)
	walletPage, err := handler.walletLinkService.List(context.Request.Context(), userID, after, pageRequest)
	if err != nil {
		response.WriteError(context, err)
		return
	}
	walletResponses := make([]WalletResponse, 0, len(walletPage.Items))
	for _, wallet := range walletPage.Items {
		walletResponses = append(walletResponses, toWalletResponse(wallet))
	}
	response.WriteList(context, http.StatusOK, pagination.Page[WalletResponse]{Items: walletResponses, Info: walletPage.Info})
}

func toWalletResponse(wallet models.Wallet) WalletResponse {
	return WalletResponse{Address: wallet.Address, ChainID: wallet.ChainID, LinkedAt: wallet.LinkedAt}
}

package upgrades

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"nebula-exchange/backend/internal/auth/authentication"
	"nebula-exchange/backend/internal/platform/httpserver/response"
)

type Handler struct {
	service      *Service
	accountGuard gin.HandlerFunc
}

func NewHandler(service *Service, accountGuard gin.HandlerFunc) *Handler {
	return &Handler{service: service, accountGuard: accountGuard}
}

func (handler *Handler) RegisterRoutes(router gin.IRouter) {
	upgradeRoutes := router.Group("/upgrades/:upgradeID", authentication.RequireUser(), handler.accountGuard)
	upgradeRoutes.POST("/crafts", handler.perform(PathCraft))
	upgradeRoutes.POST("/purchases", handler.perform(PathBuy))
}

func (handler *Handler) perform(path Path) gin.HandlerFunc {
	return func(context *gin.Context) {
		userID, _ := authentication.UserIDFrom(context)
		completed, err := handler.service.Perform(context.Request.Context(), userID, context.Param("upgradeID"), path)
		if err != nil {
			response.WriteError(context, err)
			return
		}
		response.WriteData(context, http.StatusCreated, completed)
	}
}

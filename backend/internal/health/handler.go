package health

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"github.com/farhapartex/nebula-exchange/backend/internal/platform/httpserver/response"
)

type Handler struct {
	healthService Service
}

func NewHandler(healthService Service) *Handler {
	return &Handler{healthService: healthService}
}

func (handler *Handler) RegisterRoutes(router gin.IRouter) {
	router.GET("/health", handler.showHealth)
}

func (handler *Handler) showHealth(context *gin.Context) {
	report := handler.healthService.Check(context.Request.Context())
	statusCode := http.StatusOK
	if !report.IsHealthy() {
		statusCode = http.StatusServiceUnavailable
	}
	response.WriteData(context, statusCode, report)
}

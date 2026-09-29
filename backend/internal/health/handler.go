package health

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"

	"nebula-exchange/backend/internal/platform/httpserver/response"
)

type Status struct {
	Status    string    `json:"status"`
	CheckedAt time.Time `json:"checked_at"`
}

type Handler struct{}

func NewHandler() *Handler {
	return &Handler{}
}

func (handler *Handler) RegisterRoutes(router gin.IRouter) {
	router.GET("/health", handler.getHealth)
}

func (handler *Handler) getHealth(context *gin.Context) {
	response.WriteData(context, http.StatusOK, Status{Status: "ok", CheckedAt: time.Now().UTC()})
}

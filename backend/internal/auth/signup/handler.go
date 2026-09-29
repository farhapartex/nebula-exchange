package signup

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"nebula-exchange/backend/internal/platform/httpserver/request"
	"nebula-exchange/backend/internal/platform/httpserver/response"
)

type Handler struct {
	service *Service
}

func NewHandler(service *Service) *Handler {
	return &Handler{service: service}
}

func (handler *Handler) RegisterRoutes(router gin.IRouter) {
	router.POST("/auth/signup", handler.postSignup)
}

func (handler *Handler) postSignup(context *gin.Context) {
	var signupRequest Request
	if err := request.BindJSON(context, &signupRequest); err != nil {
		response.WriteError(context, err)
		return
	}

	signedUpAccount, err := handler.service.SignUp(context.Request.Context(), signupRequest)
	if err != nil {
		response.WriteError(context, err)
		return
	}
	response.WriteData(context, http.StatusCreated, signedUpAccount)
}

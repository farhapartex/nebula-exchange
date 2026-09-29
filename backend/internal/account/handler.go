package account

import (
	"net/http"

	"github.com/gin-gonic/gin"

	"nebula-exchange/backend/internal/auth/authentication"
	"nebula-exchange/backend/internal/auth/session"
	"nebula-exchange/backend/internal/platform/httpserver/request"
	"nebula-exchange/backend/internal/platform/httpserver/response"
)

type Handler struct {
	service             *Service
	profileGuards       []gin.HandlerFunc
	passwordChangeGuard gin.HandlerFunc
}

func NewHandler(service *Service, passwordChangeGuard gin.HandlerFunc, profileGuards ...gin.HandlerFunc) *Handler {
	return &Handler{service: service, profileGuards: profileGuards, passwordChangeGuard: passwordChangeGuard}
}

func (handler *Handler) RegisterRoutes(router gin.IRouter) {
	router.PATCH("/me", append([]gin.HandlerFunc{authentication.RequireUser()}, append(handler.profileGuards, handler.patchProfile)...)...)
	router.POST("/auth/password-changes", authentication.RequireUser(), handler.passwordChangeGuard, handler.postPasswordChange)
}

func (handler *Handler) patchProfile(context *gin.Context) {
	var profileUpdate ProfileUpdate
	if err := request.BindJSON(context, &profileUpdate); err != nil {
		response.WriteError(context, err)
		return
	}
	userID, _ := authentication.UserIDFrom(context)
	updatedProfile, err := handler.service.UpdateProfile(context.Request.Context(), userID, profileUpdate)
	if err != nil {
		response.WriteError(context, err)
		return
	}
	response.WriteData(context, http.StatusOK, updatedProfile)
}

func (handler *Handler) postPasswordChange(context *gin.Context) {
	var passwordChange PasswordChange
	if err := request.BindJSON(context, &passwordChange); err != nil {
		response.WriteError(context, err)
		return
	}
	userID, _ := authentication.UserIDFrom(context)
	changeResult, err := handler.service.ChangePassword(context.Request.Context(), userID, passwordChange, session.CurrentTokenHash(context))
	if err != nil {
		response.WriteError(context, err)
		return
	}
	response.WriteData(context, http.StatusOK, changeResult)
}

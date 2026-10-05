package handler

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"github.com/farhapartex/nebula-exchange/backend/internal/identity/models"
	"github.com/farhapartex/nebula-exchange/backend/internal/identity/service"
	"github.com/farhapartex/nebula-exchange/backend/internal/platform/httpserver/request"
	"github.com/farhapartex/nebula-exchange/backend/internal/platform/httpserver/response"
)

type signupRequest struct {
	Email        string `json:"email" binding:"required"`
	Username     string `json:"username" binding:"required"`
	Password     string `json:"password" binding:"required"`
	AcceptsTerms bool   `json:"accepts_terms"`
}

type SignedUpAccountResponse struct {
	ID                      uuid.UUID         `json:"id"`
	Email                   string            `json:"email"`
	Username                string            `json:"username"`
	Status                  models.UserStatus `json:"status"`
	ActivationLinkExpiresAt time.Time         `json:"activation_link_expires_at"`
}

type SignupHandler struct {
	signupService service.SignupService
}

func NewSignupHandler(signupService service.SignupService) *SignupHandler {
	return &SignupHandler{signupService: signupService}
}

func (handler *SignupHandler) RegisterRoutes(router gin.IRouter) {
	router.POST("/auth/signup", handler.postSignup)
}

func (handler *SignupHandler) postSignup(context *gin.Context) {
	var body signupRequest
	if err := request.BindJSON(context, &body); err != nil {
		response.WriteError(context, err)
		return
	}
	signedUpAccount, err := handler.signupService.SignUp(context.Request.Context(), service.SignupInput{
		Email:        body.Email,
		Username:     body.Username,
		Password:     body.Password,
		AcceptsTerms: body.AcceptsTerms,
	})
	if err != nil {
		response.WriteError(context, err)
		return
	}
	response.WriteData(context, http.StatusCreated, SignedUpAccountResponse{
		ID:                      signedUpAccount.User.ID,
		Email:                   signedUpAccount.User.Email,
		Username:                signedUpAccount.User.Username,
		Status:                  signedUpAccount.User.Status,
		ActivationLinkExpiresAt: signedUpAccount.ActivationLinkExpiresAt,
	})
}

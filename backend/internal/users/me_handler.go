package users

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"

	"nebula-exchange/backend/internal/auth/authentication"
	"nebula-exchange/backend/internal/platform/apierror"
	"nebula-exchange/backend/internal/platform/httpserver/response"
)

type MeHandler struct {
	pool       *pgxpool.Pool
	repository *Repository
}

func NewMeHandler(pool *pgxpool.Pool, repository *Repository) *MeHandler {
	return &MeHandler{pool: pool, repository: repository}
}

func (handler *MeHandler) RegisterRoutes(router gin.IRouter) {
	router.GET("/me", authentication.RequireUser(), handler.getMe)
}

func (handler *MeHandler) getMe(context *gin.Context) {
	userID, _ := authentication.UserIDFrom(context)
	currentUser, isFound, err := handler.repository.FindByID(context.Request.Context(), handler.pool, userID)
	if err != nil {
		response.WriteError(context, err)
		return
	}
	if !isFound {
		response.WriteError(context, apierror.Unauthorized("Log in to continue"))
		return
	}
	response.WriteData(context, http.StatusOK, currentUser.Profile())
}

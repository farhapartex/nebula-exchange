package users

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"

	"nebula-exchange/backend/internal/auth/authentication"
	"nebula-exchange/backend/internal/platform/apierror"
	"nebula-exchange/backend/internal/platform/httpserver/response"
)

const currentUserContextKey = "current_user"

type AccountGuard struct {
	pool       *pgxpool.Pool
	repository *Repository
}

func NewAccountGuard(pool *pgxpool.Pool, repository *Repository) *AccountGuard {
	return &AccountGuard{pool: pool, repository: repository}
}

func (guard *AccountGuard) RequireStatus(allowedStatuses ...Status) gin.HandlerFunc {
	return func(context *gin.Context) {
		currentUser, isLoaded := guard.loadCurrentUser(context)
		if !isLoaded {
			return
		}
		for _, allowedStatus := range allowedStatuses {
			if currentUser.Status == allowedStatus {
				context.Next()
				return
			}
		}
		response.WriteError(context, NotActiveError(currentUser.Status))
	}
}

func (guard *AccountGuard) RequireAdmin() gin.HandlerFunc {
	return func(context *gin.Context) {
		currentUser, isLoaded := guard.loadCurrentUser(context)
		if !isLoaded {
			return
		}
		if !currentUser.IsAdmin {
			response.WriteError(context, apierror.Forbidden("Admin access is required"))
			return
		}
		context.Next()
	}
}

func CurrentUserFrom(context *gin.Context) (User, bool) {
	storedValue, isSet := context.Get(currentUserContextKey)
	if !isSet {
		return User{}, false
	}
	currentUser, isUser := storedValue.(User)
	return currentUser, isUser
}

func (guard *AccountGuard) loadCurrentUser(context *gin.Context) (User, bool) {
	if cachedUser, isCached := CurrentUserFrom(context); isCached {
		return cachedUser, true
	}
	userID, isAuthenticated := authentication.UserIDFrom(context)
	if !isAuthenticated {
		response.WriteError(context, apierror.Unauthorized("Log in to continue"))
		return User{}, false
	}
	currentUser, isFound, err := guard.repository.FindByID(context.Request.Context(), guard.pool, userID)
	if err != nil {
		response.WriteError(context, err)
		return User{}, false
	}
	if !isFound || !currentUser.IsActive || currentUser.Status == StatusBanned {
		response.WriteError(context, apierror.Unauthorized("Log in to continue"))
		return User{}, false
	}
	context.Set(currentUserContextKey, currentUser)
	return currentUser, true
}

func accountStatusMessage(status Status) string {
	switch status {
	case StatusPendingPayment:
		return "Pay the entry fee to unlock this feature"
	case StatusFrozen:
		return "Your account is frozen. You can view your account but not make changes"
	default:
		return "Your account can't do this right now"
	}
}

func NotActiveError(status Status) *apierror.Error {
	return apierror.New(http.StatusForbidden, apierror.CodeAccountNotActive, accountStatusMessage(status)).
		WithDetails(map[string]string{"status": string(status)})
}

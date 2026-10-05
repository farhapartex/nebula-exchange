package service

import (
	"net/http"

	"github.com/farhapartex/nebula-exchange/backend/internal/platform/apierror"
)

var (
	ErrInvalidActivationLink = apierror.NotFound("This activation link is invalid or has expired")
	ErrInvalidCredentials    = apierror.Unauthorized("Invalid email or password")
	ErrAccountNotActivated   = apierror.New(http.StatusForbidden, apierror.CodeAccountNotActivated,
		"Your account is not activated yet. Check your email for the activation link.")
	ErrAccountBanned  = apierror.Forbidden("This account is banned")
	ErrSessionExpired = apierror.Unauthorized("Your session has expired. Log in again.")
	ErrNotLoggedIn    = apierror.Unauthorized("Log in to continue")
)

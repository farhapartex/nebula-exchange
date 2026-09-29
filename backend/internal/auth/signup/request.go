package signup

import (
	"net/mail"
	"regexp"
	"strings"

	"nebula-exchange/backend/internal/platform/apierror"
)

var usernamePattern = regexp.MustCompile(`^[A-Za-z0-9_]{3,20}$`)

type Request struct {
	Email        string `json:"email" binding:"required,max=254"`
	Username     string `json:"username" binding:"required"`
	Password     string `json:"password" binding:"required,min=10,max=128"`
	AcceptsTerms bool   `json:"accepts_terms"`
}

func (request Request) normalized() Request {
	request.Email = strings.ToLower(strings.TrimSpace(request.Email))
	request.Username = strings.TrimSpace(request.Username)
	return request
}

func (request Request) validateBusinessRules() error {
	fieldErrors := map[string]string{}
	if !isValidEmailAddress(request.Email) {
		fieldErrors["email"] = "must be a valid email address"
	}
	if !usernamePattern.MatchString(request.Username) {
		fieldErrors["username"] = "must be 3 to 20 letters, numbers or underscores"
	}
	if !request.AcceptsTerms {
		fieldErrors["accepts_terms"] = "must be accepted"
	}
	if len(fieldErrors) > 0 {
		return apierror.ValidationFailed(fieldErrors)
	}
	return nil
}

func isValidEmailAddress(emailAddress string) bool {
	parsedAddress, err := mail.ParseAddress(emailAddress)
	return err == nil && parsedAddress.Address == emailAddress && strings.Contains(emailAddress[strings.LastIndex(emailAddress, "@"):], ".")
}

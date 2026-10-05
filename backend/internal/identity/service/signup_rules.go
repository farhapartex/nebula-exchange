package service

import (
	"net/mail"
	"regexp"
	"strings"
	"unicode/utf8"

	"github.com/farhapartex/nebula-exchange/backend/internal/platform/apierror"
)

const (
	MinimumPasswordLength = 10
	MaximumPasswordLength = 128
	maximumEmailLength    = 254
)

var usernamePattern = regexp.MustCompile(`^[A-Za-z0-9_]{3,20}$`)

func normalizeEmail(emailAddress string) string {
	return strings.ToLower(strings.TrimSpace(emailAddress))
}

func (input SignupInput) normalized() SignupInput {
	input.Email = normalizeEmail(input.Email)
	input.Username = strings.TrimSpace(input.Username)
	return input
}

func (input SignupInput) validate() error {
	fieldErrors := map[string]string{}
	if !isValidEmailAddress(input.Email) {
		fieldErrors["email"] = "must be a valid email address"
	}
	if !usernamePattern.MatchString(input.Username) {
		fieldErrors["username"] = "must be 3 to 20 letters, numbers or underscores"
	}
	passwordLength := utf8.RuneCountInString(input.Password)
	if passwordLength < MinimumPasswordLength || passwordLength > MaximumPasswordLength {
		fieldErrors["password"] = "must be 10 to 128 characters"
	}
	if !input.AcceptsTerms {
		fieldErrors["accepts_terms"] = "must be accepted"
	}
	if len(fieldErrors) > 0 {
		return apierror.ValidationFailed(fieldErrors)
	}
	return nil
}

func isValidEmailAddress(emailAddress string) bool {
	if len(emailAddress) > maximumEmailLength {
		return false
	}
	parsedAddress, err := mail.ParseAddress(emailAddress)
	if err != nil || parsedAddress.Address != emailAddress {
		return false
	}
	domain := emailAddress[strings.LastIndex(emailAddress, "@")+1:]
	return strings.Contains(domain, ".")
}

package service

import (
	"errors"
	"testing"

	"github.com/farhapartex/nebula-exchange/backend/internal/platform/apierror"
)

func validSignupInput() SignupInput {
	return SignupInput{Email: "  Boy@Street.Test ", Username: " the_boy ", Password: "streets-are-cold", AcceptsTerms: true}
}

func TestSignupInputIsNormalized(t *testing.T) {
	normalizedInput := validSignupInput().normalized()
	if normalizedInput.Email != "boy@street.test" || normalizedInput.Username != "the_boy" {
		t.Fatalf("got %+v", normalizedInput)
	}
	if err := normalizedInput.validate(); err != nil {
		t.Fatalf("expected a valid input, got %v", err)
	}
}

func TestSignupInputRejectsEachBrokenRule(t *testing.T) {
	brokenInputs := map[string]func(*SignupInput){
		"email":         func(input *SignupInput) { input.Email = "not-an-email" },
		"username":      func(input *SignupInput) { input.Username = "no spaces allowed" },
		"password":      func(input *SignupInput) { input.Password = "short" },
		"accepts_terms": func(input *SignupInput) { input.AcceptsTerms = false },
	}
	for fieldName, breakRule := range brokenInputs {
		t.Run(fieldName, func(t *testing.T) {
			input := validSignupInput().normalized()
			breakRule(&input)
			var apiError *apierror.Error
			if err := input.validate(); !errors.As(err, &apiError) || apiError.Code != apierror.CodeValidationFailed {
				t.Fatalf("expected VALIDATION_FAILED, got %v", err)
			}
			fieldErrors, _ := apiError.Details.(map[string]string)
			if _, hasField := fieldErrors[fieldName]; !hasField || len(fieldErrors) != 1 {
				t.Fatalf("expected only %s to fail, got %v", fieldName, fieldErrors)
			}
		})
	}
}

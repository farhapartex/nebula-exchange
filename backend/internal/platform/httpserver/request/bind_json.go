package request

import (
	"encoding/json"
	"errors"
	"io"

	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"
	"github.com/go-playground/validator/v10"

	"nebula-exchange/backend/internal/platform/apierror"
)

func BindJSON(context *gin.Context, destination any) error {
	err := context.ShouldBindWith(destination, binding.JSON)
	if err == nil {
		return nil
	}

	var validationErrors validator.ValidationErrors
	if errors.As(err, &validationErrors) {
		return apierror.ValidationFailed(describeFieldErrors(validationErrors))
	}

	var syntaxError *json.SyntaxError
	var typeError *json.UnmarshalTypeError
	switch {
	case errors.Is(err, io.EOF):
		return apierror.BadRequest("Request body is required")
	case errors.As(err, &typeError):
		return apierror.ValidationFailed(map[string]string{typeError.Field: "has the wrong type"})
	case errors.As(err, &syntaxError):
		return apierror.BadRequest("Request body is not valid JSON")
	default:
		return apierror.BadRequest("Request body could not be read")
	}
}

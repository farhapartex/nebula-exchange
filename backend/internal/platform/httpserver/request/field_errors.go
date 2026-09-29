package request

import (
	"fmt"
	"reflect"
	"strings"
	"sync"

	"github.com/gin-gonic/gin/binding"
	"github.com/go-playground/validator/v10"
)

var registerJSONFieldNamesOnce sync.Once

func RegisterJSONFieldNames() {
	registerJSONFieldNamesOnce.Do(func() {
		validatorEngine, isValidator := binding.Validator.Engine().(*validator.Validate)
		if !isValidator {
			return
		}
		validatorEngine.RegisterTagNameFunc(jsonFieldName)
	})
}

func jsonFieldName(field reflect.StructField) string {
	jsonName := strings.SplitN(field.Tag.Get("json"), ",", 2)[0]
	if jsonName == "-" {
		return ""
	}
	if jsonName == "" {
		return field.Name
	}
	return jsonName
}

func describeFieldErrors(validationErrors validator.ValidationErrors) map[string]string {
	fieldErrors := make(map[string]string, len(validationErrors))
	for _, fieldError := range validationErrors {
		fieldPath := strings.SplitN(fieldError.Namespace(), ".", 2)
		fieldName := fieldError.Field()
		if len(fieldPath) == 2 {
			fieldName = fieldPath[1]
		}
		fieldErrors[fieldName] = describeRule(fieldError)
	}
	return fieldErrors
}

func describeRule(fieldError validator.FieldError) string {
	switch fieldError.Tag() {
	case "required":
		return "is required"
	case "email":
		return "must be a valid email address"
	case "min":
		return fmt.Sprintf("must be at least %s", fieldError.Param())
	case "max":
		return fmt.Sprintf("must be at most %s", fieldError.Param())
	case "len":
		return fmt.Sprintf("must be exactly %s long", fieldError.Param())
	case "oneof":
		return fmt.Sprintf("must be one of: %s", fieldError.Param())
	case "gt":
		return fmt.Sprintf("must be greater than %s", fieldError.Param())
	case "gte":
		return fmt.Sprintf("must be %s or more", fieldError.Param())
	case "uuid", "uuid4":
		return "must be a valid UUID"
	default:
		return "is invalid"
	}
}

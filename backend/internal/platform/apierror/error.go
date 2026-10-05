package apierror

import (
	"errors"
	"net/http"
)

type Error struct {
	StatusCode int
	Code       Code
	Message    string
	Details    any
}

func (apiError *Error) Error() string {
	return string(apiError.Code) + ": " + apiError.Message
}

func (apiError *Error) WithDetails(details any) *Error {
	detailedError := *apiError
	detailedError.Details = details
	return &detailedError
}

func New(statusCode int, code Code, message string) *Error {
	return &Error{StatusCode: statusCode, Code: code, Message: message}
}

func ValidationFailed(fieldErrors map[string]string) *Error {
	return New(http.StatusUnprocessableEntity, CodeValidationFailed, "Some fields are invalid").WithDetails(fieldErrors)
}

func BadRequest(message string) *Error {
	return New(http.StatusBadRequest, CodeValidationFailed, message)
}

func Unauthorized(message string) *Error {
	return New(http.StatusUnauthorized, CodeUnauthorized, message)
}

func Forbidden(message string) *Error {
	return New(http.StatusForbidden, CodeForbidden, message)
}

func NotFound(message string) *Error {
	return New(http.StatusNotFound, CodeNotFound, message)
}

func Conflict(message string) *Error {
	return New(http.StatusConflict, CodeConflict, message)
}

func MethodNotAllowed() *Error {
	return New(http.StatusMethodNotAllowed, CodeMethodNotAllowed, "Method not allowed")
}

func ServiceUnavailable(message string) *Error {
	return New(http.StatusServiceUnavailable, CodeServiceUnavailable, message)
}

func Internal() *Error {
	return New(http.StatusInternalServerError, CodeInternalError, "Something went wrong")
}

type Convertible interface {
	APIError() *Error
}

func From(err error) *Error {
	var apiError *Error
	if errors.As(err, &apiError) {
		return apiError
	}
	var convertible Convertible
	if errors.As(err, &convertible) {
		return convertible.APIError()
	}
	return Internal()
}

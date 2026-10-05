package apierror

type Code string

const (
	CodeValidationFailed    Code = "VALIDATION_FAILED"
	CodeUnauthorized        Code = "UNAUTHORIZED"
	CodeForbidden           Code = "FORBIDDEN"
	CodeNotFound            Code = "NOT_FOUND"
	CodeMethodNotAllowed    Code = "METHOD_NOT_ALLOWED"
	CodeConflict            Code = "CONFLICT"
	CodeAccountNotActivated Code = "ACCOUNT_NOT_ACTIVATED"
	CodeInternalError       Code = "INTERNAL_ERROR"
	CodeServiceUnavailable  Code = "SERVICE_UNAVAILABLE"
)

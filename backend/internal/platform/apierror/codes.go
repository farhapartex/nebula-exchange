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
	CodeAccountNotActive    Code = "ACCOUNT_NOT_ACTIVE"
	CodeInsufficientFunds   Code = "INSUFFICIENT_FUNDS"
	CodeInsufficientItems   Code = "INSUFFICIENT_ITEMS"
	CodeMarketHalted        Code = "MARKET_HALTED"
	CodeBidTooLow           Code = "BID_TOO_LOW"
	CodeLimitExceeded       Code = "LIMIT_EXCEEDED"
	CodeWalletRequired      Code = "WALLET_REQUIRED"
	CodeTwoFactorRequired   Code = "TWO_FA_REQUIRED"
	CodeRateLimited         Code = "RATE_LIMITED"
	CodeEngineBusy          Code = "ENGINE_BUSY"
	CodeInternalError       Code = "INTERNAL_ERROR"
	CodeServiceUnavailable  Code = "SERVICE_UNAVAILABLE"
)

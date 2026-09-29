package ledger

import (
	"errors"
	"net/http"

	"nebula-exchange/backend/internal/platform/apierror"
)

var (
	ErrDuplicateJournal  = errors.New("journal already posted for this business event")
	ErrDuplicateHold     = errors.New("hold already exists for this reference")
	ErrUnbalancedJournal = errors.New("journal entries do not sum to zero per asset")
	ErrHoldNotAvailable  = errors.New("hold is not active or has less remaining than requested")
	ErrInvalidAmount     = errors.New("ledger amounts must be positive")
)

type InsufficientBalanceError struct {
	Asset Asset
}

func (insufficientBalance *InsufficientBalanceError) Error() string {
	if insufficientBalance.Asset.IsNC() {
		return "insufficient NC balance"
	}
	return "insufficient item balance"
}

func (insufficientBalance *InsufficientBalanceError) APIError() *apierror.Error {
	if insufficientBalance.Asset.IsNC() {
		return apierror.New(http.StatusConflict, apierror.CodeInsufficientFunds, "You don't have enough NC for this")
	}
	return apierror.New(http.StatusConflict, apierror.CodeInsufficientItems, "You don't have enough of this item").
		WithDetails(map[string]any{"item_id": int(insufficientBalance.Asset)})
}

func IsInsufficientBalance(err error) bool {
	var insufficientBalance *InsufficientBalanceError
	return errors.As(err, &insufficientBalance)
}

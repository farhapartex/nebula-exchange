package handler

import "strconv"

func formatCoinAmount(amount *int64) *string {
	if amount == nil {
		return nil
	}
	formattedAmount := strconv.FormatInt(*amount, 10)
	return &formattedAmount
}

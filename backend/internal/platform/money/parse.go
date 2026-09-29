package money

import (
	"errors"
	"strconv"
	"strings"
)

var ErrInvalidNCAmount = errors.New("NC amounts are whole numbers with up to 6 decimals")

const microDecimals = 6

func ParseNC(amountText string) (Micro, error) {
	wholePart, fractionPart, _ := strings.Cut(strings.TrimSpace(amountText), ".")
	if wholePart == "" || len(fractionPart) > microDecimals || strings.HasPrefix(wholePart, "-") || strings.HasPrefix(wholePart, "+") {
		return 0, ErrInvalidNCAmount
	}
	paddedFraction := fractionPart + strings.Repeat("0", microDecimals-len(fractionPart))
	microUnits, err := strconv.ParseInt(wholePart+paddedFraction, 10, 64)
	if err != nil {
		return 0, ErrInvalidNCAmount
	}
	return Micro(microUnits), nil
}

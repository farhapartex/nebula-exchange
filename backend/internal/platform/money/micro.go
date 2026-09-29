package money

import (
	"encoding/json"
	"fmt"
	"strconv"
)

type Micro int64

const MicroPerNC Micro = 1_000_000

func (amount Micro) MarshalJSON() ([]byte, error) {
	return json.Marshal(strconv.FormatInt(int64(amount), 10))
}

func (amount *Micro) UnmarshalJSON(encodedAmount []byte) error {
	var amountText string
	if err := json.Unmarshal(encodedAmount, &amountText); err != nil {
		return fmt.Errorf("amount must be a string of micro-units: %w", err)
	}
	parsedAmount, err := strconv.ParseInt(amountText, 10, 64)
	if err != nil {
		return fmt.Errorf("amount %q is not a whole number of micro-units", amountText)
	}
	*amount = Micro(parsedAmount)
	return nil
}

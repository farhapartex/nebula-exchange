package payments

import (
	"time"

	"nebula-exchange/backend/internal/platform/money"
)

const (
	cardPaymentLifetime   = 24 * time.Hour
	cryptoPaymentLifetime = 30 * time.Minute
	dailyCardTopupLimit   = 500 * money.MicroPerNC
	monthlyCardTopupLimit = 2000 * money.MicroPerNC
	microPerCent          = 10_000
)

var allowedCardTopupAmounts = map[money.Micro]bool{
	5 * money.MicroPerNC:   true,
	10 * money.MicroPerNC:  true,
	25 * money.MicroPerNC:  true,
	50 * money.MicroPerNC:  true,
	100 * money.MicroPerNC: true,
}

func lifetimeFor(method Method) time.Duration {
	if method == MethodCrypto {
		return cryptoPaymentLifetime
	}
	return cardPaymentLifetime
}

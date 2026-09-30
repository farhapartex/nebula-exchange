package payments

import (
	"fmt"
	"strconv"

	"nebula-exchange/backend/internal/notify/inapp"
	"nebula-exchange/backend/internal/payments/paymentsstore"
	"nebula-exchange/backend/internal/platform/money"
	"nebula-exchange/backend/internal/purpose"
)

var successTitles = map[purpose.Kind]string{
	purpose.KindEntryFee:        "Welcome aboard, your account is active",
	purpose.KindTopup:           "Top-up complete",
	purpose.KindShopPurchase:    "Shop purchase complete",
	purpose.KindUpgradePurchase: "Upgrade complete",
}

func formatNC(amount money.Micro) string {
	wholePart := int64(amount) / int64(money.MicroPerNC)
	centsPart := (int64(amount) % int64(money.MicroPerNC)) / 10_000
	return strconv.FormatInt(wholePart, 10) + "." + fmt.Sprintf("%02d", centsPart)
}

func paymentNotice(payment paymentsstore.Payment, credited money.Micro, purposeFailure string) inapp.Notice {
	notice := inapp.Notice{
		UserID:        payment.UserID,
		Link:          "/payment/result?id=" + payment.ID.String(),
		AlsoSendEmail: true,
	}
	if purposeFailure != "" {
		notice.Kind = inapp.KindPaymentNeedsCare
		notice.Title = "Payment received, NC kept in your balance"
		notice.Body = fmt.Sprintf("we received %s NC, but it couldn't be used for this order (%s). The NC is in your balance.", formatNC(credited), purposeFailure)
		return notice
	}
	notice.Kind = inapp.KindPaymentSucceeded
	notice.Title = successTitles[purpose.Kind(payment.Purpose)]
	notice.Body = fmt.Sprintf("your payment of %s NC went through.", formatNC(credited))
	return notice
}

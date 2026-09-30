import { describe, expect, it } from "vitest";

import { describePaymentChoices, preferredPaymentChoice } from "@/features/payments/payment-method-rules";

describe("describePaymentChoices", () => {
  it("never offers the NC balance for the entry fee or top-ups", () => {
    expect(describePaymentChoices("entry_fee", 5_000_000n, 0n).map((availability) => availability.choice)).toEqual([
      "card",
      "wallet",
    ]);
    expect(describePaymentChoices("topup", 10_000_000n, null).map((availability) => availability.choice)).toEqual([
      "card",
      "wallet",
    ]);
  });

  it("offers the balance in the shop only when it covers the price", () => {
    const affordable = describePaymentChoices("shop", 2_000_000n, 2_000_000n);
    expect(affordable[0]).toEqual({ choice: "balance", isAllowed: true, unavailableReason: null });
    expect(preferredPaymentChoice(affordable)).toBe("balance");

    const tooExpensive = describePaymentChoices("shop", 2_000_001n, 2_000_000n);
    expect(tooExpensive[0]).toEqual({ choice: "balance", isAllowed: false, unavailableReason: "Not enough NC" });
    expect(preferredPaymentChoice(tooExpensive)).toBe("card");
  });

  it("keeps the wallet disabled until wallets ship", () => {
    const walletChoice = describePaymentChoices("shop", 1n, 10n).find(
      (availability) => availability.choice === "wallet",
    );
    expect(walletChoice?.isAllowed).toBe(false);
  });
});

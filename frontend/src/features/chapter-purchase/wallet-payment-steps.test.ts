import { describe, expect, it } from "vitest";

import { walletPaymentStepState } from "@/features/chapter-purchase/wallet-payment-steps";

describe("walletPaymentStepState", () => {
  it("marks earlier steps done, the current one active and later ones waiting", () => {
    expect(walletPaymentStepState("PREPARING", "PAYING")).toBe("DONE");
    expect(walletPaymentStepState("PAYING", "PAYING")).toBe("ACTIVE");
    expect(walletPaymentStepState("SUBMITTED", "PAYING")).toBe("WAITING");
  });
});

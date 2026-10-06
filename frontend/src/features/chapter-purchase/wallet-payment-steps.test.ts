import { describe, expect, it } from "vitest";

import { walletPaymentStepState, walletPaymentStepsFor } from "@/features/chapter-purchase/wallet-payment-steps";

describe("wallet payment steps", () => {
  it("asks for a USDC approval but not for ETH", () => {
    expect(walletPaymentStepsFor("USDC")).toContain("APPROVING");
    expect(walletPaymentStepsFor("ETH")).not.toContain("APPROVING");
  });

  it("marks earlier steps done, the current one active and later ones waiting", () => {
    const steps = walletPaymentStepsFor("USDC");
    expect(walletPaymentStepState(steps, "PREPARING", "PAYING")).toBe("DONE");
    expect(walletPaymentStepState(steps, "PAYING", "PAYING")).toBe("ACTIVE");
    expect(walletPaymentStepState(steps, "SUBMITTED", "PAYING")).toBe("WAITING");
  });
});

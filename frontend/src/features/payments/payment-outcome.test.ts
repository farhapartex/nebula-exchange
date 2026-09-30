import { describe, expect, it } from "vitest";

import type { Payment } from "@/features/payments/api/payments-api";
import { describePaymentOutcome, isPaymentSettled } from "@/features/payments/payment-outcome";

const pendingPayment: Payment = {
  id: "payment-1",
  purpose: "ENTRY_FEE",
  method: "card",
  status: "PENDING",
  amount: "5000000",
  credited: null,
  sku: null,
  upgrade_id: null,
  quantity: 1,
  purpose_status: "PENDING",
  purpose_failure_code: null,
  checkout_url: null,
  expires_at: "2026-10-01T00:00:00Z",
  succeeded_at: null,
  created_at: "2026-09-30T00:00:00Z",
};

describe("describePaymentOutcome", () => {
  it("keeps polling until the timeout", () => {
    expect(describePaymentOutcome(undefined, false)).toBe("processing");
    expect(describePaymentOutcome(pendingPayment, false)).toBe("processing");
    expect(describePaymentOutcome(pendingPayment, true)).toBe("timed_out");
    expect(isPaymentSettled(pendingPayment)).toBe(false);
  });

  it("separates a fully applied payment from one that only credited NC", () => {
    const succeeded = { ...pendingPayment, status: "SUCCEEDED" as const, credited: "5000000" };
    expect(describePaymentOutcome({ ...succeeded, purpose_status: "APPLIED" }, false)).toBe("completed");
    expect(describePaymentOutcome({ ...succeeded, purpose_status: "FAILED" }, true)).toBe("credited_only");
    expect(isPaymentSettled({ ...succeeded, purpose_status: "APPLIED" })).toBe(true);
  });

  it("reports failed and expired payments", () => {
    expect(describePaymentOutcome({ ...pendingPayment, status: "EXPIRED" }, false)).toBe("failed");
    expect(describePaymentOutcome({ ...pendingPayment, status: "FAILED" }, false)).toBe("failed");
  });
});

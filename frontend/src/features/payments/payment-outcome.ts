import type { Payment } from "@/features/payments/api/payments-api";

export type PaymentOutcome = "processing" | "completed" | "credited_only" | "failed" | "timed_out";

export function describePaymentOutcome(payment: Payment | undefined, hasPollingTimedOut: boolean): PaymentOutcome {
  if (!payment || payment.status === "PENDING") {
    return hasPollingTimedOut ? "timed_out" : "processing";
  }
  if (payment.status === "FAILED" || payment.status === "EXPIRED") {
    return "failed";
  }
  if (payment.status === "SUCCEEDED" && payment.purpose_status === "APPLIED") {
    return "completed";
  }
  if (payment.status === "SUCCEEDED" && payment.purpose_status === "FAILED") {
    return "credited_only";
  }
  return "processing";
}

const finalUnpaidStatuses = new Set<Payment["status"]>(["FAILED", "EXPIRED"]);

export function isPaymentSettled(payment: Payment | undefined): boolean {
  if (payment === undefined || payment.status === "PENDING") {
    return false;
  }
  return finalUnpaidStatuses.has(payment.status) || payment.purpose_status !== "PENDING";
}

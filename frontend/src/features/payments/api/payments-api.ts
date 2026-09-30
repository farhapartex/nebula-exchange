import { requestData, requestList } from "@/lib/api/api-client";
import type { PaginationParameters } from "@/lib/api/api-types";

export type PaymentPurpose = "ENTRY_FEE" | "TOPUP" | "SHOP_PURCHASE" | "UPGRADE_PURCHASE";
export type PaymentMethod = "card" | "crypto";
export type PaymentStatus = "PENDING" | "SUCCEEDED" | "FAILED" | "EXPIRED" | "REFUNDED" | "DISPUTED";
export type PurposeStatus = "PENDING" | "APPLIED" | "FAILED";

export type Payment = {
  id: string;
  purpose: PaymentPurpose;
  method: PaymentMethod;
  status: PaymentStatus;
  amount: string;
  credited: string | null;
  sku: string | null;
  upgrade_id: string | null;
  quantity: number;
  purpose_status: PurposeStatus;
  purpose_failure_code: string | null;
  checkout_url: string | null;
  expires_at: string;
  succeeded_at: string | null;
  created_at: string;
};

export type CreatePaymentRequest = {
  purpose: PaymentPurpose;
  method: PaymentMethod;
  amount_nc?: string;
  sku?: string;
  upgrade_id?: string;
  quantity?: number;
};

export const paymentQueryKey = (paymentID: string) => ["payments", paymentID] as const;

export function createPayment(createRequest: CreatePaymentRequest, idempotencyKey: string): Promise<Payment> {
  return requestData<Payment>("/payments", { method: "POST", body: createRequest, idempotencyKey });
}

export function fetchPayment(paymentID: string): Promise<Payment> {
  return requestData<Payment>(`/payments/${paymentID}`);
}

export function listPayments(pagination: PaginationParameters) {
  return requestList<Payment>("/payments", pagination);
}

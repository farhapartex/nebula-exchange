import { http } from "msw";

import type { CreatePaymentRequest, Payment } from "@/features/payments/api/payments-api";
import { buildApiUrl } from "@/lib/api/api-config";
import { mockDataResponse, mockErrorResponse, mockListResponse, simulateLatency } from "@/mocks/utils/mock-responses";

const mockSettlementDelayInMilliseconds = 5_000;
const mockPayments = new Map<string, Payment>();

const mockPriceBySku: Record<string, string> = {
  "fuel-cell": "200000",
  scout: "2000000",
  "drill-t1": "1000000",
  "starter-bundle": "3500000",
};

function priceOf(createRequest: CreatePaymentRequest): string {
  if (createRequest.purpose === "ENTRY_FEE") {
    return "5000000";
  }
  if (createRequest.purpose === "TOPUP") {
    return `${createRequest.amount_nc ?? "0"}000000`;
  }
  const unitPrice = BigInt(mockPriceBySku[createRequest.sku ?? ""] ?? "0");
  return (unitPrice * BigInt(createRequest.quantity ?? 1)).toString();
}

function settleWhenDue(payment: Payment): Payment {
  const isDue = Date.now() - new Date(payment.created_at).getTime() >= mockSettlementDelayInMilliseconds;
  if (payment.status !== "PENDING" || !isDue) {
    return payment;
  }
  const settledPayment: Payment = {
    ...payment,
    status: "SUCCEEDED",
    credited: payment.amount,
    purpose_status: "APPLIED",
    succeeded_at: new Date().toISOString(),
  };
  mockPayments.set(payment.id, settledPayment);
  return settledPayment;
}

export const paymentsHandlers = [
  http.post(buildApiUrl("/payments"), async ({ request }) => {
    await simulateLatency();
    const createRequest = (await request.json()) as CreatePaymentRequest;
    if (createRequest.method === "crypto") {
      return mockErrorResponse(422, "WALLET_REQUIRED", "Link a wallet before paying with USDC");
    }
    const paymentID = crypto.randomUUID();
    const createdAt = new Date();
    const payment: Payment = {
      id: paymentID,
      purpose: createRequest.purpose,
      method: createRequest.method,
      status: "PENDING",
      amount: priceOf(createRequest),
      credited: null,
      sku: createRequest.sku ?? null,
      quantity: createRequest.quantity ?? 1,
      purpose_status: "PENDING",
      purpose_failure_code: null,
      checkout_url: `${window.location.origin}/payment/result?id=${paymentID}`,
      expires_at: new Date(createdAt.getTime() + 24 * 3_600_000).toISOString(),
      succeeded_at: null,
      created_at: createdAt.toISOString(),
    };
    mockPayments.set(paymentID, payment);
    return mockDataResponse(payment, 201);
  }),
  http.get(buildApiUrl("/payments/:paymentID"), async ({ params }) => {
    await simulateLatency(150);
    const payment = mockPayments.get(String(params.paymentID));
    return payment
      ? mockDataResponse(settleWhenDue(payment))
      : mockErrorResponse(404, "NOT_FOUND", "This payment does not exist");
  }),
  http.get(buildApiUrl("/payments"), async ({ request }) => {
    await simulateLatency();
    const newestFirst = [...mockPayments.values()].map(settleWhenDue).reverse();
    return mockListResponse(newestFirst, new URL(request.url));
  }),
];

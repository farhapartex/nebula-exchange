import { requestData } from "@/lib/api/api-client";

export type CheckoutSessionStatus = "OPEN" | "PAID" | "EXPIRED";

export type CreatedCheckoutSession = {
  id: string;
  checkout_url: string;
  expires_at: string;
};

export type CheckoutSessionState = {
  id: string;
  status: CheckoutSessionStatus;
};

export type CheckoutSessionRequest = {
  plan_id: string;
  chapter_count: number;
};

export function createCheckoutSession(checkoutRequest: CheckoutSessionRequest): Promise<CreatedCheckoutSession> {
  return requestData<CreatedCheckoutSession>("/checkout-sessions", { method: "POST", body: checkoutRequest });
}

export function fetchCheckoutSession(checkoutSessionID: string): Promise<CheckoutSessionState> {
  return requestData<CheckoutSessionState>(`/checkout-sessions/${encodeURIComponent(checkoutSessionID)}`);
}

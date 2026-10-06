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

export type PaymentMethod = "CARD" | "WALLET";

export type CheckoutSessionRequest = {
  plan_id: string;
  chapter_count: number;
};

export type CreatedWalletCheckout = {
  id: string;
  payment_reference: `0x${string}`;
  amount_units: string;
  token_address: `0x${string}`;
  vault_address: `0x${string}`;
  chain_id: number;
  expires_at: string;
};

export function createCheckoutSession(checkoutRequest: CheckoutSessionRequest): Promise<CreatedCheckoutSession> {
  return requestData<CreatedCheckoutSession>("/checkout-sessions", {
    method: "POST",
    body: { ...checkoutRequest, payment_method: "CARD" satisfies PaymentMethod },
  });
}

export function createWalletCheckout(checkoutRequest: CheckoutSessionRequest): Promise<CreatedWalletCheckout> {
  return requestData<CreatedWalletCheckout>("/checkout-sessions", {
    method: "POST",
    body: { ...checkoutRequest, payment_method: "WALLET" satisfies PaymentMethod },
  });
}

export function fetchCheckoutSession(checkoutSessionID: string): Promise<CheckoutSessionState> {
  return requestData<CheckoutSessionState>(`/checkout-sessions/${encodeURIComponent(checkoutSessionID)}`);
}

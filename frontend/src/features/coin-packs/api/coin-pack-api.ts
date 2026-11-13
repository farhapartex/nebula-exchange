import { requestAllPages, requestData } from "@/lib/api/api-client";
import type { CreatedCheckoutSession } from "@/features/subscriptions/api/checkout-session-api";

export type CoinPack = {
  id: string;
  name: string;
  coins: string;
  price_cents: string;
};

export type CoinPackCheckoutRequest = {
  coin_pack_id: string;
  return_path: string;
};

export const coinPacksQueryKey = ["coin-packs"] as const;

export function fetchCoinPacks(): Promise<CoinPack[]> {
  return requestAllPages<CoinPack>("/coin-packs");
}

export function createCoinPackCheckout(checkoutRequest: CoinPackCheckoutRequest): Promise<CreatedCheckoutSession> {
  return requestData<CreatedCheckoutSession>("/checkout-sessions", {
    method: "POST",
    body: { ...checkoutRequest, payment_method: "CARD" },
  });
}

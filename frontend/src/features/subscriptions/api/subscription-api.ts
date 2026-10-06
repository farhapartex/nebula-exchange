import { requestAllPages } from "@/lib/api/api-client";
import type { PlanKind } from "@/features/chapter-purchase/api/plan-api";
import type { WalletPaymentAsset } from "@/features/chapter-purchase/wallet-payment-asset";
import type { PaymentMethod } from "@/features/subscriptions/api/checkout-session-api";

export type SubscriptionStatus = "PAID" | "REFUNDED" | "DISPUTED";

export type SubscriptionChapter = {
  id: string;
  number: number;
  title: string;
};

export type WalletPaymentDetails = {
  asset: WalletPaymentAsset;
  amount_units: string;
  payer_address: string;
  transaction_hash: string;
};

export type Subscription = {
  id: string;
  plan_id: string;
  plan_name: string;
  plan_kind: PlanKind;
  status: SubscriptionStatus;
  chapters: SubscriptionChapter[];
  subtotal_cents: string;
  discount_percent: number;
  discount_cents: string;
  total_cents: string;
  paid_at: string;
  refunded_at: string | null;
  payment_method?: PaymentMethod;
  wallet_payment?: WalletPaymentDetails | null;
};

export const subscriptionsQueryKey = ["subscriptions"] as const;

export function fetchSubscriptions(): Promise<Subscription[]> {
  return requestAllPages<Subscription>("/subscriptions");
}

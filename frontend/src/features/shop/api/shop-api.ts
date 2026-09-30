import { requestData } from "@/lib/api/api-client";

export type CompletedPurchase = {
  id: string;
  sku: string;
  quantity: number;
  total: string;
  journal_id: string;
};

export function purchaseWithBalance(sku: string, quantity: number, idempotencyKey: string): Promise<CompletedPurchase> {
  return requestData<CompletedPurchase>(`/shop-items/${sku}/purchases`, {
    method: "POST",
    body: { quantity },
    idempotencyKey,
  });
}

import { requestAllPages } from "@/lib/api/api-client";

export type PurchasableChapter = {
  id: string;
  number: number;
  title: string;
  price_cents: string;
  is_owned: boolean;
};

export const purchasableChaptersQueryKey = ["chapters"] as const;

export function fetchPurchasableChapters(): Promise<PurchasableChapter[]> {
  return requestAllPages<PurchasableChapter>("/chapters");
}

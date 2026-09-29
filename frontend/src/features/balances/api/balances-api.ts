import { requestData } from "@/lib/api/api-client";

export type NcBucket = "card" | "earned_pending" | "earned" | "crypto";

export type BucketBalance = {
  bucket: NcBucket;
  available: string;
  held: string;
};

export type BalanceSummary = {
  total: string;
  available: string;
  held: string;
  withdrawable: string;
  buckets: BucketBalance[];
};

export const balancesQueryKey = ["me", "balances"] as const;

export function fetchBalances(): Promise<BalanceSummary> {
  return requestData<BalanceSummary>("/me/balances");
}

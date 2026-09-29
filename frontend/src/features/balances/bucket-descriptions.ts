import type { NcBucket } from "@/features/balances/api/balances-api";

export type BucketDescription = {
  label: string;
  explanation: string;
  isWithdrawable: boolean;
};

export const bucketDescriptions: Record<NcBucket, BucketDescription> = {
  card: { label: "Card", explanation: "Bought by card. Spendable, never withdrawable.", isWithdrawable: false },
  earned_pending: {
    label: "Earned (pending)",
    explanation: "Sales from the last 72 hours. Withdrawable once they settle.",
    isWithdrawable: false,
  },
  earned: { label: "Earned", explanation: "Settled sales and auction proceeds.", isWithdrawable: true },
  crypto: { label: "Crypto", explanation: "Paid or deposited in USDC.", isWithdrawable: true },
};

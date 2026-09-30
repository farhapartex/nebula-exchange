import type { Tone } from "@/components/ui/tone";
import type { PaymentMethod, PaymentPurpose, PaymentStatus } from "@/features/payments/api/payments-api";

export const paymentPurposeLabels: Record<PaymentPurpose, string> = {
  ENTRY_FEE: "Entry fee",
  TOPUP: "Top-up",
  SHOP_PURCHASE: "Shop purchase",
  UPGRADE_PURCHASE: "Upgrade",
};

export const paymentMethodLabels: Record<PaymentMethod, string> = {
  card: "Card",
  crypto: "USDC",
};

export const paymentStatusPresentation: Record<PaymentStatus, { label: string; tone: Tone }> = {
  PENDING: { label: "Pending", tone: "info" },
  SUCCEEDED: { label: "Succeeded", tone: "success" },
  FAILED: { label: "Failed", tone: "danger" },
  EXPIRED: { label: "Expired", tone: "neutral" },
  REFUNDED: { label: "Refunded", tone: "warning" },
  DISPUTED: { label: "Disputed", tone: "danger" },
};

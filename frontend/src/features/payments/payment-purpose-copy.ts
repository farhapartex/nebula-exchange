import type { PaymentPurpose } from "@/features/payments/api/payments-api";

export type PurposeCopy = {
  successTitle: string;
  successDescription: string;
  continueHref: string;
  continueLabel: string;
  retryHref: string;
};

export const purposeCopy: Record<PaymentPurpose, PurposeCopy> = {
  ENTRY_FEE: {
    successTitle: "Welcome aboard, pilot",
    successDescription: "Your account is active and your starter pack is in your hangar.",
    continueHref: "/hangar",
    continueLabel: "Go to the hangar",
    retryHref: "/onboarding/pay",
  },
  TOPUP: {
    successTitle: "Top-up complete",
    successDescription: "The NC is in your card balance and ready to spend.",
    continueHref: "/wallet",
    continueLabel: "Open wallet",
    retryHref: "/wallet",
  },
  SHOP_PURCHASE: {
    successTitle: "Purchase complete",
    successDescription: "Your items are in your inventory.",
    continueHref: "/inventory",
    continueLabel: "Open inventory",
    retryHref: "/shop",
  },
};

export const purposeFailureMessages: Record<string, string> = {
  ACCOUNT_NOT_PENDING_PAYMENT: "Your account was not waiting for the entry fee, so the NC stayed in your balance.",
  ACCOUNT_NOT_ACTIVE: "Your account can't make purchases right now, so the NC stayed in your balance.",
  INSUFFICIENT_FUNDS: "The payment arrived but didn't cover the price, so the NC stayed in your balance.",
  SKU_UNAVAILABLE: "That item is no longer sold, so the NC stayed in your balance.",
};

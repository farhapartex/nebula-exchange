"use client";

import Link from "next/link";
import { CircleCheck, Clock, TriangleAlert } from "lucide-react";
import type { ReactNode } from "react";

import { Button } from "@/components/ui/button";
import { Spinner } from "@/components/ui/spinner";
import type {
  CheckoutConfirmationState,
  NetworkConfirmationProgress,
} from "@/features/subscriptions/use-checkout-confirmation";
import { cn } from "@/utils/class-names";

type BannerContent = {
  icon: ReactNode;
  title: string;
  message: string;
  toneClassName: string;
};

const bannerContentByState: Record<CheckoutConfirmationState, BannerContent> = {
  CONFIRMING: {
    icon: <Spinner className="size-5" label="Confirming payment" />,
    title: "Confirming your payment",
    message: "Stripe is telling us about your payment. This takes a few seconds.",
    toneClassName: "border-border-strong bg-surface-raised/60 text-foreground",
  },
  PAID: {
    icon: <CircleCheck className="size-5" aria-hidden="true" />,
    title: "Payment received",
    message: "Your chapters are unlocked. Every level in them is ready to play.",
    toneClassName: "border-up/40 bg-up/10 text-up",
  },
  EXPIRED: {
    icon: <TriangleAlert className="size-5" aria-hidden="true" />,
    title: "This checkout was not completed",
    message: "No money was taken. You can start a new payment from the fight page.",
    toneClassName: "border-down/40 bg-down-soft/20 text-down",
  },
  DELAYED: {
    icon: <Clock className="size-5" aria-hidden="true" />,
    title: "Still waiting for Stripe",
    message: "Your payment is taking longer to confirm than usual. Your chapters unlock as soon as it arrives.",
    toneClassName: "border-border-strong bg-surface-raised/60 text-foreground",
  },
  FAILED: {
    icon: <TriangleAlert className="size-5" aria-hidden="true" />,
    title: "We could not check this payment",
    message: "Reload the page in a moment. If you were charged, your chapters unlock automatically.",
    toneClassName: "border-down/40 bg-down-soft/20 text-down",
  },
};

type CheckoutConfirmationBannerProps = {
  confirmationState: CheckoutConfirmationState;
  networkProgress: NetworkConfirmationProgress | null;
};

function describeNetworkProgress(networkProgress: NetworkConfirmationProgress): string {
  return `Your payment is on the network. Confirmations: ${networkProgress.confirmations} of ${networkProgress.requiredConfirmations}.`;
}

export function CheckoutConfirmationBanner({ confirmationState, networkProgress }: CheckoutConfirmationBannerProps) {
  const bannerContent = bannerContentByState[confirmationState];
  const message =
    confirmationState === "CONFIRMING" && networkProgress
      ? describeNetworkProgress(networkProgress)
      : bannerContent.message;
  return (
    <div
      role="status"
      className={cn("flex flex-wrap items-center gap-4 rounded-2xl border p-5", bannerContent.toneClassName)}
    >
      <span className="shrink-0">{bannerContent.icon}</span>
      <div className="min-w-0 flex-1">
        <p className="font-medium">{bannerContent.title}</p>
        <p className="mt-0.5 text-sm text-muted">{message}</p>
      </div>
      {confirmationState === "PAID" && (
        <Button asChild size="sm">
          <Link href="/fight">Play now</Link>
        </Button>
      )}
    </div>
  );
}

"use client";

import Link from "next/link";
import { useSearchParams } from "next/navigation";
import { useQuery } from "@tanstack/react-query";

import { Button } from "@/components/ui/button";
import { ErrorState } from "@/components/ui/error-state";
import { Skeleton } from "@/components/ui/skeleton";
import { HubPanel } from "@/features/fight-hub/hub-panel";
import { fetchSubscriptions, subscriptionsQueryKey } from "@/features/subscriptions/api/subscription-api";
import { CheckoutConfirmationBanner } from "@/features/subscriptions/checkout-confirmation-banner";
import { checkoutSessionQueryParameter } from "@/features/subscriptions/checkout-return-query";
import { SubscriptionCard } from "@/features/subscriptions/subscription-card";
import { useCheckoutConfirmation } from "@/features/subscriptions/use-checkout-confirmation";

export function SubscriptionView() {
  const checkoutSessionID = useSearchParams().get(checkoutSessionQueryParameter);
  const checkoutConfirmation = useCheckoutConfirmation(checkoutSessionID);
  const subscriptionsQuery = useQuery({ queryKey: subscriptionsQueryKey, queryFn: fetchSubscriptions });

  return (
    <div className="space-y-6">
      <div>
        <h1 className="font-display text-3xl tracking-[0.06em] text-foreground">Your chapters</h1>
        <p className="mt-1 text-sm text-muted">Every chapter purchase you have made, newest first.</p>
      </div>

      {checkoutConfirmation && (
        <CheckoutConfirmationBanner
          confirmationState={checkoutConfirmation.state}
          networkProgress={checkoutConfirmation.networkProgress}
        />
      )}

      {subscriptionsQuery.isError ? (
        <ErrorState
          message="Your purchases could not be loaded. Try again."
          onRetry={() => void subscriptionsQuery.refetch()}
          isRetrying={subscriptionsQuery.isFetching}
        />
      ) : !subscriptionsQuery.data ? (
        <div className="space-y-4">
          <Skeleton className="h-36 rounded-2xl" />
          <Skeleton className="h-36 rounded-2xl" />
        </div>
      ) : subscriptionsQuery.data.length === 0 ? (
        <HubPanel className="text-center">
          <p className="font-medium text-foreground">No purchases yet</p>
          <p className="mt-1 text-sm text-muted">
            Unlock the next chapter from the fight page to keep the story going.
          </p>
          <Button asChild size="sm" className="mt-4">
            <Link href="/fight">Go to fight</Link>
          </Button>
        </HubPanel>
      ) : (
        <div className="space-y-4">
          {subscriptionsQuery.data.map((subscription) => (
            <SubscriptionCard key={subscription.id} subscription={subscription} />
          ))}
        </div>
      )}
    </div>
  );
}

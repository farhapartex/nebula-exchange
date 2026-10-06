import { Suspense } from "react";
import type { Metadata } from "next";

import { PageContainer } from "@/components/layout/page-container";
import { SubscriptionView } from "@/features/subscriptions/subscription-view";

export const metadata: Metadata = {
  title: "Subscription",
};

export default function SubscriptionPage() {
  return (
    <PageContainer className="max-w-3xl py-8 sm:py-10">
      <Suspense>
        <SubscriptionView />
      </Suspense>
    </PageContainer>
  );
}

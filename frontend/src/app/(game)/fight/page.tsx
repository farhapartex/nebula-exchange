import { Suspense } from "react";
import type { Metadata } from "next";

import { PageContainer } from "@/components/layout/page-container";
import { FightHubView } from "@/features/fight-hub/fight-hub-view";
import { CheckoutCancelledNotice } from "@/features/subscriptions/checkout-cancelled-notice";

export const metadata: Metadata = {
  title: "Fight",
};

export default function FightPage() {
  return (
    <PageContainer className="py-8 sm:py-10">
      <Suspense>
        <CheckoutCancelledNotice />
      </Suspense>
      <FightHubView />
    </PageContainer>
  );
}

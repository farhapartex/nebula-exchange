import { Suspense } from "react";
import type { Metadata } from "next";

import { PageContainer } from "@/components/layout/page-container";
import { MarketView } from "@/features/market/market-view";

export const metadata: Metadata = {
  title: "Market",
};

export default function MarketPage() {
  return (
    <PageContainer className="py-8 sm:py-10">
      <Suspense>
        <MarketView />
      </Suspense>
    </PageContainer>
  );
}

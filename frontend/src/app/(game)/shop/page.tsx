import type { Metadata } from "next";

import { navigationLinkFor } from "@/components/app-shell/navigation-links";
import { PageContainer } from "@/components/layout/page-container";
import { SectionHeader } from "@/components/layout/section-header";
import { ShopView } from "@/features/shop/shop-view";

export const metadata: Metadata = {
  title: "Shop",
};

export default function ShopPage() {
  const navigationLink = navigationLinkFor("shop");
  return (
    <PageContainer className="py-8 sm:py-10">
      <SectionHeader title={navigationLink.label} description={navigationLink.description} icon={navigationLink.icon} />
      <ShopView />
    </PageContainer>
  );
}

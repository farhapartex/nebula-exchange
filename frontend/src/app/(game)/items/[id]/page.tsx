import type { Metadata } from "next";

import { PageContainer } from "@/components/layout/page-container";
import { ItemDetailView } from "@/features/item-detail/item-detail-view";

export const metadata: Metadata = {
  title: "Item",
};

export default async function ItemDetailPage({ params }: { params: Promise<{ id: string }> }) {
  const { id } = await params;
  return (
    <PageContainer className="py-8 sm:py-10">
      <ItemDetailView itemID={Number(id)} />
    </PageContainer>
  );
}

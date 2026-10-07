import type { Metadata } from "next";

import { PageContainer } from "@/components/layout/page-container";
import { FightHistoryView } from "@/features/fight-history/fight-history-view";

export const metadata: Metadata = {
  title: "Fight history",
};

export default function FightHistoryPage() {
  return (
    <PageContainer className="max-w-4xl py-8 sm:py-10">
      <FightHistoryView />
    </PageContainer>
  );
}

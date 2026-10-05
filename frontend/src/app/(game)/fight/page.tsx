import type { Metadata } from "next";

import { PageContainer } from "@/components/layout/page-container";
import { FightHubView } from "@/features/fight-hub/fight-hub-view";

export const metadata: Metadata = {
  title: "Fight",
};

export default function FightPage() {
  return (
    <PageContainer className="py-8 sm:py-10">
      <FightHubView />
    </PageContainer>
  );
}

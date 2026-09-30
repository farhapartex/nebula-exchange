import type { Metadata } from "next";

import { PageContainer } from "@/components/layout/page-container";
import { HangarView } from "@/features/hangar/hangar-view";

export const metadata: Metadata = {
  title: "Hangar",
};

export default function HangarPage() {
  return (
    <PageContainer className="py-8 sm:py-10">
      <HangarView />
    </PageContainer>
  );
}

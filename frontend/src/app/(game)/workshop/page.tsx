import type { Metadata } from "next";

import { SectionPlaceholder } from "@/components/app-shell/section-placeholder";

export const metadata: Metadata = {
  title: "Workshop",
};

export default function WorkshopPage() {
  return <SectionPlaceholder section="workshop" />;
}

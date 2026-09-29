import type { Metadata } from "next";

import { SectionPlaceholder } from "@/components/app-shell/section-placeholder";

export const metadata: Metadata = {
  title: "History",
};

export default function HistoryPage() {
  return <SectionPlaceholder section="history" />;
}

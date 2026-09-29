import type { Metadata } from "next";

import { SectionPlaceholder } from "@/components/app-shell/section-placeholder";

export const metadata: Metadata = {
  title: "Exchange",
};

export default function ExchangePage() {
  return <SectionPlaceholder section="exchange" />;
}

import type { Metadata } from "next";

import { SectionPlaceholder } from "@/components/app-shell/section-placeholder";

export const metadata: Metadata = {
  title: "Auctions",
};

export default function AuctionsPage() {
  return <SectionPlaceholder section="auctions" />;
}

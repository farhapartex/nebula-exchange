import type { Metadata } from "next";

import { SectionPlaceholder } from "@/components/app-shell/section-placeholder";

export const metadata: Metadata = {
  title: "Inventory",
};

export default function InventoryPage() {
  return <SectionPlaceholder section="inventory" />;
}

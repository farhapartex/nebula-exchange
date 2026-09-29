import type { Metadata } from "next";

import { SectionPlaceholder } from "@/components/app-shell/section-placeholder";

export const metadata: Metadata = {
  title: "Shop",
};

export default function ShopPage() {
  return <SectionPlaceholder section="shop" />;
}

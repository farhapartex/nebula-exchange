import type { Metadata } from "next";

import { SectionPlaceholder } from "@/components/app-shell/section-placeholder";

export const metadata: Metadata = {
  title: "Wallet",
};

export default function WalletPage() {
  return <SectionPlaceholder section="wallet" />;
}

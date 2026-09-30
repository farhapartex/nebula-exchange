import type { Metadata } from "next";

import { EntryFeeView } from "@/features/onboarding/entry-fee-view";

export const metadata: Metadata = {
  title: "Entry fee",
};

export default function EntryFeePage() {
  return <EntryFeeView />;
}

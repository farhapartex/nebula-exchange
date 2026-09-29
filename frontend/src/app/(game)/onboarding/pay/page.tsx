import type { Metadata } from "next";

import { EntryFeePlaceholder } from "@/features/onboarding/entry-fee-placeholder";

export const metadata: Metadata = {
  title: "Entry fee",
};

export default function EntryFeePage() {
  return <EntryFeePlaceholder />;
}

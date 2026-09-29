import type { Metadata } from "next";

import { SectionPlaceholder } from "@/components/app-shell/section-placeholder";

export const metadata: Metadata = {
  title: "Hangar",
};

export default function HangarPage() {
  return <SectionPlaceholder section="hangar" />;
}

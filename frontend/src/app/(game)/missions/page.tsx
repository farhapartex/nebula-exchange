import type { Metadata } from "next";

import { SectionPlaceholder } from "@/components/app-shell/section-placeholder";

export const metadata: Metadata = {
  title: "Missions",
};

export default function MissionsPage() {
  return <SectionPlaceholder section="missions" />;
}

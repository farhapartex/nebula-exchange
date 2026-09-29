import type { Metadata } from "next";

import { SectionPlaceholder } from "@/components/app-shell/section-placeholder";

export const metadata: Metadata = {
  title: "Settings",
};

export default function SettingsPage() {
  return <SectionPlaceholder section="settings" />;
}

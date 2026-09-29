import type { Metadata } from "next";

import { SettingsPageContent } from "@/features/settings/settings-page-content";

export const metadata: Metadata = {
  title: "Settings",
};

export default function SettingsPage() {
  return <SettingsPageContent />;
}

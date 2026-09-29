import type { Metadata } from "next";

import { SectionPlaceholder } from "@/components/app-shell/section-placeholder";

export const metadata: Metadata = {
  title: "Notifications",
};

export default function NotificationsPage() {
  return <SectionPlaceholder section="notifications" />;
}

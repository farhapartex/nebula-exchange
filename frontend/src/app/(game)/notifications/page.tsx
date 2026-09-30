import type { Metadata } from "next";

import { navigationLinkFor } from "@/components/app-shell/navigation-links";
import { PageContainer } from "@/components/layout/page-container";
import { SectionHeader } from "@/components/layout/section-header";
import { NotificationsFeed } from "@/features/notifications/notifications-feed";

export const metadata: Metadata = {
  title: "Notifications",
};

export default function NotificationsPage() {
  const navigationLink = navigationLinkFor("notifications");
  return (
    <PageContainer className="max-w-3xl py-8 sm:py-10">
      <SectionHeader title={navigationLink.label} description={navigationLink.description} icon={navigationLink.icon} />
      <NotificationsFeed />
    </PageContainer>
  );
}

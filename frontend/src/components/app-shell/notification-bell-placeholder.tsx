import Link from "next/link";
import { Bell } from "lucide-react";

import { placeholderPilot } from "@/components/app-shell/placeholder-pilot";

export function NotificationBellPlaceholder() {
  const unreadCount = placeholderPilot.unreadNotificationCount;

  return (
    <Link
      href="/notifications"
      aria-label={`Notifications, ${unreadCount} unread`}
      className="relative flex size-9 items-center justify-center rounded-lg text-muted transition-colors hover:bg-surface-raised hover:text-foreground focus-visible:ring-2 focus-visible:ring-highlight focus-visible:outline-none"
    >
      <Bell className="size-5" />
      {unreadCount > 0 && (
        <span className="absolute top-1 right-1 flex min-w-4 items-center justify-center rounded-full bg-down px-1 text-[0.625rem] leading-4 font-semibold text-white">
          {unreadCount}
        </span>
      )}
    </Link>
  );
}

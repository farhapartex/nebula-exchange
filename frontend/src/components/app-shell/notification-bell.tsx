"use client";

import { useState } from "react";
import Link from "next/link";
import { Bell, BellOff, CheckCheck } from "lucide-react";
import { Popover } from "radix-ui";

import { EmptyState } from "@/components/ui/empty-state";
import { Skeleton } from "@/components/ui/skeleton";
import { useAuth } from "@/features/auth/session/use-auth";
import { NotificationListItem } from "@/features/notifications/notification-list-item";
import {
  useMarkNotificationsRead,
  useRecentNotifications,
  useUnreadNotificationCount,
} from "@/features/notifications/use-notifications";

export function NotificationBell() {
  const { status } = useAuth();
  const [isOpen, setIsOpen] = useState(false);
  const unreadCountQuery = useUnreadNotificationCount();
  const recentQuery = useRecentNotifications(isOpen);
  const markReadMutation = useMarkNotificationsRead();

  if (status !== "authenticated") {
    return null;
  }

  const unreadCount = unreadCountQuery.data ?? 0;
  const recentNotifications = recentQuery.data ?? [];

  function markNotificationRead(notificationId: string) {
    const openedNotification = recentNotifications.find((notification) => notification.id === notificationId);
    if (openedNotification && !openedNotification.isRead) {
      markReadMutation.mutate({ ids: [notificationId] });
    }
    setIsOpen(false);
  }

  function markAllNotificationsRead() {
    markReadMutation.mutate({ all: true });
  }

  return (
    <Popover.Root open={isOpen} onOpenChange={setIsOpen}>
      <Popover.Trigger
        aria-label={unreadCount > 0 ? `Notifications, ${unreadCount} unread` : "Notifications"}
        className="relative flex size-9 items-center justify-center rounded-lg text-muted transition-colors hover:bg-surface-raised hover:text-foreground focus-visible:ring-2 focus-visible:ring-highlight focus-visible:outline-none data-[state=open]:bg-surface-raised data-[state=open]:text-foreground"
      >
        <Bell className="size-5" />
        {unreadCount > 0 && (
          <span className="absolute top-1 right-1 flex min-w-4 items-center justify-center rounded-full bg-down px-1 text-[0.625rem] leading-4 font-semibold text-white">
            {unreadCount > 9 ? "9+" : unreadCount}
          </span>
        )}
      </Popover.Trigger>
      <Popover.Portal>
        <Popover.Content
          align="end"
          sideOffset={8}
          collisionPadding={16}
          className="z-50 w-[min(24rem,calc(100vw-2rem))] rounded-xl border border-border-strong bg-surface-raised shadow-xl shadow-black/50 data-[state=open]:animate-pop-in"
        >
          <div className="flex items-center justify-between gap-3 border-b border-border px-4 py-3">
            <div>
              <p className="text-sm font-semibold text-foreground">Notifications</p>
              <p className="text-xs text-muted">{unreadCount > 0 ? `${unreadCount} unread` : "All caught up"}</p>
            </div>
            <button
              type="button"
              onClick={markAllNotificationsRead}
              disabled={unreadCount === 0}
              className="flex items-center gap-1.5 rounded-md px-2 py-1 text-xs font-medium text-accent-soft transition-colors hover:bg-border disabled:pointer-events-none disabled:opacity-40"
            >
              <CheckCheck className="size-3.5" />
              Mark all read
            </button>
          </div>

          {recentQuery.isPending ? (
            <div className="space-y-2 p-3">
              <Skeleton className="h-12 rounded-lg" />
              <Skeleton className="h-12 rounded-lg" />
            </div>
          ) : recentNotifications.length > 0 ? (
            <ul className="max-h-[min(26rem,60vh)] space-y-0.5 overflow-y-auto p-1.5">
              {recentNotifications.map((notification) => (
                <NotificationListItem key={notification.id} notification={notification} onOpen={markNotificationRead} />
              ))}
            </ul>
          ) : (
            <EmptyState
              className="m-3 border-none py-8"
              icon={BellOff}
              title="No notifications yet"
              description="Mission results, fills and auction updates show up here."
            />
          )}

          <div className="border-t border-border p-1.5">
            <Link
              href="/notifications"
              onClick={() => setIsOpen(false)}
              className="flex h-9 items-center justify-center rounded-lg text-sm font-medium text-muted transition-colors hover:bg-border hover:text-foreground"
            >
              View all notifications
            </Link>
          </div>
        </Popover.Content>
      </Popover.Portal>
    </Popover.Root>
  );
}

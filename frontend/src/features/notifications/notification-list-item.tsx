import Link from "next/link";

import { notificationAppearanceByKind } from "@/features/notifications/notification-appearance";
import type { GameNotification } from "@/features/notifications/notification-types";
import { cn } from "@/utils/class-names";
import { formatRelativeTime } from "@/utils/time/relative-time";

type NotificationListItemProps = {
  notification: GameNotification;
  onOpen: (notificationId: string) => void;
};

export function NotificationListItem({ notification, onOpen }: NotificationListItemProps) {
  const appearance = notificationAppearanceByKind[notification.kind];
  const KindIcon = appearance.icon;

  return (
    <li>
      <Link
        href={notification.href}
        onClick={() => onOpen(notification.id)}
        className={cn(
          "flex gap-3 rounded-lg px-3 py-2.5 transition-colors hover:bg-border/60 focus-visible:bg-border/60 focus-visible:outline-none",
          !notification.isRead && "bg-accent/5",
        )}
      >
        <span className={cn("flex size-8 shrink-0 items-center justify-center rounded-lg", appearance.iconClassName)}>
          <KindIcon className="size-4" />
        </span>
        <span className="min-w-0 flex-1">
          <span className="flex items-center justify-between gap-2">
            <span
              className={cn("truncate text-sm", notification.isRead ? "text-muted" : "font-medium text-foreground")}
            >
              {notification.title}
            </span>
            <span className="shrink-0 text-[0.6875rem] text-subtle">{formatRelativeTime(notification.createdAt)}</span>
          </span>
          <span className="mt-0.5 line-clamp-2 block text-xs leading-relaxed text-muted">{notification.body}</span>
        </span>
        {!notification.isRead && (
          <span className="mt-1.5 size-2 shrink-0 rounded-full bg-highlight" aria-label="Unread" />
        )}
      </Link>
    </li>
  );
}

import type { GameNotification, NotificationKind } from "@/features/notifications/notification-types";
import { requestData, requestList } from "@/lib/api/api-client";
import type { ListEnvelope, PaginationParameters } from "@/lib/api/api-types";

export type ApiNotification = {
  id: string;
  kind: string;
  title: string;
  body: string;
  link: string;
  created_at: string;
  read_at: string | null;
};

export const notificationsQueryKeys = {
  all: ["me", "notifications"] as const,
  recent: ["me", "notifications", "recent"] as const,
  unreadCount: ["me", "notifications", "unread-count"] as const,
  feed: (isUnreadOnly: boolean) => ["me", "notifications", "feed", isUnreadOnly] as const,
};

export function toGameNotification(apiNotification: ApiNotification): GameNotification {
  return {
    id: apiNotification.id,
    kind: apiNotification.kind as NotificationKind,
    title: apiNotification.title,
    body: apiNotification.body,
    href: apiNotification.link || "/notifications",
    createdAt: apiNotification.created_at,
    isRead: apiNotification.read_at !== null,
  };
}

export async function listNotifications(
  pagination: PaginationParameters,
  isUnreadOnly = false,
): Promise<ListEnvelope<GameNotification>> {
  const listPage = await requestList<ApiNotification>("/me/notifications", pagination, {
    query: { unread_only: isUnreadOnly ? "true" : undefined },
  });
  return { ...listPage, data: listPage.data.map(toGameNotification) };
}

export async function fetchUnreadCount(): Promise<number> {
  return (await requestData<{ unread_count: number }>("/me/notifications/unread-count")).unread_count;
}

export function markNotificationsRead(request: { ids: string[] } | { all: true }): Promise<{ marked: number }> {
  return requestData<{ marked: number }>("/me/notifications/read", { method: "POST", body: request });
}

import { http } from "msw";

import type { ApiNotification } from "@/features/notifications/api/notifications-api";
import { buildApiUrl } from "@/lib/api/api-config";
import { createMockNotifications } from "@/mocks/fixtures/notification-fixtures";
import { mockDataResponse, mockListResponse, simulateLatency } from "@/mocks/utils/mock-responses";

let mockNotifications: ApiNotification[] = createMockNotifications();

export const notificationsHandlers = [
  http.get(buildApiUrl("/me/notifications/unread-count"), async () => {
    await simulateLatency(150);
    return mockDataResponse({ unread_count: mockNotifications.filter((notification) => !notification.read_at).length });
  }),
  http.get(buildApiUrl("/me/notifications"), async ({ request }) => {
    await simulateLatency();
    const requestUrl = new URL(request.url);
    const isUnreadOnly = requestUrl.searchParams.get("unread_only") === "true";
    return mockListResponse(
      isUnreadOnly ? mockNotifications.filter((notification) => !notification.read_at) : mockNotifications,
      requestUrl,
    );
  }),
  http.post(buildApiUrl("/me/notifications/read"), async ({ request }) => {
    await simulateLatency(150);
    const markRequest = (await request.json()) as { ids?: string[]; all?: boolean };
    let markedCount = 0;
    const readAt = new Date().toISOString();
    mockNotifications = mockNotifications.map((notification) => {
      const shouldMark = !notification.read_at && (markRequest.all || markRequest.ids?.includes(notification.id));
      if (!shouldMark) {
        return notification;
      }
      markedCount += 1;
      return { ...notification, read_at: readAt };
    });
    return mockDataResponse({ marked: markedCount });
  }),
];

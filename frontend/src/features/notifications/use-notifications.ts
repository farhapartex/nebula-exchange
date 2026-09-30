"use client";

import { useMutation, useQuery, useQueryClient } from "@tanstack/react-query";

import { useAuth } from "@/features/auth/session/use-auth";
import {
  fetchUnreadCount,
  listNotifications,
  markNotificationsRead,
  notificationsQueryKeys,
} from "@/features/notifications/api/notifications-api";

const notificationRefreshIntervalInMilliseconds = 30_000;
const recentNotificationLimit = 5;

export function useUnreadNotificationCount() {
  const { status } = useAuth();
  return useQuery({
    queryKey: notificationsQueryKeys.unreadCount,
    queryFn: fetchUnreadCount,
    enabled: status === "authenticated",
    refetchInterval: notificationRefreshIntervalInMilliseconds,
  });
}

export function useRecentNotifications(isEnabled: boolean) {
  return useQuery({
    queryKey: notificationsQueryKeys.recent,
    queryFn: async () => (await listNotifications({ limit: recentNotificationLimit })).data,
    enabled: isEnabled,
  });
}

export function useMarkNotificationsRead() {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: markNotificationsRead,
    onSuccess: () => void queryClient.invalidateQueries({ queryKey: notificationsQueryKeys.all }),
  });
}

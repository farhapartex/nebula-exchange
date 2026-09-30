"use client";

import { useState } from "react";
import { useInfiniteQuery } from "@tanstack/react-query";
import { BellOff, CheckCheck } from "lucide-react";

import { Button } from "@/components/ui/button";
import { EmptyState } from "@/components/ui/empty-state";
import { ErrorState } from "@/components/ui/error-state";
import { Skeleton } from "@/components/ui/skeleton";
import { Tabs, TabsList, TabsTrigger } from "@/components/ui/tabs";
import { listNotifications, notificationsQueryKeys } from "@/features/notifications/api/notifications-api";
import { NotificationListItem } from "@/features/notifications/notification-list-item";
import { useMarkNotificationsRead, useUnreadNotificationCount } from "@/features/notifications/use-notifications";

const feedPageSize = 20;

export function NotificationsFeed() {
  const [isUnreadOnly, setIsUnreadOnly] = useState(false);
  const unreadCountQuery = useUnreadNotificationCount();
  const markReadMutation = useMarkNotificationsRead();
  const feedQuery = useInfiniteQuery({
    queryKey: notificationsQueryKeys.feed(isUnreadOnly),
    queryFn: ({ pageParam }) => listNotifications({ cursor: pageParam, limit: feedPageSize }, isUnreadOnly),
    initialPageParam: null as string | null,
    getNextPageParam: (lastPage) => lastPage.pagination.next_cursor,
  });
  const notifications = feedQuery.data?.pages.flatMap((feedPage) => feedPage.data) ?? [];
  const unreadCount = unreadCountQuery.data ?? 0;

  return (
    <div className="space-y-4">
      <div className="flex flex-wrap items-center justify-between gap-3">
        <Tabs
          value={isUnreadOnly ? "unread" : "all"}
          onValueChange={(nextValue) => setIsUnreadOnly(nextValue === "unread")}
        >
          <TabsList>
            <TabsTrigger value="all">All</TabsTrigger>
            <TabsTrigger value="unread">
              Unread{unreadCount > 0 && <span className="ml-1.5 text-xs text-subtle tabular-nums">{unreadCount}</span>}
            </TabsTrigger>
          </TabsList>
        </Tabs>
        <Button
          variant="secondary"
          size="sm"
          disabled={unreadCount === 0}
          isLoading={markReadMutation.isPending}
          onClick={() => markReadMutation.mutate({ all: true })}
        >
          <CheckCheck className="size-4" aria-hidden="true" />
          Mark all read
        </Button>
      </div>

      {feedQuery.isPending && <Skeleton className="h-64 rounded-2xl" />}
      {feedQuery.isError && (
        <ErrorState message="We couldn't load your notifications." onRetry={() => void feedQuery.refetch()} />
      )}
      {feedQuery.isSuccess && notifications.length === 0 && (
        <EmptyState
          icon={BellOff}
          title={isUnreadOnly ? "You're all caught up" : "No notifications yet"}
          description="Mission results, crafts, payments, fills and auction updates show up here."
        />
      )}
      {notifications.length > 0 && (
        <ul className="space-y-0.5 rounded-2xl border border-border bg-surface/80 p-1.5">
          {notifications.map((notification) => (
            <NotificationListItem
              key={notification.id}
              notification={notification}
              onOpen={(notificationID) => !notification.isRead && markReadMutation.mutate({ ids: [notificationID] })}
            />
          ))}
        </ul>
      )}
      {feedQuery.hasNextPage && (
        <Button
          variant="ghost"
          size="sm"
          isLoading={feedQuery.isFetchingNextPage}
          onClick={() => void feedQuery.fetchNextPage()}
        >
          Show older
        </Button>
      )}
    </div>
  );
}

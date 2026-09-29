"use client";

import { useState } from "react";
import { useInfiniteQuery, useMutation, useQueryClient } from "@tanstack/react-query";
import { Laptop, Smartphone } from "lucide-react";

import { ContentSection } from "@/components/layout/content-section";
import { Button } from "@/components/ui/button";
import { ConfirmDialog } from "@/components/ui/confirm-dialog";
import { EmptyState } from "@/components/ui/empty-state";
import { ErrorState } from "@/components/ui/error-state";
import { Skeleton } from "@/components/ui/skeleton";
import { StatusBadge } from "@/components/ui/status-badge";
import { useToast } from "@/components/ui/toast/use-toast";
import { useLogOutAction } from "@/features/auth/session/use-log-out-action";
import { listActiveSessions, revokeSession, type ActiveSession } from "@/features/settings/api/settings-api";
import { describeUserAgent } from "@/utils/device/describe-user-agent";
import { formatRelativeTime } from "@/utils/time/relative-time";

const sessionsPageSize = 10;
const activeSessionsQueryKey = ["active-sessions"];

function isMobileAgent(userAgent: string): boolean {
  return /iPhone|iPad|Android|Mobile/.test(userAgent);
}

export function SessionsSection() {
  const queryClient = useQueryClient();
  const { showToast } = useToast();
  const { logOutAndLeave } = useLogOutAction();
  const [sessionPendingRevocation, setSessionPendingRevocation] = useState<ActiveSession | null>(null);

  const sessionsQuery = useInfiniteQuery({
    queryKey: activeSessionsQueryKey,
    queryFn: ({ pageParam }) => listActiveSessions({ cursor: pageParam, limit: sessionsPageSize }),
    initialPageParam: null as string | null,
    getNextPageParam: (lastPage) => lastPage.pagination.next_cursor,
  });

  const revokeMutation = useMutation({
    mutationFn: (activeSession: ActiveSession) => revokeSession(activeSession.id),
    onSuccess: () => {
      void queryClient.invalidateQueries({ queryKey: activeSessionsQueryKey });
      showToast({ tone: "success", title: "Session signed out" });
    },
  });

  const activeSessions = sessionsQuery.data?.pages.flatMap((sessionPage) => sessionPage.data) ?? [];

  return (
    <ContentSection title="Active sessions" description="Devices that are signed in to your account right now.">
      {sessionsQuery.isPending && (
        <div className="space-y-2">
          <Skeleton className="h-16 w-full rounded-xl" />
          <Skeleton className="h-16 w-full rounded-xl" />
        </div>
      )}
      {sessionsQuery.isError && (
        <ErrorState message="We couldn't load your sessions." onRetry={() => void sessionsQuery.refetch()} />
      )}
      {sessionsQuery.isSuccess && activeSessions.length === 0 && <EmptyState title="No active sessions" />}

      <ul className="space-y-2">
        {activeSessions.map((activeSession) => {
          const DeviceIcon = isMobileAgent(activeSession.user_agent) ? Smartphone : Laptop;
          return (
            <li
              key={activeSession.id}
              className="flex flex-wrap items-center gap-3 rounded-xl border border-border bg-background/50 p-3"
            >
              <span className="flex size-10 shrink-0 items-center justify-center rounded-lg bg-surface-raised text-muted">
                <DeviceIcon className="size-5" />
              </span>
              <div className="min-w-0 flex-1">
                <p className="flex flex-wrap items-center gap-2 text-sm font-medium text-foreground">
                  {describeUserAgent(activeSession.user_agent)}
                  {activeSession.is_current && <StatusBadge label="This device" tone="accent" />}
                </p>
                <p className="mt-0.5 text-xs text-muted">
                  {activeSession.ip_address || "Unknown IP"} · Active {formatRelativeTime(activeSession.last_active_at)}{" "}
                  · Signed in {formatRelativeTime(activeSession.started_at)}
                </p>
              </div>
              {activeSession.is_current ? (
                <Button variant="ghost" size="sm" onClick={() => void logOutAndLeave()}>
                  Log out
                </Button>
              ) : (
                <Button variant="secondary" size="sm" onClick={() => setSessionPendingRevocation(activeSession)}>
                  Sign out
                </Button>
              )}
            </li>
          );
        })}
      </ul>

      {sessionsQuery.hasNextPage && (
        <Button
          variant="ghost"
          size="sm"
          className="mt-3"
          isLoading={sessionsQuery.isFetchingNextPage}
          onClick={() => void sessionsQuery.fetchNextPage()}
        >
          Show more sessions
        </Button>
      )}

      <ConfirmDialog
        isOpen={sessionPendingRevocation !== null}
        onOpenChange={(isOpen) => !isOpen && setSessionPendingRevocation(null)}
        title="Sign out this device?"
        description={
          sessionPendingRevocation
            ? `${describeUserAgent(sessionPendingRevocation.user_agent)} will need to log in again.`
            : undefined
        }
        confirmLabel="Sign out device"
        tone="danger"
        onConfirm={async () => {
          if (sessionPendingRevocation) {
            await revokeMutation.mutateAsync(sessionPendingRevocation);
          }
        }}
      />
    </ContentSection>
  );
}

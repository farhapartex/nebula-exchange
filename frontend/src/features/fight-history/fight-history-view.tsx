"use client";

import Link from "next/link";
import { useInfiniteQuery } from "@tanstack/react-query";
import { ArrowLeft, Trophy } from "lucide-react";

import { Button } from "@/components/ui/button";
import { ErrorState } from "@/components/ui/error-state";
import { Skeleton } from "@/components/ui/skeleton";
import { useCurrentPlayer } from "@/features/auth/session/use-current-player";
import { HubPanel } from "@/features/fight-hub/hub-panel";
import { fetchWonFightsPage, wonFightsQueryKey } from "@/features/fight-history/api/fight-history-api";
import { ChapterWinsSection } from "@/features/fight-history/chapter-wins-section";
import { groupWinsByChapter } from "@/features/fight-history/group-wins-by-chapter";

export function FightHistoryView() {
  const currentPlayer = useCurrentPlayer().data;
  const wonFightsQuery = useInfiniteQuery({
    queryKey: wonFightsQueryKey,
    queryFn: ({ pageParam }) => fetchWonFightsPage(pageParam),
    initialPageParam: null as string | null,
    getNextPageParam: (lastPage) => lastPage.pagination.next_cursor,
  });
  const wonFights = wonFightsQuery.data?.pages.flatMap((wonFightsPage) => wonFightsPage.data) ?? [];

  return (
    <div className="space-y-6">
      <Link
        href="/fight"
        className="inline-flex items-center gap-1.5 text-sm text-muted transition-colors hover:text-foreground"
      >
        <ArrowLeft className="size-4" aria-hidden="true" />
        Back to fight
      </Link>
      <div className="flex flex-wrap items-end justify-between gap-4">
        <div>
          <h1 className="font-display text-4xl tracking-[0.06em] text-foreground">Fight history</h1>
          <p className="mt-1 text-sm text-muted">Every fight you have won, chapter by chapter.</p>
        </div>
        {currentPlayer && (
          <span className="flex items-center gap-2 rounded-lg border border-border-strong bg-surface-raised px-3 py-2 text-sm">
            <Trophy className="size-4 text-amber-400" aria-hidden="true" />
            <span className="font-mono text-foreground tabular-nums">{currentPlayer.total_win}</span>
            <span className="text-xs text-subtle">wins in total</span>
          </span>
        )}
      </div>

      {wonFightsQuery.isError ? (
        <ErrorState
          message="Your fight history could not be loaded. Try again."
          onRetry={() => void wonFightsQuery.refetch()}
          isRetrying={wonFightsQuery.isFetching}
        />
      ) : !wonFightsQuery.data ? (
        <div className="space-y-4">
          <Skeleton className="h-56 rounded-2xl" />
          <Skeleton className="h-40 rounded-2xl" />
        </div>
      ) : wonFights.length === 0 ? (
        <HubPanel className="text-center">
          <p className="font-medium text-foreground">No wins yet</p>
          <p className="mt-1 text-sm text-muted">Win your first fight and it will show up here.</p>
          <Button asChild size="sm" className="mt-4">
            <Link href="/fight">Go to fight</Link>
          </Button>
        </HubPanel>
      ) : (
        <div className="space-y-4">
          {groupWinsByChapter(wonFights).map((chapterWins) => (
            <ChapterWinsSection key={chapterWins.chapterNumber} chapterWins={chapterWins} />
          ))}
          {wonFightsQuery.hasNextPage && (
            <div className="flex justify-center">
              <Button
                variant="secondary"
                onClick={() => void wonFightsQuery.fetchNextPage()}
                isLoading={wonFightsQuery.isFetchingNextPage}
              >
                Show more
              </Button>
            </div>
          )}
        </div>
      )}
    </div>
  );
}

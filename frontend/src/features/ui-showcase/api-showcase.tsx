"use client";

import { useInfiniteQuery, useMutation, useQuery } from "@tanstack/react-query";

import { ContentSection } from "@/components/layout/content-section";
import { NcAmount } from "@/components/money/nc-amount";
import { Button } from "@/components/ui/button";
import { ErrorState } from "@/components/ui/error-state";
import { Skeleton } from "@/components/ui/skeleton";
import { StatusBadge } from "@/components/ui/status-badge";
import { requestData, requestList } from "@/lib/api/api-client";
import { areApiMocksEnabled } from "@/lib/api/api-config";
import { isApiError } from "@/lib/api/api-error";
import type { SampleTrade } from "@/mocks/handlers/dev-showcase-handlers";

type HealthStatus = { status: string; checked_at: string };

const sampleTradesPageSize = 5;

export function ApiShowcase() {
  const healthQuery = useQuery({
    queryKey: ["health"],
    queryFn: ({ signal }) => requestData<HealthStatus>("/health", { signal }),
  });

  const sampleTradesQuery = useInfiniteQuery({
    queryKey: ["dev", "sample-trades"],
    queryFn: ({ pageParam, signal }) =>
      requestList<SampleTrade>("/dev/sample-trades", { cursor: pageParam, limit: sampleTradesPageSize }, { signal }),
    initialPageParam: null as string | null,
    getNextPageParam: (lastPage) => lastPage.pagination.next_cursor,
  });

  const failingOrderMutation = useMutation({
    mutationFn: () => requestData("/dev/sample-orders", { method: "POST", body: { symbol: "IRON/NC", quantity: 40 } }),
  });

  const loadedTrades = sampleTradesQuery.data?.pages.flatMap((tradePage) => tradePage.data) ?? [];
  const mutationError = isApiError(failingOrderMutation.error) ? failingOrderMutation.error : null;

  return (
    <ContentSection
      title="API client and mocks"
      description="Requests go through the typed client. With mocks on, MSW answers them in the browser."
    >
      <div className="space-y-6">
        <div className="flex flex-wrap items-center gap-3 text-sm">
          <StatusBadge
            label={areApiMocksEnabled ? "Mocks on" : "Mocks off"}
            tone={areApiMocksEnabled ? "accent" : "neutral"}
          />
          <span className="text-muted">GET /health</span>
          {healthQuery.isPending && <Skeleton className="h-5 w-24" />}
          {healthQuery.isSuccess && <StatusBadge label={healthQuery.data.status} tone="success" />}
          {healthQuery.isError && <StatusBadge label="unreachable" tone="danger" />}
        </div>

        <div>
          <p className="mb-2 text-sm text-muted">Cursor pagination: GET /dev/sample-trades</p>
          <ul className="divide-y divide-border overflow-hidden rounded-xl border border-border">
            {loadedTrades.map((sampleTrade) => (
              <li key={sampleTrade.id} className="flex items-center justify-between gap-4 px-4 py-2 text-sm">
                <span className="font-mono text-foreground">{sampleTrade.symbol}</span>
                <span className={sampleTrade.side === "buy" ? "text-up" : "text-down"}>
                  {sampleTrade.side === "buy" ? "▲ Buy" : "▼ Sell"}
                </span>
                <span className="font-mono text-muted tabular-nums">× {sampleTrade.quantity}</span>
                <NcAmount amount={sampleTrade.price} fractionDigits={4} />
              </li>
            ))}
            {sampleTradesQuery.isPending &&
              Array.from({ length: sampleTradesPageSize }, (_, skeletonIndex) => (
                <li key={skeletonIndex} className="px-4 py-2.5">
                  <Skeleton className="h-4 w-full" />
                </li>
              ))}
          </ul>
          <div className="mt-3 flex items-center gap-3">
            <Button
              variant="secondary"
              size="sm"
              isLoading={sampleTradesQuery.isFetchingNextPage}
              disabled={!sampleTradesQuery.hasNextPage}
              onClick={() => sampleTradesQuery.fetchNextPage()}
            >
              {sampleTradesQuery.hasNextPage ? "Load more" : "No more trades"}
            </Button>
            <span className="text-xs text-subtle">{loadedTrades.length} loaded</span>
          </div>
        </div>

        <div>
          <p className="mb-2 text-sm text-muted">Error envelope: POST /dev/sample-orders always fails</p>
          <Button
            size="sm"
            variant="buy"
            isLoading={failingOrderMutation.isPending}
            onClick={() => failingOrderMutation.mutate()}
          >
            Place sample order
          </Button>
          {mutationError && (
            <ErrorState
              className="mt-3"
              title={mutationError.code}
              message={`${mutationError.statusCode} · ${mutationError.message}`}
            />
          )}
          {healthQuery.isError && !areApiMocksEnabled && (
            <p className="mt-3 text-xs text-subtle">Start the backend or set NEXT_PUBLIC_API_MOCKS=true.</p>
          )}
        </div>
      </div>
    </ContentSection>
  );
}

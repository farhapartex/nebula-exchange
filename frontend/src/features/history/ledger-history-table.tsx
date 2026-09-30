"use client";

import { useState } from "react";
import { useInfiniteQuery } from "@tanstack/react-query";
import { ScrollText } from "lucide-react";

import { DataTable, type DataTableColumn } from "@/components/ui/data-table";
import { EmptyState } from "@/components/ui/empty-state";
import { ErrorState } from "@/components/ui/error-state";
import { Select } from "@/components/ui/select";
import { Skeleton } from "@/components/ui/skeleton";
import { Tooltip } from "@/components/ui/tooltip";
import type { CatalogItem } from "@/features/catalog/api/catalog-types";
import { useCatalogItems } from "@/features/catalog/use-catalog";
import {
  ledgerHistoryQueryKey,
  listLedgerJournals,
  type JournalType,
  type LedgerJournal,
} from "@/features/history/api/ledger-history-api";
import { journalTypeFilterOptions, journalTypeLabels } from "@/features/history/journal-type-labels";
import { LedgerEntryChange } from "@/features/history/ledger-entry-change";
import { formatRelativeTime } from "@/utils/time/relative-time";

const ledgerPageSize = 25;

function buildLedgerColumns(itemsByID: Map<number, CatalogItem>): DataTableColumn<LedgerJournal>[] {
  return [
    {
      key: "time",
      header: "Time",
      sortValue: (journal) => journal.created_at,
      className: "w-40 whitespace-nowrap",
      render: (journal) => (
        <Tooltip content={new Date(journal.created_at).toLocaleString()}>
          <span tabIndex={0} className="text-muted">
            {formatRelativeTime(journal.created_at)}
          </span>
        </Tooltip>
      ),
    },
    {
      key: "type",
      header: "Type",
      sortValue: (journal) => journalTypeLabels[journal.type],
      className: "w-44",
      render: (journal) => <span className="font-medium text-foreground">{journalTypeLabels[journal.type]}</span>,
    },
    {
      key: "changes",
      header: "Changes",
      render: (journal) => (
        <ul className="space-y-1">
          {journal.entries.map((entry) => (
            <li key={`${entry.item_id ?? "nc"}-${entry.bucket ?? "item"}`}>
              <LedgerEntryChange entry={entry} itemsByID={itemsByID} />
            </li>
          ))}
        </ul>
      ),
    },
    {
      key: "reference",
      header: "Reference",
      alignment: "end",
      className: "w-44",
      render: (journal) => (
        <span className="font-mono text-xs whitespace-nowrap text-subtle" title={journal.reference.id}>
          {journal.reference.type} · {journal.reference.id.slice(0, 8)}
        </span>
      ),
    },
  ];
}

export function LedgerHistoryTable() {
  const [journalTypeFilter, setJournalTypeFilter] = useState<JournalType | null>(null);
  const catalogQuery = useCatalogItems();
  const journalsQuery = useInfiniteQuery({
    queryKey: ledgerHistoryQueryKey(journalTypeFilter),
    queryFn: ({ pageParam }) => listLedgerJournals(journalTypeFilter, { cursor: pageParam, limit: ledgerPageSize }),
    initialPageParam: null as string | null,
    getNextPageParam: (lastPage) => lastPage.pagination.next_cursor,
  });
  const journals = journalsQuery.data?.pages.flatMap((journalPage) => journalPage.data) ?? [];

  return (
    <div className="space-y-4">
      <Select
        className="max-w-60"
        label="Show"
        options={journalTypeFilterOptions}
        value={journalTypeFilter ?? "all"}
        onValueChange={(selectedValue) =>
          setJournalTypeFilter(selectedValue === "all" ? null : (selectedValue as JournalType))
        }
      />
      {journalsQuery.isPending && <Skeleton className="h-72 rounded-2xl" />}
      {journalsQuery.isError && (
        <ErrorState message="We couldn't load your ledger." onRetry={() => void journalsQuery.refetch()} />
      )}
      {journalsQuery.isSuccess && journals.length === 0 && (
        <EmptyState
          icon={ScrollText}
          title="Nothing here yet"
          description="Every NC and item movement on your account will appear in this list."
        />
      )}
      {journals.length > 0 && (
        <DataTable
          caption="Ledger movements"
          columns={buildLedgerColumns(catalogQuery.itemsByID)}
          rows={journals}
          rowKey={(journal) => journal.id}
          hasMoreRows={journalsQuery.hasNextPage}
          isLoadingMoreRows={journalsQuery.isFetchingNextPage}
          onLoadMoreRows={() => void journalsQuery.fetchNextPage()}
        />
      )}
    </div>
  );
}

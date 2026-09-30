"use client";

import { useState, type ReactNode } from "react";
import { ArrowDown, ArrowUp, ChevronsUpDown } from "lucide-react";

import { Button } from "@/components/ui/button";
import { cn } from "@/utils/class-names";
import { sortRows, type SortableValue, type SortDirection } from "@/utils/collections/sort-rows";

export type DataTableColumn<Row> = {
  key: string;
  header: string;
  render: (row: Row) => ReactNode;
  sortValue?: (row: Row) => SortableValue;
  alignment?: "start" | "end";
  className?: string;
};

type DataTableSort = { columnKey: string; direction: SortDirection } | null;

type DataTableProps<Row> = {
  columns: DataTableColumn<Row>[];
  rows: Row[];
  rowKey: (row: Row) => string;
  caption: string;
  initialSort?: DataTableSort;
  hasMoreRows?: boolean;
  isLoadingMoreRows?: boolean;
  onLoadMoreRows?: () => void;
};

export function DataTable<Row>({
  columns,
  rows,
  rowKey,
  caption,
  initialSort = null,
  hasMoreRows = false,
  isLoadingMoreRows = false,
  onLoadMoreRows,
}: DataTableProps<Row>) {
  const [activeSort, setActiveSort] = useState<DataTableSort>(initialSort);
  const sortedColumn = columns.find((column) => column.key === activeSort?.columnKey);
  const visibleRows =
    activeSort && sortedColumn?.sortValue ? sortRows(rows, sortedColumn.sortValue, activeSort.direction) : rows;

  function toggleSort(columnKey: string) {
    setActiveSort((currentSort) =>
      currentSort?.columnKey === columnKey
        ? { columnKey, direction: currentSort.direction === "descending" ? "ascending" : "descending" }
        : { columnKey, direction: "descending" },
    );
  }

  return (
    <div className="overflow-hidden rounded-2xl border border-border bg-surface/80">
      <div className="overflow-x-auto">
        <table className="w-full min-w-[40rem] border-collapse text-sm">
          <caption className="sr-only">{caption}</caption>
          <thead>
            <tr className="border-b border-border text-xs text-muted">
              {columns.map((column) => {
                const isSorted = activeSort?.columnKey === column.key;
                const SortIcon = !isSorted
                  ? ChevronsUpDown
                  : activeSort.direction === "ascending"
                    ? ArrowUp
                    : ArrowDown;
                return (
                  <th
                    key={column.key}
                    scope="col"
                    aria-sort={isSorted ? activeSort.direction : undefined}
                    className={cn(
                      "px-4 py-3 font-medium",
                      column.alignment === "end" ? "text-right" : "text-left",
                      column.className,
                    )}
                  >
                    {column.sortValue ? (
                      <button
                        type="button"
                        onClick={() => toggleSort(column.key)}
                        className={cn(
                          "inline-flex items-center gap-1 rounded hover:text-foreground focus-visible:ring-2 focus-visible:ring-highlight focus-visible:outline-none",
                          isSorted && "text-foreground",
                        )}
                      >
                        {column.header}
                        <SortIcon className="size-3.5" aria-hidden="true" />
                      </button>
                    ) : (
                      column.header
                    )}
                  </th>
                );
              })}
            </tr>
          </thead>
          <tbody>
            {visibleRows.map((row) => (
              <tr key={rowKey(row)} className="border-b border-border/60 last:border-b-0 hover:bg-surface-raised/40">
                {columns.map((column) => (
                  <td
                    key={column.key}
                    className={cn(
                      "px-4 py-3 align-top",
                      column.alignment === "end" ? "text-right" : "text-left",
                      column.className,
                    )}
                  >
                    {column.render(row)}
                  </td>
                ))}
              </tr>
            ))}
          </tbody>
        </table>
      </div>
      {hasMoreRows && onLoadMoreRows && (
        <div className="border-t border-border p-2 text-center">
          <Button variant="ghost" size="sm" isLoading={isLoadingMoreRows} onClick={onLoadMoreRows}>
            Load more
          </Button>
        </div>
      )}
    </div>
  );
}

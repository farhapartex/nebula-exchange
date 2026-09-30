"use client";

import Link from "next/link";
import { useInfiniteQuery } from "@tanstack/react-query";
import { Receipt } from "lucide-react";

import { NcAmount } from "@/components/money/nc-amount";
import { DataTable, type DataTableColumn } from "@/components/ui/data-table";
import { EmptyState } from "@/components/ui/empty-state";
import { ErrorState } from "@/components/ui/error-state";
import { Skeleton } from "@/components/ui/skeleton";
import { StatusBadge } from "@/components/ui/status-badge";
import { listPayments, type Payment } from "@/features/payments/api/payments-api";
import {
  paymentMethodLabels,
  paymentPurposeLabels,
  paymentStatusPresentation,
} from "@/features/payments/payment-labels";
import { formatRelativeTime } from "@/utils/time/relative-time";

const paymentsPageSize = 25;

const paymentColumns: DataTableColumn<Payment>[] = [
  {
    key: "time",
    header: "Time",
    sortValue: (payment) => payment.created_at,
    className: "w-36 whitespace-nowrap",
    render: (payment) => (
      <span className="text-muted" title={new Date(payment.created_at).toLocaleString()}>
        {formatRelativeTime(payment.created_at)}
      </span>
    ),
  },
  {
    key: "purpose",
    header: "Purpose",
    sortValue: (payment) => paymentPurposeLabels[payment.purpose],
    render: (payment) => (
      <span className="text-foreground">
        {paymentPurposeLabels[payment.purpose]}
        {payment.sku && <span className="ml-1.5 text-xs text-subtle">{payment.sku}</span>}
      </span>
    ),
  },
  {
    key: "method",
    header: "Method",
    render: (payment) => <span className="text-muted">{paymentMethodLabels[payment.method]}</span>,
  },
  {
    key: "status",
    header: "Status",
    sortValue: (payment) => payment.status,
    render: (payment) => {
      const presentation = paymentStatusPresentation[payment.status];
      return (
        <span className="inline-flex flex-wrap items-center gap-1.5">
          <StatusBadge label={presentation.label} tone={presentation.tone} />
          {payment.purpose_status === "FAILED" && <StatusBadge label="NC kept" tone="warning" />}
        </span>
      );
    },
  },
  {
    key: "amount",
    header: "Amount",
    alignment: "end",
    sortValue: (payment) => BigInt(payment.amount),
    render: (payment) => <NcAmount amount={payment.amount} className="text-foreground" />,
  },
  {
    key: "reference",
    header: "Reference",
    alignment: "end",
    render: (payment) => (
      <Link
        href={`/payment/result?id=${payment.id}`}
        className="font-mono text-xs whitespace-nowrap text-subtle hover:text-foreground"
      >
        {payment.id.slice(0, 8)}
      </Link>
    ),
  },
];

export function PaymentsHistoryTable() {
  const paymentsQuery = useInfiniteQuery({
    queryKey: ["payments", "list"],
    queryFn: ({ pageParam }) => listPayments({ cursor: pageParam, limit: paymentsPageSize }),
    initialPageParam: null as string | null,
    getNextPageParam: (lastPage) => lastPage.pagination.next_cursor,
  });
  const paymentList = paymentsQuery.data?.pages.flatMap((paymentPage) => paymentPage.data) ?? [];

  if (paymentsQuery.isPending) {
    return <Skeleton className="h-72 rounded-2xl" />;
  }
  if (paymentsQuery.isError) {
    return <ErrorState message="We couldn't load your payments." onRetry={() => void paymentsQuery.refetch()} />;
  }
  if (paymentList.length === 0) {
    return (
      <EmptyState
        icon={Receipt}
        title="No payments yet"
        description="Card and USDC payments you make will appear here."
      />
    );
  }
  return (
    <DataTable
      caption="Payments"
      columns={paymentColumns}
      rows={paymentList}
      rowKey={(payment) => payment.id}
      hasMoreRows={paymentsQuery.hasNextPage}
      isLoadingMoreRows={paymentsQuery.isFetchingNextPage}
      onLoadMoreRows={() => void paymentsQuery.fetchNextPage()}
    />
  );
}

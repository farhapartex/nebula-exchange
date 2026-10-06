import { PackageOpen } from "lucide-react";

import { Skeleton } from "@/components/ui/skeleton";

export function MarketGridSkeleton() {
  return (
    <div className="grid gap-4 sm:grid-cols-2 lg:grid-cols-3 xl:grid-cols-4">
      {Array.from({ length: 4 }, (_, skeletonIndex) => (
        <Skeleton key={skeletonIndex} className="h-96 rounded-2xl" />
      ))}
    </div>
  );
}

export function MarketEmptyState({ message }: { message: string }) {
  return (
    <div className="flex flex-col items-center gap-2 rounded-2xl border border-dashed border-border-strong px-6 py-14 text-center">
      <PackageOpen className="size-8 text-subtle" aria-hidden="true" />
      <p className="text-sm text-muted">{message}</p>
    </div>
  );
}

import Link from "next/link";

import type { CatalogItem } from "@/features/catalog/api/catalog-types";
import { ItemIcon } from "@/features/catalog/components/item-icon";

type ItemChipProps = {
  item: CatalogItem;
  quantityLabel?: string;
};

export function ItemChip({ item, quantityLabel }: ItemChipProps) {
  return (
    <Link
      href={`/items/${item.id}`}
      className="inline-flex items-center gap-2 rounded-lg border border-border bg-background/50 py-1 pr-2.5 pl-1 text-sm text-foreground transition-colors hover:border-border-strong focus-visible:ring-2 focus-visible:ring-highlight focus-visible:outline-none"
    >
      <ItemIcon item={item} size="sm" />
      <span>{item.name}</span>
      {quantityLabel && <span className="font-mono text-xs text-muted tabular-nums">×{quantityLabel}</span>}
    </Link>
  );
}

import type { ReactNode } from "react";

import { StatusBadge } from "@/components/ui/status-badge";
import type { CatalogItem } from "@/features/catalog/api/catalog-types";
import { ItemAttributeList } from "@/features/catalog/components/item-attributes";
import { ItemIcon } from "@/features/catalog/components/item-icon";
import { describeItemAppearance, describeItemTier } from "@/features/catalog/item-appearance";
import { cn } from "@/utils/class-names";

type ItemCardProps = {
  item: CatalogItem;
  footer?: ReactNode;
  className?: string;
};

export function ItemCard({ item, footer, className }: ItemCardProps) {
  const appearance = describeItemAppearance(item);
  const tierLabel = describeItemTier(item);

  return (
    <article
      className={cn(
        "flex flex-col gap-3 rounded-2xl border border-border bg-surface/80 p-4 transition-colors hover:border-border-strong",
        item.category === "legendary" && "border-amber-500/40",
        className,
      )}
    >
      <div className="flex items-start gap-3">
        <ItemIcon item={item} />
        <div className="min-w-0 flex-1">
          <h3 className="truncate text-sm font-semibold text-foreground">{item.name}</h3>
          <p className="mt-0.5 text-xs text-muted">
            {appearance.label}
            {tierLabel && ` · ${tierLabel}`}
          </p>
        </div>
        {item.is_auction_only && <StatusBadge label="Auction" tone="warning" />}
      </div>
      <p className="line-clamp-2 text-xs leading-relaxed text-muted">{item.description}</p>
      <ItemAttributeList item={item} />
      {footer && <div className="mt-auto border-t border-border pt-3">{footer}</div>}
    </article>
  );
}

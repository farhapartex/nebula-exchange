import { NcAmount } from "@/components/money/nc-amount";
import { bucketDescriptions } from "@/features/balances/bucket-descriptions";
import type { CatalogItem } from "@/features/catalog/api/catalog-types";
import { ItemIcon } from "@/features/catalog/components/item-icon";
import type { LedgerEntry } from "@/features/history/api/ledger-history-api";
import { cn } from "@/utils/class-names";

type LedgerEntryChangeProps = {
  entry: LedgerEntry;
  itemsByID: Map<number, CatalogItem>;
};

export function LedgerEntryChange({ entry, itemsByID }: LedgerEntryChangeProps) {
  const isIncoming = BigInt(entry.amount) > 0n;
  const directionClassName = isIncoming ? "text-up" : "text-down";

  if (entry.item_id === null) {
    return (
      <span className="inline-flex items-center gap-2">
        <NcAmount amount={entry.amount} showPlusSign className={directionClassName} />
        {entry.bucket && <span className="text-xs text-subtle">{bucketDescriptions[entry.bucket].label}</span>}
      </span>
    );
  }

  const entryItem = itemsByID.get(entry.item_id);
  return (
    <span className="inline-flex items-center gap-2">
      <span className={cn("font-mono tabular-nums", directionClassName)}>
        {isIncoming ? "+" : ""}
        {entry.amount}
      </span>
      {entryItem ? (
        <>
          <ItemIcon item={entryItem} size="sm" className="size-6 rounded-md" />
          <span className="text-foreground">{entryItem.name}</span>
        </>
      ) : (
        <span className="text-muted">Item {entry.item_id}</span>
      )}
    </span>
  );
}

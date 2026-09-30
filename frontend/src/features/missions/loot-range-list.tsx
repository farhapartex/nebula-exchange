import type { CatalogItem } from "@/features/catalog/api/catalog-types";
import { ItemIcon } from "@/features/catalog/components/item-icon";
import { formatLootChance, formatQuantityRange } from "@/features/catalog/loot-chance";
import type { LootRange } from "@/features/missions/mission-rules";

type LootRangeListProps = {
  ranges: LootRange[];
  itemsByID: Map<number, CatalogItem>;
};

export function LootRangeList({ ranges, itemsByID }: LootRangeListProps) {
  return (
    <ul className="space-y-1.5">
      {ranges.map((range) => {
        const lootItem = itemsByID.get(range.itemID);
        return (
          <li key={range.itemID} className="flex items-center gap-2 text-sm">
            {lootItem && <ItemIcon item={lootItem} size="sm" className="size-6 rounded-md" />}
            <span className="text-foreground">{lootItem?.name ?? `Item ${range.itemID}`}</span>
            <span className="ml-auto font-mono text-muted tabular-nums">
              {formatQuantityRange(range.minimum, range.maximum)}
            </span>
            {range.chanceBasisPoints < 10_000 && (
              <span className="text-xs text-warning">{formatLootChance(range.chanceBasisPoints)}</span>
            )}
          </li>
        );
      })}
    </ul>
  );
}

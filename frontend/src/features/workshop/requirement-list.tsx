import type { CatalogItem, ItemQuantity } from "@/features/catalog/api/catalog-types";
import { ItemIcon } from "@/features/catalog/components/item-icon";
import { cn } from "@/utils/class-names";

type RequirementListProps = {
  requirements: ItemQuantity[];
  ownedByItem: Map<number, number>;
  itemsByID: Map<number, CatalogItem>;
  multiplier?: number;
};

export function RequirementList({ requirements, ownedByItem, itemsByID, multiplier = 1 }: RequirementListProps) {
  return (
    <ul className="space-y-1.5">
      {requirements.map((requirement) => {
        const requiredItem = itemsByID.get(requirement.item_id);
        const needed = requirement.quantity * multiplier;
        const owned = ownedByItem.get(requirement.item_id) ?? 0;
        return (
          <li key={requirement.item_id} className="flex items-center gap-2 text-sm">
            {requiredItem && <ItemIcon item={requiredItem} size="sm" className="size-6 rounded-md" />}
            <span className="text-foreground">{requiredItem?.name ?? `Item ${requirement.item_id}`}</span>
            <span className={cn("ml-auto font-mono tabular-nums", owned >= needed ? "text-up" : "text-down")}>
              {owned}/{needed}
            </span>
          </li>
        );
      })}
    </ul>
  );
}

import type { CatalogItem, ItemQuantity } from "@/features/catalog/api/catalog-types";
import { ItemIcon } from "@/features/catalog/components/item-icon";

type ShopItemContentsProps = {
  contents: ItemQuantity[];
  itemsByID: Map<number, CatalogItem>;
  multiplier?: number;
};

export function ShopItemContents({ contents, itemsByID, multiplier = 1 }: ShopItemContentsProps) {
  return (
    <ul className="space-y-1.5">
      {contents.map((content) => {
        const contentItem = itemsByID.get(content.item_id);
        return (
          <li key={content.item_id} className="flex items-center gap-2 text-sm">
            {contentItem && <ItemIcon item={contentItem} size="sm" className="size-7 rounded-md" />}
            <span className="font-mono text-accent-soft tabular-nums">{content.quantity * multiplier}×</span>
            <span className="text-foreground">{contentItem?.name ?? `Item ${content.item_id}`}</span>
          </li>
        );
      })}
    </ul>
  );
}

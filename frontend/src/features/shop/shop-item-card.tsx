import { NcAmount } from "@/components/money/nc-amount";
import { Button } from "@/components/ui/button";
import type { CatalogItem, ShopItem } from "@/features/catalog/api/catalog-types";
import { ItemIcon } from "@/features/catalog/components/item-icon";
import { ShopItemContents } from "@/features/shop/shop-item-contents";

type ShopItemCardProps = {
  shopItem: ShopItem;
  itemsByID: Map<number, CatalogItem>;
  onBuy: (shopItem: ShopItem) => void;
};

export function ShopItemCard({ shopItem, itemsByID, onBuy }: ShopItemCardProps) {
  const isBundle = shopItem.contents.length > 1 || (shopItem.contents[0]?.quantity ?? 1) > 1;
  const leadItem = itemsByID.get(shopItem.contents[0]?.item_id ?? 0);

  return (
    <article className="flex flex-col gap-4 rounded-2xl border border-border bg-surface/80 p-5 transition-colors hover:border-border-strong">
      <div className="flex items-start gap-3">
        {leadItem && <ItemIcon item={leadItem} size="md" />}
        <div className="min-w-0 flex-1">
          <h3 className="text-base font-semibold text-foreground">{shopItem.name}</h3>
          <p className="mt-0.5 text-sm text-muted">{shopItem.description}</p>
        </div>
      </div>
      {isBundle && (
        <div className="rounded-xl border border-border bg-background/50 p-3">
          <p className="mb-2 text-xs font-medium tracking-wide text-subtle uppercase">Includes</p>
          <ShopItemContents contents={shopItem.contents} itemsByID={itemsByID} />
        </div>
      )}
      <div className="mt-auto flex items-center justify-between gap-3 border-t border-border pt-4">
        <NcAmount amount={shopItem.price} className="text-lg font-semibold text-foreground" />
        <Button onClick={() => onBuy(shopItem)}>Buy</Button>
      </div>
    </article>
  );
}

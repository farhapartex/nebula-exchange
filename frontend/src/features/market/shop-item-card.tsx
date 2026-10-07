import { Lock, ShieldAlert } from "lucide-react";

import { Button } from "@/components/ui/button";
import type { ShopItem } from "@/features/market/api/market-types";
import { CoinPrice } from "@/features/market/coin-price";
import { RarityBadge } from "@/features/market/rarity-badge";
import { ToolArtwork } from "@/features/market/tool-artwork";
import { categoryLabels } from "@/features/market/tool-rarity";
import { ToolStatList } from "@/features/market/tool-stat-list";
import { cn } from "@/utils/class-names";

type ShopItemCardProps = {
  shopItem: ShopItem;
  onBuy: (shopItem: ShopItem) => void;
};

export function ShopItemCard({ shopItem, onBuy }: ShopItemCardProps) {
  const toolType = shopItem.tool_type;
  const unlockLevel = shopItem.unlock_level;
  return (
    <article
      className={cn(
        "flex flex-col gap-3 rounded-2xl border border-border bg-surface/70 p-4",
        !shopItem.is_unlocked && "opacity-75",
      )}
    >
      <ToolArtwork toolType={toolType} isLocked={!shopItem.is_unlocked} />
      <div className="flex items-start justify-between gap-2">
        <div>
          <h3 className="font-display text-xl leading-tight tracking-[0.04em] text-foreground">{toolType.name}</h3>
          <p className="text-xs text-subtle">
            {categoryLabels[toolType.category]} · Level {toolType.minimum_fighter_level}+ · Mastery up to{" "}
            {toolType.max_mastery_level}
          </p>
        </div>
        <RarityBadge rarity={toolType.rarity} />
      </div>
      <p className="text-sm text-muted">{toolType.description}</p>
      <ToolStatList baseStats={toolType.base_stats} />
      {!shopItem.is_usable && (
        <p className="flex items-center gap-1.5 text-xs text-amber-300">
          <ShieldAlert className="size-3.5" aria-hidden="true" />
          Usable from fighter level {toolType.minimum_fighter_level}
        </p>
      )}
      <div className="mt-auto flex items-center justify-between gap-3 border-t border-border pt-3">
        <CoinPrice coins={shopItem.price_coins} />
        {shopItem.is_unlocked ? (
          <Button size="sm" onClick={() => onBuy(shopItem)}>
            Buy
          </Button>
        ) : (
          <span className="flex items-center gap-1.5 text-xs text-muted">
            <Lock className="size-3.5" aria-hidden="true" />
            Locked
          </span>
        )}
      </div>
      {!shopItem.is_unlocked && unlockLevel && (
        <p className="rounded-lg bg-background/60 px-3 py-2 text-xs text-muted">
          Win level {unlockLevel.chapter_number}-{unlockLevel.number}, {unlockLevel.title}, to unlock it and get one
          free.
        </p>
      )}
      {shopItem.owned_count > 0 && <p className="text-xs text-up">You own {shopItem.owned_count}</p>}
    </article>
  );
}

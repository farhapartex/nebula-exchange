import { Clock, User } from "lucide-react";

import { Button } from "@/components/ui/button";
import type { MarketListing } from "@/features/market/api/market-types";
import { CoinPrice } from "@/features/market/coin-price";
import { LevelRequirementNote } from "@/features/market/level-requirement-note";
import { MasteryMeter } from "@/features/market/mastery-meter";
import { RarityBadge } from "@/features/market/rarity-badge";
import { ToolArtwork } from "@/features/market/tool-artwork";
import { categoryLabels } from "@/features/market/tool-rarity";
import { ToolStatList } from "@/features/market/tool-stat-list";
import { formatRelativeTime } from "@/utils/time/format-relative-time";

type ListingCardProps = {
  listing: MarketListing;
  levelsNeeded: number;
  onBuy: (listing: MarketListing) => void;
};

export function ListingCard({ listing, levelsNeeded, onBuy }: ListingCardProps) {
  const toolType = listing.tool.tool_type;
  return (
    <article className="flex flex-col gap-3 rounded-2xl border border-border bg-surface/70 p-4">
      <ToolArtwork toolType={toolType} />
      <div className="flex items-start justify-between gap-2">
        <div>
          <h3 className="font-display text-xl leading-tight tracking-[0.04em] text-foreground">{toolType.name}</h3>
          <p className="text-xs text-subtle">
            {categoryLabels[toolType.category]} · {listing.tool.fights_used} fights · {listing.tool.hits_landed} hits
          </p>
        </div>
        <RarityBadge rarity={toolType.rarity} />
      </div>
      <MasteryMeter masteryLevel={listing.tool.mastery_level} maxMasteryLevel={toolType.max_mastery_level} />
      <ToolStatList baseStats={toolType.base_stats} />
      <div className="flex flex-wrap gap-x-4 gap-y-1 text-xs text-subtle">
        <span className="flex items-center gap-1.5">
          <User className="size-3.5" aria-hidden="true" />
          {listing.is_own ? "Your listing" : listing.seller.username}
        </span>
        <span className="flex items-center gap-1.5" title={listing.expires_at}>
          <Clock className="size-3.5" aria-hidden="true" />
          Ends {formatRelativeTime(listing.expires_at)}
        </span>
      </div>
      <div className="mt-auto flex items-center justify-between gap-3 border-t border-border pt-3">
        <CoinPrice coins={listing.price_coins} />
        {!listing.is_own && levelsNeeded > 0 ? (
          <LevelRequirementNote minimumFighterLevel={toolType.minimum_fighter_level} levelsNeeded={levelsNeeded} />
        ) : (
          <Button size="sm" onClick={() => onBuy(listing)} disabled={listing.is_own}>
            {listing.is_own ? "Yours" : "Buy"}
          </Button>
        )}
      </div>
    </article>
  );
}

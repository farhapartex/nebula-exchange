import type { ToolRarity } from "@/features/market/api/market-types";
import { rarityBadgeClassNames, rarityLabels } from "@/features/market/tool-rarity";
import { cn } from "@/utils/class-names";

export function RarityBadge({ rarity }: { rarity: ToolRarity }) {
  return (
    <span className={cn("rounded-md px-2 py-0.5 text-[0.6875rem] font-semibold", rarityBadgeClassNames[rarity])}>
      {rarityLabels[rarity]}
    </span>
  );
}

import type { ToolCategory, ToolRarity } from "@/features/market/api/market-types";

export const toolRarities: ToolRarity[] = ["COMMON", "UNCOMMON", "RARE", "EPIC", "LEGENDARY"];

export const rarityLabels: Record<ToolRarity, string> = {
  COMMON: "Common",
  UNCOMMON: "Uncommon",
  RARE: "Rare",
  EPIC: "Epic",
  LEGENDARY: "Legendary",
};

export const rarityBadgeClassNames: Record<ToolRarity, string> = {
  COMMON: "bg-zinc-400/15 text-zinc-300",
  UNCOMMON: "bg-emerald-400/15 text-emerald-300",
  RARE: "bg-sky-400/15 text-sky-300",
  EPIC: "bg-fuchsia-400/15 text-fuchsia-300",
  LEGENDARY: "bg-amber-400/15 text-amber-300",
};

export const rarityArtworkClassNames: Record<ToolRarity, string> = {
  COMMON: "from-zinc-400/20 text-zinc-300",
  UNCOMMON: "from-emerald-400/25 text-emerald-300",
  RARE: "from-sky-400/25 text-sky-300",
  EPIC: "from-fuchsia-400/30 text-fuchsia-300",
  LEGENDARY: "from-amber-400/35 text-amber-300",
};

export const categoryLabels: Record<ToolCategory, string> = {
  WEAPON: "Weapon",
  GUARD: "Guard",
};

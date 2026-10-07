export type ToolCategory = "WEAPON" | "GUARD";

export type ToolRarity = "COMMON" | "UNCOMMON" | "RARE" | "EPIC" | "LEGENDARY";

export type ToolTypeSummary = {
  id: string;
  name: string;
  description: string;
  category: ToolCategory;
  rarity: ToolRarity;
  base_stats: Record<string, number>;
  max_mastery_level: number;
  minimum_fighter_level: number;
  image_svg: string | null;
};

export type UnlockLevel = {
  id: string;
  chapter_number: number;
  number: number;
  title: string;
};

export type ShopItem = {
  tool_type: ToolTypeSummary;
  price_coins: string;
  is_unlocked: boolean;
  unlock_level: UnlockLevel | null;
  is_usable: boolean;
  owned_count: number;
};

export type ListedTool = {
  id: string;
  tool_type: ToolTypeSummary;
  mastery_level: number;
  fights_used: number;
  hits_landed: number;
};

export type MarketListing = {
  id: string;
  tool: ListedTool;
  seller: { username: string };
  price_coins: string;
  listed_at: string;
  expires_at: string;
  is_own: boolean;
};

export type ListingSort = "PRICE_LOW" | "PRICE_HIGH" | "MASTERY_HIGH" | "NEWEST";

export type MarketFilters = {
  category: ToolCategory | "ALL";
  rarity: ToolRarity | "ALL";
  sort: ListingSort;
  search: string;
};

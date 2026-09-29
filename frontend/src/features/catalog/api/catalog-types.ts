export type ItemCategory = "resource" | "component" | "drill" | "ship" | "consumable" | "legendary";

export type ItemAttributes = {
  drill_multiplier?: string;
  cargo_capacity?: number;
  speed_multiplier?: string;
};

export type CatalogItem = {
  id: number;
  slug: string;
  name: string;
  category: ItemCategory;
  tier: number | null;
  rarity_rank: number;
  is_tradeable: boolean;
  is_auction_only: boolean;
  max_supply: number | null;
  description: string;
  attributes: ItemAttributes;
};

export type ItemQuantity = {
  item_id: number;
  quantity: number;
};

export type Recipe = {
  id: string;
  output_item_id: number;
  output_quantity: number;
  craft_seconds: number;
  fee: string;
  inputs: ItemQuantity[];
};

export type Upgrade = {
  id: string;
  from_item_id: number;
  to_item_id: number;
  craft_fee: string;
  buy_price: string | null;
  inputs: ItemQuantity[];
};

export type LootEntry = {
  item_id: number;
  minimum_quantity: number;
  maximum_quantity: number;
  chance_basis_points: number;
};

export type Zone = {
  id: string;
  name: string;
  description: string;
  duration_seconds: number;
  fuel_cost: number;
  minimum_drill_tier: number;
  allowed_ship_item_ids: number[];
  loot: LootEntry[];
};

export type ShopItem = {
  sku: string;
  name: string;
  description: string;
  price: string;
  contents: ItemQuantity[];
};

export type StarterPackItem = {
  itemID: number;
  quantity: number;
  description: string;
};

export const entryFeeInMicroUnits = "5000000";
export const starterBonusInMicroUnits = "2000000";

export const starterPackItems: StarterPackItem[] = [
  { itemID: 301, quantity: 1, description: "Tier 1 ship with 20 cargo" },
  { itemID: 201, quantity: 1, description: "Tier 1 drill, 1.0× loot" },
  { itemID: 401, quantity: 10, description: "Enough for five Asteroid Belt runs" },
];

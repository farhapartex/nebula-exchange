export type StarterPackEntry = {
  label: string;
  quantity: number;
  description: string;
};

export const entryFeeInMicroUnits = "5000000";

export const starterPackEntries: StarterPackEntry[] = [
  { label: "Scout", quantity: 1, description: "Tier 1 ship with 20 cargo" },
  { label: "Basic Drill", quantity: 1, description: "Tier 1 drill, 1.0× loot" },
  { label: "Fuel Cell", quantity: 10, description: "Enough for five Asteroid Belt runs" },
  { label: "Bonus NC", quantity: 2, description: "Spendable credits to get you going" },
];

import type { ItemCategory } from "@/features/catalog/api/catalog-types";

export type CategoryFilter = ItemCategory | "all";

export const categoryFilters: { value: CategoryFilter; label: string }[] = [
  { value: "all", label: "All" },
  { value: "resource", label: "Resources" },
  { value: "component", label: "Components" },
  { value: "drill", label: "Drills" },
  { value: "ship", label: "Ships" },
  { value: "consumable", label: "Consumables" },
  { value: "legendary", label: "Legendary" },
];

export function matchesCategoryFilter(category: ItemCategory, categoryFilter: CategoryFilter): boolean {
  return categoryFilter === "all" || category === categoryFilter;
}

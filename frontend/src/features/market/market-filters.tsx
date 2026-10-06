import { Search } from "lucide-react";

import type { ListingSort, MarketFilters, ToolCategory, ToolRarity } from "@/features/market/api/market-types";
import { rarityLabels, toolRarities } from "@/features/market/tool-rarity";
import { cn } from "@/utils/class-names";

const categoryChoices: { value: ToolCategory | "ALL"; label: string }[] = [
  { value: "ALL", label: "All" },
  { value: "WEAPON", label: "Weapons" },
  { value: "GUARD", label: "Guards" },
];

const sortChoices: { value: ListingSort; label: string }[] = [
  { value: "PRICE_LOW", label: "Lowest price" },
  { value: "PRICE_HIGH", label: "Highest price" },
  { value: "MASTERY_HIGH", label: "Highest mastery" },
  { value: "NEWEST", label: "Newest" },
];

const selectClassName =
  "h-9 rounded-lg border border-border-strong bg-surface-raised px-3 text-sm text-foreground focus-visible:ring-2 focus-visible:ring-highlight focus-visible:outline-none";

type MarketFiltersBarProps = {
  filters: MarketFilters;
  onChange: (changes: Partial<MarketFilters>) => void;
  showsSort: boolean;
};

export function MarketFiltersBar({ filters, onChange, showsSort }: MarketFiltersBarProps) {
  return (
    <div className="flex flex-wrap items-center gap-3">
      <div
        role="radiogroup"
        aria-label="Tool type"
        className="flex rounded-lg border border-border-strong bg-surface-raised p-0.5"
      >
        {categoryChoices.map((categoryChoice) => {
          const isSelected = filters.category === categoryChoice.value;
          return (
            <button
              key={categoryChoice.value}
              type="button"
              role="radio"
              aria-checked={isSelected}
              onClick={() => onChange({ category: categoryChoice.value })}
              className={cn(
                "rounded-md px-3 py-1.5 text-sm transition-colors",
                isSelected ? "bg-accent text-background" : "text-muted hover:text-foreground",
              )}
            >
              {categoryChoice.label}
            </button>
          );
        })}
      </div>
      <select
        aria-label="Rarity"
        value={filters.rarity}
        onChange={(changeEvent) => onChange({ rarity: changeEvent.target.value as ToolRarity | "ALL" })}
        className={selectClassName}
      >
        <option value="ALL">All rarities</option>
        {toolRarities.map((rarity) => (
          <option key={rarity} value={rarity}>
            {rarityLabels[rarity]}
          </option>
        ))}
      </select>
      {showsSort && (
        <select
          aria-label="Sort"
          value={filters.sort}
          onChange={(changeEvent) => onChange({ sort: changeEvent.target.value as ListingSort })}
          className={selectClassName}
        >
          {sortChoices.map((sortChoice) => (
            <option key={sortChoice.value} value={sortChoice.value}>
              {sortChoice.label}
            </option>
          ))}
        </select>
      )}
      <label className="relative ml-auto w-full sm:w-64">
        <span className="sr-only">Search tools</span>
        <Search
          className="pointer-events-none absolute top-1/2 left-3 size-4 -translate-y-1/2 text-subtle"
          aria-hidden="true"
        />
        <input
          type="search"
          value={filters.search}
          onChange={(changeEvent) => onChange({ search: changeEvent.target.value })}
          placeholder="Search tools"
          className="h-9 w-full rounded-lg border border-border-strong bg-surface-raised pr-3 pl-9 text-sm text-foreground placeholder:text-subtle focus-visible:ring-2 focus-visible:ring-highlight focus-visible:outline-none"
        />
      </label>
    </div>
  );
}

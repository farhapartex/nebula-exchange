"use client";

import { Tabs, TabsList, TabsTrigger } from "@/components/ui/tabs";
import { categoryFilters, type CategoryFilter } from "@/features/catalog/category-filters";

type CategoryFilterTabsProps = {
  value: CategoryFilter;
  onValueChange: (categoryFilter: CategoryFilter) => void;
  countsByFilter?: Partial<Record<CategoryFilter, number>>;
};

export function CategoryFilterTabs({ value, onValueChange, countsByFilter }: CategoryFilterTabsProps) {
  return (
    <Tabs value={value} onValueChange={(nextFilter) => onValueChange(nextFilter as CategoryFilter)}>
      <TabsList>
        {categoryFilters.map((filterOption) => {
          const filterCount = countsByFilter?.[filterOption.value];
          return (
            <TabsTrigger key={filterOption.value} value={filterOption.value} disabled={filterCount === 0}>
              {filterOption.label}
              {filterCount !== undefined && (
                <span className="ml-1.5 text-xs text-subtle tabular-nums">{filterCount}</span>
              )}
            </TabsTrigger>
          );
        })}
      </TabsList>
    </Tabs>
  );
}

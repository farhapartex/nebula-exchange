import type { CatalogItem } from "@/features/catalog/api/catalog-types";
import { listItemAttributes } from "@/features/catalog/item-attributes";

export function ItemAttributeList({ item }: { item: CatalogItem }) {
  const itemAttributes = listItemAttributes(item);
  if (itemAttributes.length === 0) {
    return null;
  }
  return (
    <dl className="flex flex-wrap gap-x-4 gap-y-1 text-xs">
      {itemAttributes.map((itemAttribute) => (
        <div key={itemAttribute.label} className="flex gap-1">
          <dt className="text-subtle">{itemAttribute.label}</dt>
          <dd className="font-medium text-foreground tabular-nums">{itemAttribute.value}</dd>
        </div>
      ))}
    </dl>
  );
}

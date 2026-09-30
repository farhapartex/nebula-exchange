export type SortDirection = "ascending" | "descending";

export type SortableValue = string | number | bigint;

export function sortRows<Row>(rows: Row[], sortValueOf: (row: Row) => SortableValue, direction: SortDirection): Row[] {
  const directionFactor = direction === "ascending" ? 1 : -1;
  return [...rows].sort((firstRow, secondRow) => {
    const firstValue = sortValueOf(firstRow);
    const secondValue = sortValueOf(secondRow);
    if (firstValue === secondValue) {
      return 0;
    }
    return (firstValue < secondValue ? -1 : 1) * directionFactor;
  });
}

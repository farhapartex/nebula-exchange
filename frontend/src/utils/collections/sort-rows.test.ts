import { describe, expect, it } from "vitest";

import { sortRows } from "@/utils/collections/sort-rows";

describe("sortRows", () => {
  const rows = [
    { name: "Iron", amount: 20n },
    { name: "Crystal", amount: 9_223_372_036_854_775_807n },
    { name: "Alloy", amount: -5n },
  ];

  it("sorts BigInt values exactly in both directions", () => {
    expect(sortRows(rows, (row) => row.amount, "ascending").map((row) => row.name)).toEqual([
      "Alloy",
      "Iron",
      "Crystal",
    ]);
    expect(sortRows(rows, (row) => row.amount, "descending").map((row) => row.name)).toEqual([
      "Crystal",
      "Iron",
      "Alloy",
    ]);
  });

  it("does not mutate the original rows", () => {
    sortRows(rows, (row) => row.name, "ascending");
    expect(rows[0].name).toBe("Iron");
  });
});

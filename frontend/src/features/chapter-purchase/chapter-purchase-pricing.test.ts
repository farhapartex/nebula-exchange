import { describe, expect, it } from "vitest";

import type { PurchasableChapter } from "@/features/chapter-purchase/api/chapter-catalog-api";
import {
  chaptersStillToBuy,
  discountPercentFor,
  quoteChapterPurchase,
} from "@/features/chapter-purchase/chapter-purchase-pricing";

function chapter(number: number, isOwned = false): PurchasableChapter {
  return { id: String(number), number, title: `Chapter ${number}`, price_cents: "499", is_owned: isOwned };
}

const fiveChapters = [chapter(3), chapter(1), chapter(2), chapter(5), chapter(4)];

describe("chapter purchase pricing", () => {
  it("only offers chapters the player does not own, in story order", () => {
    expect(chaptersStillToBuy([chapter(2), chapter(1, true), chapter(3)]).map((item) => item.number)).toEqual([2, 3]);
  });

  it("grows the discount with the number of chapters", () => {
    expect(discountPercentFor(1, false)).toBe(0);
    expect(discountPercentFor(2, false)).toBe(5);
    expect(discountPercentFor(3, false)).toBe(10);
    expect(discountPercentFor(4, false)).toBe(10);
    expect(discountPercentFor(5, true)).toBe(20);
  });

  it("charges the next chapter at full price", () => {
    const quote = quoteChapterPurchase(chaptersStillToBuy(fiveChapters), "next_chapter", 2);
    expect(quote.chapters.map((item) => item.number)).toEqual([1]);
    expect(quote.totalCents).toBe(499n);
  });

  it("charges several chapters from the next one with the count discount", () => {
    const quote = quoteChapterPurchase(chaptersStillToBuy(fiveChapters), "several_chapters", 3);
    expect(quote.chapters.map((item) => item.number)).toEqual([1, 2, 3]);
    expect(quote.subtotalCents).toBe(1497n);
    expect(quote.discountCents).toBe(150n);
    expect(quote.totalCents).toBe(1347n);
  });

  it("gives 20 percent off for every remaining chapter", () => {
    const quote = quoteChapterPurchase(chaptersStillToBuy(fiveChapters), "all_chapters", 2);
    expect(quote.subtotalCents).toBe(2495n);
    expect(quote.discountPercent).toBe(20);
    expect(quote.totalCents).toBe(1996n);
  });
});

import type { PurchasableChapter } from "@/features/chapter-purchase/api/chapter-catalog-api";

export type PurchaseChoice = "next_chapter" | "several_chapters" | "all_chapters";

export type ChapterPurchaseQuote = {
  chapters: PurchasableChapter[];
  subtotalCents: bigint;
  discountPercent: number;
  discountCents: bigint;
  totalCents: bigint;
};

export function chaptersStillToBuy(chapters: PurchasableChapter[]): PurchasableChapter[] {
  return chapters.filter((chapter) => !chapter.is_owned).sort((first, second) => first.number - second.number);
}

export function discountPercentFor(chapterCount: number, isEveryChapter: boolean): number {
  if (isEveryChapter && chapterCount >= 2) {
    return 20;
  }
  if (chapterCount >= 3) {
    return 10;
  }
  return chapterCount === 2 ? 5 : 0;
}

export function chaptersForChoice(
  chaptersToBuy: PurchasableChapter[],
  purchaseChoice: PurchaseChoice,
  severalCount: number,
): PurchasableChapter[] {
  if (purchaseChoice === "all_chapters") {
    return chaptersToBuy;
  }
  if (purchaseChoice === "several_chapters") {
    return chaptersToBuy.slice(0, severalCount);
  }
  return chaptersToBuy.slice(0, 1);
}

export function quoteChapterPurchase(
  chaptersToBuy: PurchasableChapter[],
  purchaseChoice: PurchaseChoice,
  severalCount: number,
): ChapterPurchaseQuote {
  const chapters = chaptersForChoice(chaptersToBuy, purchaseChoice, severalCount);
  const subtotalCents = chapters.reduce((sum, chapter) => sum + BigInt(chapter.price_cents), 0n);
  const isEveryChapter = chapters.length === chaptersToBuy.length;
  const discountPercent = discountPercentFor(chapters.length, purchaseChoice === "all_chapters" && isEveryChapter);
  const discountCents = (subtotalCents * BigInt(discountPercent) + 50n) / 100n;
  return { chapters, subtotalCents, discountPercent, discountCents, totalCents: subtotalCents - discountCents };
}

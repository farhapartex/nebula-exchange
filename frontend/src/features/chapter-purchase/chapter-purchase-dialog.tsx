"use client";

import { useState } from "react";
import { useQuery } from "@tanstack/react-query";
import { Minus, Plus } from "lucide-react";

import { Button } from "@/components/ui/button";
import { Dialog } from "@/components/ui/dialog";
import { Skeleton } from "@/components/ui/skeleton";
import {
  fetchPurchasableChapters,
  purchasableChaptersQueryKey,
} from "@/features/chapter-purchase/api/chapter-catalog-api";
import {
  chaptersStillToBuy,
  discountPercentFor,
  quoteChapterPurchase,
  type PurchaseChoice,
} from "@/features/chapter-purchase/chapter-purchase-pricing";
import { PurchaseChoiceCard } from "@/features/chapter-purchase/purchase-choice-card";
import { PurchaseSummary } from "@/features/chapter-purchase/purchase-summary";
import { publishToastEvent } from "@/lib/notifications/toast-events";
import { formatUsd } from "@/utils/money/format-usd";

const mockCheckoutDelayInMilliseconds = 900;

type ChapterPurchaseDialogProps = {
  isOpen: boolean;
  onOpenChange: (isOpen: boolean) => void;
};

export function ChapterPurchaseDialog({ isOpen, onOpenChange }: ChapterPurchaseDialogProps) {
  const chaptersQuery = useQuery({
    queryKey: purchasableChaptersQueryKey,
    queryFn: fetchPurchasableChapters,
    enabled: isOpen,
  });
  const [purchaseChoice, setPurchaseChoice] = useState<PurchaseChoice>("next_chapter");
  const [severalCount, setSeveralCount] = useState(2);
  const [isPaying, setIsPaying] = useState(false);

  const chaptersToBuy = chaptersStillToBuy(chaptersQuery.data ?? []);
  const canChooseSeveral = chaptersToBuy.length >= 3;
  const canChooseAll = chaptersToBuy.length >= 2;
  const maximumSeveralCount = chaptersToBuy.length - 1;
  const quote = quoteChapterPurchase(chaptersToBuy, purchaseChoice, severalCount);
  const nextChapter = chaptersToBuy[0];

  function pay() {
    setIsPaying(true);
    window.setTimeout(() => {
      setIsPaying(false);
      onOpenChange(false);
      publishToastEvent({
        tone: "info",
        title: "Checkout is not connected yet",
        description: `This will open Stripe to pay ${formatUsd(quote.totalCents)} for ${quote.chapters.length} ${quote.chapters.length === 1 ? "chapter" : "chapters"}.`,
      });
    }, mockCheckoutDelayInMilliseconds);
  }

  return (
    <Dialog
      isOpen={isOpen}
      onOpenChange={onOpenChange}
      title="Unlock the story"
      description="Pay once with your card. Unlocked chapters stay yours to play and replay."
    >
      {!nextChapter ? (
        <div className="space-y-3">
          <Skeleton className="h-20 rounded-xl" />
          <Skeleton className="h-20 rounded-xl" />
          <Skeleton className="h-32 rounded-xl" />
        </div>
      ) : (
        <div className="space-y-6">
          <fieldset className="space-y-3">
            <legend className="sr-only">How many chapters do you want to unlock?</legend>
            <PurchaseChoiceCard
              choice="next_chapter"
              selectedChoice={purchaseChoice}
              onSelect={setPurchaseChoice}
              title="Chapter by chapter"
              detail={`Chapter ${nextChapter.number} · ${nextChapter.title}`}
              price={formatUsd(BigInt(nextChapter.price_cents))}
            />
            {canChooseSeveral && (
              <PurchaseChoiceCard
                choice="several_chapters"
                selectedChoice={purchaseChoice}
                onSelect={setPurchaseChoice}
                title="A number of chapters"
                detail={`Chapters ${nextChapter.number} to ${chaptersToBuy[severalCount - 1].number}`}
                discountPercent={discountPercentFor(severalCount, false)}
              >
                <div className="mt-3 flex items-center gap-3">
                  <Button
                    type="button"
                    size="sm"
                    variant="secondary"
                    aria-label="Fewer chapters"
                    disabled={severalCount <= 2}
                    onClick={() => {
                      setPurchaseChoice("several_chapters");
                      setSeveralCount((count) => Math.max(2, count - 1));
                    }}
                  >
                    <Minus className="size-3.5" />
                  </Button>
                  <span className="w-24 text-center font-mono text-sm whitespace-nowrap text-foreground tabular-nums">
                    {severalCount} chapters
                  </span>
                  <Button
                    type="button"
                    size="sm"
                    variant="secondary"
                    aria-label="More chapters"
                    disabled={severalCount >= maximumSeveralCount}
                    onClick={() => {
                      setPurchaseChoice("several_chapters");
                      setSeveralCount((count) => Math.min(maximumSeveralCount, count + 1));
                    }}
                  >
                    <Plus className="size-3.5" />
                  </Button>
                </div>
              </PurchaseChoiceCard>
            )}
            {canChooseAll && (
              <PurchaseChoiceCard
                choice="all_chapters"
                selectedChoice={purchaseChoice}
                onSelect={setPurchaseChoice}
                title="All chapters"
                detail={`All ${chaptersToBuy.length} chapters of the story`}
                discountPercent={discountPercentFor(chaptersToBuy.length, true)}
              />
            )}
          </fieldset>

          <PurchaseSummary quote={quote} />

          <div className="flex flex-col-reverse gap-3 sm:flex-row sm:justify-end">
            <Button type="button" variant="secondary" size="lg" onClick={() => onOpenChange(false)} disabled={isPaying}>
              Cancel
            </Button>
            <Button type="button" size="lg" onClick={pay} isLoading={isPaying} className="sm:min-w-52">
              {isPaying ? "Opening checkout" : `Pay ${formatUsd(quote.totalCents)}`}
            </Button>
          </div>
        </div>
      )}
    </Dialog>
  );
}

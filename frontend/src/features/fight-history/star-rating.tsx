import { Star } from "lucide-react";

import { cn } from "@/utils/class-names";

const maximumStars = 3;

export function StarRating({ stars }: { stars: number | null }) {
  const earnedStars = stars ?? 0;
  return (
    <span className="flex items-center gap-0.5" aria-label={`${earnedStars} of ${maximumStars} stars`}>
      {Array.from({ length: maximumStars }, (_, starIndex) => (
        <Star
          key={starIndex}
          aria-hidden="true"
          className={cn("size-4", starIndex < earnedStars ? "fill-amber-400 text-amber-400" : "text-border-strong")}
        />
      ))}
    </span>
  );
}

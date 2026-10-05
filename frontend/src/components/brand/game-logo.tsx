import Link from "next/link";
import { HandFist } from "lucide-react";

import { gameBrand } from "@/lib/brand/game-brand";
import { cn } from "@/utils/class-names";

type GameLogoProps = {
  href?: string;
  wordmarkClassName?: string;
};

export function GameLogo({ href = "/", wordmarkClassName }: GameLogoProps) {
  return (
    <Link href={href} className="group flex shrink-0 items-center gap-2.5" aria-label={`${gameBrand.title} home`}>
      <span className="flex size-9 -rotate-6 items-center justify-center rounded-md bg-linear-to-br from-amber-400 via-accent to-red-600 shadow-[0_0_22px_-4px] shadow-accent/70 transition-transform group-hover:rotate-0">
        <HandFist className="size-5 text-background" strokeWidth={2.5} aria-hidden="true" />
      </span>
      <span
        className={cn(
          "font-display text-[1.65rem] leading-none tracking-[0.06em] whitespace-nowrap text-foreground",
          wordmarkClassName,
        )}
      >
        {gameBrand.titleLead} <span className="text-accent">{gameBrand.titleAccent}</span>
      </span>
    </Link>
  );
}

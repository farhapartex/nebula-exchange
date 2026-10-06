import Link from "next/link";
import { ArrowLeft, Swords } from "lucide-react";

import { Button } from "@/components/ui/button";

export function FightArenaPlaceholder() {
  return (
    <div className="flex min-h-svh flex-col items-center justify-center px-6 text-center">
      <span className="flex size-16 items-center justify-center rounded-2xl bg-accent/10 text-accent motion-safe:animate-pop-in">
        <Swords className="size-8" aria-hidden="true" />
      </span>
      <h1 className="mt-6 font-display text-5xl tracking-[0.03em] text-foreground">The fight starts here</h1>
      <p className="mt-3 max-w-md text-muted">
        The fight screen arrives once we decide how fighting plays. For now, this is where the game will begin.
      </p>
      <Button asChild variant="secondary" className="mt-10">
        <Link href="/fight">
          <ArrowLeft className="size-4" aria-hidden="true" />
          Back to the hub
        </Link>
      </Button>
    </div>
  );
}

import Link from "next/link";
import { Target } from "lucide-react";

import { Button } from "@/components/ui/button";
import { HubPanel } from "@/features/fight-hub/hub-panel";
import { levelRoute } from "@/features/fight-hub/story-progress";

export function TrainingCard() {
  return (
    <HubPanel>
      <div className="flex items-start gap-4">
        <span className="flex size-11 shrink-0 items-center justify-center rounded-xl bg-highlight/10 text-highlight">
          <Target className="size-5" aria-hidden="true" />
        </span>
        <div className="flex-1">
          <p className="font-display text-2xl tracking-[0.04em] text-foreground">Training ground</p>
          <p className="mt-1 text-sm text-muted">Practise your moves against a dummy. No risk, no rewards.</p>
        </div>
      </div>
      <Button asChild variant="secondary" className="mt-4 w-full">
        <Link href={levelRoute("training")}>Train</Link>
      </Button>
    </HubPanel>
  );
}

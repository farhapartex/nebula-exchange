import Link from "next/link";
import { RotateCcw, Swords, Timer } from "lucide-react";

import type { WonFight } from "@/features/fight-history/api/fight-history-api";
import { StarRating } from "@/features/fight-history/star-rating";
import { formatCountdown } from "@/utils/time/format-duration";
import { formatRelativeTime } from "@/utils/time/format-relative-time";

const millisecondsPerSecond = 1000;

export function WonFightRow({ wonFight }: { wonFight: WonFight }) {
  return (
    <li className="flex flex-wrap items-center gap-x-6 gap-y-3 rounded-xl border border-border bg-background/50 px-4 py-3">
      <div className="flex min-w-48 flex-1 items-center gap-3">
        <span className="flex h-9 min-w-12 items-center justify-center rounded-lg bg-accent/10 px-2 font-mono text-sm text-accent-soft">
          {wonFight.level.chapter_number}-{wonFight.level.number}
        </span>
        <div className="min-w-0">
          <p className="truncate font-medium text-foreground">{wonFight.level.title}</p>
          <p className="text-xs text-subtle" title={wonFight.finished_at}>
            Won {formatRelativeTime(wonFight.finished_at)}
          </p>
        </div>
      </div>
      <StarRating stars={wonFight.stars} />
      <dl className="flex gap-5 text-sm">
        <div className="flex items-center gap-1.5" title="Time taken">
          <dt>
            <Timer className="size-4 text-subtle" aria-label="Time taken" />
          </dt>
          <dd className="font-mono text-foreground tabular-nums">
            {formatCountdown(wonFight.duration_ms / millisecondsPerSecond)}
          </dd>
        </div>
        <div className="flex items-center gap-1.5" title="Damage dealt and taken">
          <dt>
            <Swords className="size-4 text-subtle" aria-label="Damage dealt and taken" />
          </dt>
          <dd className="font-mono tabular-nums">
            <span className="text-up">{wonFight.damage_dealt}</span>
            <span className="text-subtle"> / </span>
            <span className="text-down">{wonFight.damage_taken}</span>
          </dd>
        </div>
      </dl>
      <Link
        href={`/fight/${encodeURIComponent(wonFight.level.id)}`}
        className="ml-auto flex items-center gap-1.5 rounded-lg px-2.5 py-1.5 text-xs text-muted transition-colors hover:bg-surface-raised hover:text-foreground"
      >
        <RotateCcw className="size-3.5" aria-hidden="true" />
        Play again
      </Link>
    </li>
  );
}

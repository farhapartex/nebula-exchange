import { HubPanel } from "@/features/fight-hub/hub-panel";
import type { ChapterWins } from "@/features/fight-history/group-wins-by-chapter";
import { WonFightRow } from "@/features/fight-history/won-fight-row";

export function ChapterWinsSection({ chapterWins }: { chapterWins: ChapterWins }) {
  return (
    <HubPanel>
      <div className="mb-4 flex flex-wrap items-baseline justify-between gap-2">
        <div>
          <p className="text-xs font-semibold tracking-[0.18em] text-subtle uppercase">
            Chapter {chapterWins.chapterNumber}
          </p>
          <h2 className="font-display text-2xl tracking-[0.04em] text-foreground">{chapterWins.chapterTitle}</h2>
        </div>
        <span className="text-sm text-muted">
          {chapterWins.wins.length} {chapterWins.wins.length === 1 ? "win" : "wins"}
        </span>
      </div>
      <ul className="space-y-2">
        {chapterWins.wins.map((wonFight) => (
          <WonFightRow key={wonFight.id} wonFight={wonFight} />
        ))}
      </ul>
    </HubPanel>
  );
}

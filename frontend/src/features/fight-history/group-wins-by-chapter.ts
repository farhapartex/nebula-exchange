import type { WonFight } from "@/features/fight-history/api/fight-history-api";

export type ChapterWins = {
  chapterNumber: number;
  chapterTitle: string;
  wins: WonFight[];
};

export function groupWinsByChapter(wonFights: WonFight[]): ChapterWins[] {
  const chapterGroups: ChapterWins[] = [];
  for (const wonFight of wonFights) {
    const lastGroup = chapterGroups.at(-1);
    if (lastGroup && lastGroup.chapterNumber === wonFight.level.chapter_number) {
      lastGroup.wins.push(wonFight);
      continue;
    }
    chapterGroups.push({
      chapterNumber: wonFight.level.chapter_number,
      chapterTitle: wonFight.level.chapter_title,
      wins: [wonFight],
    });
  }
  return chapterGroups;
}

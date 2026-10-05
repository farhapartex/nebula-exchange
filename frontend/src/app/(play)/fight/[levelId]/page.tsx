import type { Metadata } from "next";

import { LevelPlayView } from "@/features/level-intro/level-play-view";

export const metadata: Metadata = {
  title: "Fight",
};

export default async function LevelPlayPage({ params }: { params: Promise<{ levelId: string }> }) {
  const { levelId } = await params;
  return <LevelPlayView levelID={levelId} />;
}

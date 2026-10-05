const fighterRanks = [
  { minimumLevel: 10, title: "Contender" },
  { minimumLevel: 6, title: "Brawler" },
  { minimumLevel: 3, title: "Street rat" },
  { minimumLevel: 1, title: "Nobody" },
];

export function rankTitleForLevel(fighterLevel: number): string {
  return fighterRanks.find((fighterRank) => fighterLevel >= fighterRank.minimumLevel)?.title ?? "Nobody";
}

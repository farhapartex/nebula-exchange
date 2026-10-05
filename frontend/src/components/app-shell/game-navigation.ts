export type GameNavigationLink = {
  href: string;
  label: string;
  isAvailable: boolean;
};

export const gameNavigationLinks: GameNavigationLink[] = [
  { href: "/fight", label: "Fight", isAvailable: true },
  { href: "/arsenal", label: "Arsenal", isAvailable: false },
  { href: "/market", label: "Market", isAvailable: false },
  { href: "/wallet", label: "Wallet", isAvailable: false },
];

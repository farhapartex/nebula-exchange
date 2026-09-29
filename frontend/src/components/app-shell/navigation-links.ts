import {
  Bell,
  Boxes,
  CandlestickChart,
  Gavel,
  History,
  LayoutDashboard,
  Rocket,
  Settings,
  ShoppingBag,
  Wallet,
  Wrench,
  type LucideIcon,
} from "lucide-react";

export type GameSection =
  | "hangar"
  | "missions"
  | "workshop"
  | "shop"
  | "exchange"
  | "auctions"
  | "wallet"
  | "inventory"
  | "history"
  | "notifications"
  | "settings";

export type NavigationLink = {
  section: GameSection;
  href: `/${string}`;
  label: string;
  description: string;
  icon: LucideIcon;
  plannedTask: string;
};

export const primaryNavigationLinks: NavigationLink[] = [
  {
    section: "hangar",
    href: "/hangar",
    label: "Hangar",
    description: "Your dashboard: running missions, craft queue, orders and auctions at a glance.",
    icon: LayoutDashboard,
    plannedTask: "T-034",
  },
  {
    section: "missions",
    href: "/missions",
    label: "Missions",
    description: "Send ships to mining zones and collect the loot they bring back.",
    icon: Rocket,
    plannedTask: "T-042",
  },
  {
    section: "workshop",
    href: "/workshop",
    label: "Workshop",
    description: "Craft components from resources and upgrade your drills and ships.",
    icon: Wrench,
    plannedTask: "T-048",
  },
  {
    section: "shop",
    href: "/shop",
    label: "Shop",
    description: "Buy fuel, ships, drills and bundles at fixed prices.",
    icon: ShoppingBag,
    plannedTask: "T-039",
  },
  {
    section: "exchange",
    href: "/exchange",
    label: "Exchange",
    description: "Trade every item with other players on a live order book.",
    icon: CandlestickChart,
    plannedTask: "T-056",
  },
  {
    section: "auctions",
    href: "/auctions",
    label: "Auctions",
    description: "Bid on legendary drops and player lots, or auction your own items.",
    icon: Gavel,
    plannedTask: "T-071",
  },
  {
    section: "wallet",
    href: "/wallet",
    label: "Wallet",
    description: "Top up NC, see your balance breakdown, and move items and USDC on-chain.",
    icon: Wallet,
    plannedTask: "T-036",
  },
];

export const profileNavigationLinks: NavigationLink[] = [
  {
    section: "inventory",
    href: "/inventory",
    label: "Inventory",
    description: "Everything you own, with available and held quantities.",
    icon: Boxes,
    plannedTask: "T-028",
  },
  {
    section: "history",
    href: "/history",
    label: "History",
    description: "Every ledger movement, payment, trade, mission, craft and auction.",
    icon: History,
    plannedTask: "T-030",
  },
  {
    section: "notifications",
    href: "/notifications",
    label: "Notifications",
    description: "Everything the game has told you, newest first.",
    icon: Bell,
    plannedTask: "T-052",
  },
  {
    section: "settings",
    href: "/settings",
    label: "Settings",
    description: "Profile, password, two-factor authentication and active sessions.",
    icon: Settings,
    plannedTask: "T-021",
  },
];

const allNavigationLinks = [...primaryNavigationLinks, ...profileNavigationLinks];

export function navigationLinkFor(section: GameSection): NavigationLink {
  const matchingLink = allNavigationLinks.find((navigationLink) => navigationLink.section === section);
  if (!matchingLink) {
    throw new Error(`No navigation link for section "${section}"`);
  }
  return matchingLink;
}

export function isNavigationLinkActive(navigationLink: NavigationLink, currentPathname: string): boolean {
  return currentPathname === navigationLink.href || currentPathname.startsWith(`${navigationLink.href}/`);
}

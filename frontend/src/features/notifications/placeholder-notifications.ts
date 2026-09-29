import type { GameNotification } from "@/features/notifications/notification-types";

function minutesAgo(minutes: number) {
  return new Date(Date.now() - minutes * 60_000).toISOString();
}

export function createPlaceholderNotifications(): GameNotification[] {
  return [
    {
      id: "notification-1",
      kind: "mission_completed",
      title: "Mission complete",
      body: "Your Scout is back from the Asteroid Belt with 12 Iron Ore and 4 Copper Ore.",
      href: "/missions",
      createdAt: minutesAgo(2),
      isRead: false,
    },
    {
      id: "notification-2",
      kind: "auction_outbid",
      title: "You were outbid",
      body: "Someone bid 4.20 NC on Relic Drill of Orion. Your 4.00 NC hold was released.",
      href: "/auctions",
      createdAt: minutesAgo(18),
      isRead: false,
    },
    {
      id: "notification-3",
      kind: "order_filled",
      title: "Order filled",
      body: "Bought 40 IRON at 0.0100 NC.",
      href: "/exchange",
      createdAt: minutesAgo(64),
      isRead: false,
    },
    {
      id: "notification-4",
      kind: "craft_completed",
      title: "Craft complete",
      body: "5 Alloy Plates were added to your inventory.",
      href: "/workshop",
      createdAt: minutesAgo(190),
      isRead: true,
    },
    {
      id: "notification-5",
      kind: "payment_succeeded",
      title: "Payment succeeded",
      body: "Entry fee paid. Your starter pack is in the hangar.",
      href: "/hangar",
      createdAt: minutesAgo(1500),
      isRead: true,
    },
  ];
}

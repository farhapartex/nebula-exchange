import type { ApiNotification } from "@/features/notifications/api/notifications-api";

function minutesAgo(minutes: number) {
  return new Date(Date.now() - minutes * 60_000).toISOString();
}

export function createMockNotifications(): ApiNotification[] {
  return [
    {
      id: "notification-1",
      kind: "mission_completed",
      title: "Mission complete",
      body: "Your Scout is back from the Asteroid Belt with 12 Iron Ore and 4 Copper Ore.",
      link: "/missions",
      created_at: minutesAgo(2),
      read_at: null,
    },
    {
      id: "notification-2",
      kind: "auction_outbid",
      title: "You were outbid",
      body: "Someone bid 4.20 NC on Relic Drill of Orion. Your 4.00 NC hold was released.",
      link: "/auctions",
      created_at: minutesAgo(18),
      read_at: null,
    },
    {
      id: "notification-3",
      kind: "order_filled",
      title: "Order filled",
      body: "Bought 40 IRON at 0.0100 NC.",
      link: "/exchange",
      created_at: minutesAgo(64),
      read_at: null,
    },
    {
      id: "notification-4",
      kind: "craft_completed",
      title: "Craft complete",
      body: "5 Alloy Plates were added to your inventory.",
      link: "/workshop",
      created_at: minutesAgo(190),
      read_at: minutesAgo(1),
    },
    {
      id: "notification-5",
      kind: "payment_succeeded",
      title: "Payment succeeded",
      body: "Entry fee paid. Your starter pack is in the hangar.",
      link: "/hangar",
      created_at: minutesAgo(1500),
      read_at: minutesAgo(1),
    },
  ];
}

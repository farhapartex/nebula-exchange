export type NotificationKind =
  | "payment_succeeded"
  | "payment_failed"
  | "payment_needs_attention"
  | "mission_completed"
  | "craft_completed"
  | "order_filled"
  | "order_cancelled"
  | "auction_outbid"
  | "auction_won"
  | "auction_sold"
  | "withdrawal_completed"
  | "account_frozen";

export type GameNotification = {
  id: string;
  kind: NotificationKind;
  title: string;
  body: string;
  href: string;
  createdAt: string;
  isRead: boolean;
};

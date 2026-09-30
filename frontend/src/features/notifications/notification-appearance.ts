import {
  ArrowDownToLine,
  CircleCheck,
  CircleX,
  Gavel,
  Hammer,
  Rocket,
  ShieldAlert,
  TrendingDown,
  Trophy,
  XCircle,
  type LucideIcon,
} from "lucide-react";

import type { NotificationKind } from "@/features/notifications/notification-types";

type NotificationAppearance = {
  icon: LucideIcon;
  iconClassName: string;
};

export const fallbackNotificationAppearance: NotificationAppearance = {
  icon: CircleCheck,
  iconClassName: "bg-surface-raised text-muted",
};

export const notificationAppearanceByKind: Record<NotificationKind, NotificationAppearance> = {
  payment_succeeded: { icon: CircleCheck, iconClassName: "bg-up/10 text-up" },
  payment_failed: { icon: CircleX, iconClassName: "bg-down/10 text-down" },
  payment_needs_attention: { icon: ShieldAlert, iconClassName: "bg-warning/10 text-warning" },
  mission_completed: { icon: Rocket, iconClassName: "bg-highlight/10 text-highlight" },
  craft_completed: { icon: Hammer, iconClassName: "bg-accent/15 text-accent-soft" },
  order_filled: { icon: CircleCheck, iconClassName: "bg-up/10 text-up" },
  order_cancelled: { icon: XCircle, iconClassName: "bg-surface-raised text-muted" },
  auction_outbid: { icon: TrendingDown, iconClassName: "bg-warning/10 text-warning" },
  auction_won: { icon: Trophy, iconClassName: "bg-warning/10 text-warning" },
  auction_sold: { icon: Gavel, iconClassName: "bg-up/10 text-up" },
  withdrawal_completed: { icon: ArrowDownToLine, iconClassName: "bg-info/10 text-info" },
  account_frozen: { icon: ShieldAlert, iconClassName: "bg-down/10 text-down" },
};

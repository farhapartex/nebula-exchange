import { Coins, Flame, ShoppingBag, type LucideIcon } from "lucide-react";

export type TradeLoopStep = {
  title: string;
  body: string;
  icon: LucideIcon;
};

export const tradeLoopSteps: TradeLoopStep[] = [
  { title: "Buy tools", body: "Pick the weapons that fit your fighting style.", icon: ShoppingBag },
  { title: "Make them powerful", body: "Every fight you win makes your tools stronger.", icon: Flame },
  { title: "Sell and earn", body: "Powerful tools are worth more. Sell them to other fighters.", icon: Coins },
];

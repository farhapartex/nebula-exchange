import { CandlestickChart, CreditCard, Rocket, Wrench, type LucideIcon } from "lucide-react";

import type { OnboardingStepKey } from "@/features/hangar/api/onboarding-api";

export const onboardingStepPresentation: Record<
  OnboardingStepKey,
  { title: string; description: string; href: string; actionLabel: string; icon: LucideIcon }
> = {
  paid_entry_fee: {
    title: "Pay the entry fee",
    description: "Unlocks the game and your starter pack.",
    href: "/onboarding/pay",
    actionLabel: "Pay",
    icon: CreditCard,
  },
  first_mission: {
    title: "Finish your first mission",
    description: "Send your Scout to the Asteroid Belt and collect the loot.",
    href: "/missions",
    actionLabel: "Launch",
    icon: Rocket,
  },
  first_craft: {
    title: "Craft your first component",
    description: "Turn Iron and Copper into an Alloy Plate in the workshop.",
    href: "/workshop",
    actionLabel: "Craft",
    icon: Wrench,
  },
  first_trade: {
    title: "Make your first trade",
    description: "Sell resources to other pilots on the exchange.",
    href: "/exchange",
    actionLabel: "Trade",
    icon: CandlestickChart,
  },
};

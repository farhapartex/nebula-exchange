import type { JournalType } from "@/features/history/api/ledger-history-api";

export const journalTypeLabels: Record<JournalType, string> = {
  ENTRY_FEE: "Entry fee",
  TOPUP_CARD: "Card top-up",
  TOPUP_CRYPTO: "Crypto top-up",
  SHOP_PURCHASE: "Shop purchase",
  STARTER_PACK: "Starter pack",
  MISSION_FUEL: "Mission fuel",
  MISSION_LOOT: "Mission loot",
  CRAFT_START: "Craft started",
  CRAFT_OUTPUT: "Craft finished",
  UPGRADE: "Upgrade",
  TRADE_FILL: "Trade",
  AUCTION_SETTLE: "Auction settled",
  WITHDRAWAL: "Withdrawal",
  DEPOSIT: "Deposit",
  EARNED_SETTLE: "Earnings settled",
  REFUND: "Refund",
  DISPUTE_DEBIT: "Chargeback",
  ADMIN_ADJUSTMENT: "Adjustment",
  DEV_CREDIT: "Test credit",
};

export const journalTypeFilterOptions = [
  { value: "all", label: "All movements" },
  ...Object.entries(journalTypeLabels).map(([journalType, label]) => ({ value: journalType, label })),
];

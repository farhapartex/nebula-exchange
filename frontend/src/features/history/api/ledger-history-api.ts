import type { NcBucket } from "@/features/balances/api/balances-api";
import { requestList } from "@/lib/api/api-client";
import type { PaginationParameters } from "@/lib/api/api-types";

export type JournalType =
  | "ENTRY_FEE"
  | "TOPUP_CARD"
  | "TOPUP_CRYPTO"
  | "SHOP_PURCHASE"
  | "STARTER_PACK"
  | "MISSION_FUEL"
  | "MISSION_LOOT"
  | "CRAFT_START"
  | "CRAFT_OUTPUT"
  | "UPGRADE"
  | "TRADE_FILL"
  | "AUCTION_SETTLE"
  | "WITHDRAWAL"
  | "DEPOSIT"
  | "EARNED_SETTLE"
  | "REFUND"
  | "DISPUTE_DEBIT"
  | "ADMIN_ADJUSTMENT"
  | "DEV_CREDIT";

export type LedgerEntry = {
  item_id: number | null;
  bucket: NcBucket | null;
  amount: string;
};

export type LedgerJournal = {
  id: string;
  type: JournalType;
  reference: { type: string; id: string };
  created_at: string;
  entries: LedgerEntry[];
};

export const ledgerHistoryQueryKey = (journalType: JournalType | null) => ["me", "ledger", journalType] as const;

export function listLedgerJournals(journalType: JournalType | null, pagination: PaginationParameters) {
  return requestList<LedgerJournal>("/me/ledger", pagination, { query: { type: journalType ?? undefined } });
}

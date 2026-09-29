package ledger

type JournalType string

const (
	JournalEntryFee        JournalType = "ENTRY_FEE"
	JournalTopupCard       JournalType = "TOPUP_CARD"
	JournalTopupCrypto     JournalType = "TOPUP_CRYPTO"
	JournalShopPurchase    JournalType = "SHOP_PURCHASE"
	JournalStarterPack     JournalType = "STARTER_PACK"
	JournalMissionFuel     JournalType = "MISSION_FUEL"
	JournalMissionLoot     JournalType = "MISSION_LOOT"
	JournalCraftStart      JournalType = "CRAFT_START"
	JournalCraftOutput     JournalType = "CRAFT_OUTPUT"
	JournalUpgrade         JournalType = "UPGRADE"
	JournalTradeFill       JournalType = "TRADE_FILL"
	JournalAuctionSettle   JournalType = "AUCTION_SETTLE"
	JournalWithdrawal      JournalType = "WITHDRAWAL"
	JournalDeposit         JournalType = "DEPOSIT"
	JournalEarnedSettle    JournalType = "EARNED_SETTLE"
	JournalRefund          JournalType = "REFUND"
	JournalDisputeDebit    JournalType = "DISPUTE_DEBIT"
	JournalAdminAdjustment JournalType = "ADMIN_ADJUSTMENT"
	JournalDevCredit       JournalType = "DEV_CREDIT"
)

type Reference struct {
	Type string
	ID   string
}

type Leg struct {
	Account AccountKey
	Amount  int64
}

func Debit(account AccountKey, amount int64) Leg {
	return Leg{Account: account, Amount: -amount}
}

func Credit(account AccountKey, amount int64) Leg {
	return Leg{Account: account, Amount: amount}
}

type Journal struct {
	Type      JournalType
	Reference Reference
	Metadata  map[string]any
	Legs      []Leg
}

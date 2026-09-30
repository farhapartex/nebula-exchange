package ledger

import "github.com/google/uuid"

func MintToPlayer(userID uuid.UUID, itemID int, quantity int64) []Leg {
	return []Leg{
		Debit(System(SystemMint, Item(itemID)), quantity),
		Credit(PlayerItem(userID, itemID), quantity),
	}
}

func BurnFromPlayer(userID uuid.UUID, itemID int, quantity int64) []Leg {
	return []Leg{
		Debit(PlayerItem(userID, itemID), quantity),
		Credit(System(SystemBurn, Item(itemID)), quantity),
	}
}

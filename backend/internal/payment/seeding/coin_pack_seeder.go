package seeding

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"os"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	"github.com/farhapartex/nebula-exchange/backend/internal/payment/models"
)

type CoinPackContent struct {
	ID         string `json:"id"`
	Name       string `json:"name"`
	Coins      int64  `json:"coins"`
	PriceCents int64  `json:"price_cents"`
	SortOrder  int    `json:"sort_order"`
	IsActive   bool   `json:"is_active"`
}

func LoadCoinPacks(filePath string) ([]CoinPackContent, error) {
	coinPacksFile, err := os.ReadFile(filePath)
	if err != nil {
		return nil, err
	}
	var coinPacks []CoinPackContent
	if err := json.Unmarshal(coinPacksFile, &coinPacks); err != nil {
		return nil, fmt.Errorf("parse %s: %w", filePath, err)
	}
	seenSortOrders := map[int]string{}
	for _, coinPack := range coinPacks {
		if coinPack.ID == "" || coinPack.Name == "" {
			return nil, fmt.Errorf("every coin pack needs an id and a name")
		}
		if coinPack.Coins <= 0 || coinPack.PriceCents <= 0 {
			return nil, fmt.Errorf("coin pack %s needs positive coins and price", coinPack.ID)
		}
		if otherCoinPackID, isTaken := seenSortOrders[coinPack.SortOrder]; isTaken {
			return nil, fmt.Errorf("coin packs %s and %s share sort order %d", otherCoinPackID, coinPack.ID, coinPack.SortOrder)
		}
		seenSortOrders[coinPack.SortOrder] = coinPack.ID
	}
	return coinPacks, nil
}

func SeedCoinPacks(ctx context.Context, gormDatabase *gorm.DB, coinPacks []CoinPackContent, logger *slog.Logger) error {
	return gormDatabase.WithContext(ctx).Transaction(func(transaction *gorm.DB) error {
		for _, coinPack := range coinPacks {
			err := transaction.Clauses(clause.OnConflict{
				Columns:   []clause.Column{{Name: "id"}},
				DoUpdates: clause.AssignmentColumns([]string{"name", "coins", "price_cents", "sort_order", "is_active", "updated_at"}),
			}).Create(&models.CoinPack{
				ID:         coinPack.ID,
				Name:       coinPack.Name,
				Coins:      coinPack.Coins,
				PriceCents: coinPack.PriceCents,
				SortOrder:  coinPack.SortOrder,
				IsActive:   coinPack.IsActive,
			}).Error
			if err != nil {
				return fmt.Errorf("save coin pack %s: %w", coinPack.ID, err)
			}
		}
		logger.InfoContext(ctx, "coin packs seeded", slog.Int("coin_packs", len(coinPacks)))
		return nil
	})
}

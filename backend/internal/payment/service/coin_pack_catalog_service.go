package service

import (
	"context"

	"github.com/farhapartex/nebula-exchange/backend/internal/payment/models"
	"github.com/farhapartex/nebula-exchange/backend/internal/payment/repository"
	"github.com/farhapartex/nebula-exchange/backend/internal/platform/pagination"
)

type CoinPackCursor struct {
	SortOrder int `json:"sort_order"`
}

type CoinPackCatalogService interface {
	List(ctx context.Context, after CoinPackCursor, pageRequest pagination.Request) (pagination.Page[models.CoinPack], error)
}

type coinPackCatalogService struct {
	coinPacks repository.CoinPackRepository
}

func NewCoinPackCatalogService(coinPacks repository.CoinPackRepository) CoinPackCatalogService {
	return &coinPackCatalogService{coinPacks: coinPacks}
}

func (catalog *coinPackCatalogService) List(ctx context.Context, after CoinPackCursor, pageRequest pagination.Request) (pagination.Page[models.CoinPack], error) {
	coinPacks, err := catalog.coinPacks.ListActive(ctx, after.SortOrder, pageRequest.FetchLimit())
	if err != nil {
		return pagination.Page[models.CoinPack]{}, err
	}
	return pagination.BuildPage(coinPacks, pageRequest, func(coinPack models.CoinPack) CoinPackCursor {
		return CoinPackCursor{SortOrder: coinPack.SortOrder}
	})
}

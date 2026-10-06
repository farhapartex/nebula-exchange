package repository

import (
	"context"
	"errors"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"github.com/farhapartex/nebula-exchange/backend/internal/platform/database"
	"github.com/farhapartex/nebula-exchange/backend/internal/wallet/models"
)

const (
	walletsUserIDUniqueConstraint  = "wallets_user_id_unique"
	walletsAddressUniqueConstraint = "wallets_address_unique"
)

var (
	ErrWalletNotFound         = errors.New("wallet not found")
	ErrUserAlreadyHasWallet   = errors.New("the user already has a wallet")
	ErrAddressLinkedElsewhere = errors.New("the address is linked to another user")
)

type WalletPosition struct {
	LinkedAt time.Time
	ID       uuid.UUID
}

type WalletRepository interface {
	Create(ctx context.Context, wallet *models.Wallet) error
	FindByUser(ctx context.Context, userID uuid.UUID) (models.Wallet, error)
	ListForUser(ctx context.Context, userID uuid.UUID, after *WalletPosition, limit int) ([]models.Wallet, error)
}

type GormWalletRepository struct {
	database *gorm.DB
}

func NewWalletRepository(database *gorm.DB) *GormWalletRepository {
	return &GormWalletRepository{database: database}
}

func (repository *GormWalletRepository) Create(ctx context.Context, wallet *models.Wallet) error {
	err := database.Session(ctx, repository.database).Create(wallet).Error
	constraintName, isUniqueViolation := database.ViolatedUniqueConstraint(err)
	switch {
	case isUniqueViolation && constraintName == walletsUserIDUniqueConstraint:
		return ErrUserAlreadyHasWallet
	case isUniqueViolation && constraintName == walletsAddressUniqueConstraint:
		return ErrAddressLinkedElsewhere
	}
	return err
}

func (repository *GormWalletRepository) FindByUser(ctx context.Context, userID uuid.UUID) (models.Wallet, error) {
	var wallet models.Wallet
	err := database.Session(ctx, repository.database).Where(map[string]any{"user_id": userID}).Take(&wallet).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return models.Wallet{}, ErrWalletNotFound
	}
	return wallet, err
}

func (repository *GormWalletRepository) ListForUser(ctx context.Context, userID uuid.UUID, after *WalletPosition, limit int) ([]models.Wallet, error) {
	query := database.Session(ctx, repository.database).Where(map[string]any{"user_id": userID})
	if after != nil {
		query = query.Where("(linked_at, id) > (?, ?)", after.LinkedAt, after.ID)
	}
	var wallets []models.Wallet
	err := query.Order("linked_at").Order("id").Limit(limit).Find(&wallets).Error
	return wallets, err
}

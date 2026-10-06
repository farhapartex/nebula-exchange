package wallet

import (
	"time"

	"gorm.io/gorm"

	"github.com/farhapartex/nebula-exchange/backend/internal/platform/database"
	"github.com/farhapartex/nebula-exchange/backend/internal/platform/httpserver"
	"github.com/farhapartex/nebula-exchange/backend/internal/wallet/handler"
	"github.com/farhapartex/nebula-exchange/backend/internal/wallet/repository"
	"github.com/farhapartex/nebula-exchange/backend/internal/wallet/service"
)

type ModuleDependencies struct {
	Database        *gorm.DB
	ChainID         int64
	FrontendBaseURL string
	Now             func() time.Time
}

type Module struct {
	LinkedWallets service.LinkedWalletLookup
	registrars    []httpserver.RouteRegistrar
}

func NewModule(dependencies ModuleDependencies) (*Module, error) {
	wallets := repository.NewWalletRepository(dependencies.Database)
	walletLinkService, err := service.NewWalletLinkService(service.WalletLinkDependencies{
		Challenges:      repository.NewWalletChallengeRepository(dependencies.Database),
		Wallets:         wallets,
		Transactions:    database.NewTransactionRunner(dependencies.Database),
		ChainID:         dependencies.ChainID,
		FrontendBaseURL: dependencies.FrontendBaseURL,
		Now:             service.Clock(dependencies.Now),
	})
	if err != nil {
		return nil, err
	}
	return &Module{
		LinkedWallets: service.NewLinkedWalletLookup(wallets),
		registrars:    []httpserver.RouteRegistrar{handler.NewWalletHandler(walletLinkService)},
	}, nil
}

func (module *Module) RouteRegistrars() []httpserver.RouteRegistrar {
	return module.registrars
}

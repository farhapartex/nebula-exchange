package economy

import (
	"log/slog"
	"time"

	"gorm.io/gorm"

	"github.com/farhapartex/nebula-exchange/backend/internal/economy/repository"
	"github.com/farhapartex/nebula-exchange/backend/internal/economy/service"
	"github.com/farhapartex/nebula-exchange/backend/internal/platform/database"
)

type ModuleDependencies struct {
	Database *gorm.DB
	Logger   *slog.Logger
	Now      func() time.Time
}

type Module struct {
	CoinAccounts service.CoinAccountService
}

func NewModule(dependencies ModuleDependencies) *Module {
	return &Module{CoinAccounts: service.NewCoinAccountService(service.CoinAccountDependencies{
		Balances:      repository.NewCoinBalanceRepository(dependencies.Database),
		Ledger:        repository.NewCoinLedgerRepository(dependencies.Database),
		ClawbackFlags: repository.NewCoinClawbackFlagRepository(dependencies.Database),
		Transactions:  database.NewTransactionRunner(dependencies.Database),
		Logger:        dependencies.Logger,
		Now:           service.Clock(dependencies.Now),
	})}
}

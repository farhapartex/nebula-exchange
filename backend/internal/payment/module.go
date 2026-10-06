package payment

import (
	"log/slog"
	"time"

	"gorm.io/gorm"

	"github.com/farhapartex/nebula-exchange/backend/internal/payment/handler"
	"github.com/farhapartex/nebula-exchange/backend/internal/payment/repository"
	"github.com/farhapartex/nebula-exchange/backend/internal/payment/service"
	"github.com/farhapartex/nebula-exchange/backend/internal/platform/database"
	"github.com/farhapartex/nebula-exchange/backend/internal/platform/httpserver"
)

type ModuleDependencies struct {
	Database          *gorm.DB
	Chapters          service.ChapterCatalog
	ChapterOwnership  service.ChapterOwnership
	ChapterUnlocker   service.ChapterUnlocker
	CheckoutGateway   service.CheckoutGateway
	EventVerifier     service.PaymentEventVerifier
	WalletPayments    *service.WalletPaymentDependencies
	ChainPollInterval time.Duration
	FrontendBaseURL   string
	Logger            *slog.Logger
	Now               func() time.Time
}

type Module struct {
	VaultPaymentWatcher *service.VaultPaymentWatcher
	registrars          []httpserver.RouteRegistrar
}

func NewModule(dependencies ModuleDependencies) *Module {
	plans := repository.NewPlanRepository(dependencies.Database)
	payments := repository.NewPaymentRepository(dependencies.Database)
	transactions := database.NewTransactionRunner(dependencies.Database)
	checkoutService := service.NewCheckoutService(service.CheckoutDependencies{
		Plans:           plans,
		Payments:        payments,
		Chapters:        dependencies.Chapters,
		Ownership:       dependencies.ChapterOwnership,
		Unlocker:        dependencies.ChapterUnlocker,
		Gateway:         dependencies.CheckoutGateway,
		WalletPayments:  dependencies.WalletPayments,
		Transactions:    transactions,
		FrontendBaseURL: dependencies.FrontendBaseURL,
		Logger:          dependencies.Logger,
		Now:             service.Clock(dependencies.Now),
	})
	webhookService := service.NewStripeWebhookService(service.StripeWebhookDependencies{
		Verifier:     dependencies.EventVerifier,
		Events:       repository.NewStripeWebhookEventRepository(dependencies.Database),
		Payments:     payments,
		Unlocker:     dependencies.ChapterUnlocker,
		Transactions: transactions,
		Logger:       dependencies.Logger,
		Now:          service.Clock(dependencies.Now),
	})
	return &Module{VaultPaymentWatcher: newVaultPaymentWatcher(dependencies, payments, transactions), registrars: []httpserver.RouteRegistrar{
		handler.NewStoreHandler(service.NewStoreService(dependencies.Chapters, dependencies.ChapterOwnership, plans)),
		handler.NewCheckoutSessionHandler(checkoutService),
		handler.NewSubscriptionHandler(service.NewSubscriptionService(payments)),
		handler.NewStripeWebhookHandler(webhookService),
	}}
}

func newVaultPaymentWatcher(dependencies ModuleDependencies, payments repository.PaymentRepository, transactions service.TransactionRunner) *service.VaultPaymentWatcher {
	if dependencies.WalletPayments == nil || dependencies.WalletPayments.Chain == nil {
		return nil
	}
	return service.NewVaultPaymentWatcher(service.VaultPaymentWatcherDependencies{
		Chain:                 dependencies.WalletPayments.Chain,
		Cursors:               repository.NewChainSyncCursorRepository(dependencies.Database),
		Payments:              payments,
		Unlocker:              dependencies.ChapterUnlocker,
		Transactions:          transactions,
		RequiredConfirmations: dependencies.WalletPayments.RequiredConfirmations,
		PollInterval:          dependencies.ChainPollInterval,
		Logger:                dependencies.Logger,
		Now:                   service.Clock(dependencies.Now),
	})
}

func (module *Module) RouteRegistrars() []httpserver.RouteRegistrar {
	return module.registrars
}

package progress

import (
	"log/slog"
	"time"

	"gorm.io/gorm"

	"github.com/farhapartex/nebula-exchange/backend/internal/platform/database"
	"github.com/farhapartex/nebula-exchange/backend/internal/platform/httpserver"
	"github.com/farhapartex/nebula-exchange/backend/internal/progress/handler"
	"github.com/farhapartex/nebula-exchange/backend/internal/progress/repository"
	"github.com/farhapartex/nebula-exchange/backend/internal/progress/service"
)

type ModuleDependencies struct {
	Database *gorm.DB
	Levels   service.LevelCatalog
	Logger   *slog.Logger
	Now      func() time.Time
}

type Module struct {
	PlayerProgress   service.PlayerProgressService
	ChapterOwnership service.ChapterOwnershipService
	registrars       []httpserver.RouteRegistrar
}

func NewModule(dependencies ModuleDependencies) *Module {
	levelProgress := repository.NewLevelProgressRepository(dependencies.Database)
	fighters := repository.NewFighterRepository(dependencies.Database)
	fightSessions := repository.NewFightSessionRepository(dependencies.Database)
	transactions := database.NewTransactionRunner(dependencies.Database)
	chapterUnlocks := repository.NewChapterUnlockRepository(dependencies.Database)
	fightSessionService := service.NewFightSessionService(service.FightSessionDependencies{
		Levels:         dependencies.Levels,
		LevelProgress:  levelProgress,
		FightSessions:  fightSessions,
		Fighters:       fighters,
		ChapterUnlocks: chapterUnlocks,
		Transactions:   transactions,
		Logger:         dependencies.Logger,
		Now:            service.Clock(dependencies.Now),
	})
	fightResultService := service.NewFightResultService(service.FightResultDependencies{
		Levels:        dependencies.Levels,
		FightSessions: fightSessions,
		Fighters:      fighters,
		LevelProgress: levelProgress,
		Transactions:  transactions,
		Logger:        dependencies.Logger,
		Now:           service.Clock(dependencies.Now),
	})
	playerProgressService := service.NewPlayerProgressService(service.PlayerProgressDependencies{
		Levels:         dependencies.Levels,
		LevelProgress:  levelProgress,
		FightSessions:  fightSessions,
		Fighters:       fighters,
		ChapterUnlocks: chapterUnlocks,
	})
	return &Module{
		PlayerProgress:   playerProgressService,
		ChapterOwnership: service.NewChapterOwnershipService(chapterUnlocks),
		registrars: []httpserver.RouteRegistrar{
			handler.NewFightSessionHandler(fightSessionService),
			handler.NewFightSetupHandler(service.NewFightSetupService(dependencies.Levels, fighters)),
			handler.NewFightResultHandler(fightResultService),
			handler.NewNextLevelHandler(playerProgressService),
		},
	}
}

func (module *Module) RouteRegistrars() []httpserver.RouteRegistrar {
	return module.registrars
}

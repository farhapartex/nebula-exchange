package billing

import (
	"gorm.io/gorm"

	"github.com/farhapartex/nebula-exchange/backend/internal/billing/handler"
	"github.com/farhapartex/nebula-exchange/backend/internal/billing/repository"
	"github.com/farhapartex/nebula-exchange/backend/internal/billing/service"
	"github.com/farhapartex/nebula-exchange/backend/internal/platform/httpserver"
)

type ModuleDependencies struct {
	Database         *gorm.DB
	Chapters         service.ChapterCatalog
	ChapterOwnership service.ChapterOwnership
}

type Module struct {
	registrars []httpserver.RouteRegistrar
}

func NewModule(dependencies ModuleDependencies) *Module {
	storeService := service.NewStoreService(dependencies.Chapters, dependencies.ChapterOwnership, repository.NewPlanRepository(dependencies.Database))
	return &Module{registrars: []httpserver.RouteRegistrar{handler.NewStoreHandler(storeService)}}
}

func (module *Module) RouteRegistrars() []httpserver.RouteRegistrar {
	return module.registrars
}

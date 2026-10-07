package tools

import (
	"gorm.io/gorm"

	"github.com/farhapartex/nebula-exchange/backend/internal/platform/httpserver"
	"github.com/farhapartex/nebula-exchange/backend/internal/tools/handler"
	"github.com/farhapartex/nebula-exchange/backend/internal/tools/repository"
	"github.com/farhapartex/nebula-exchange/backend/internal/tools/service"
)

type ModuleDependencies struct {
	Database       *gorm.DB
	PlayerStanding service.PlayerStanding
	Levels         service.LevelDirectory
}

type Module struct {
	registrars []httpserver.RouteRegistrar
}

func NewModule(dependencies ModuleDependencies) *Module {
	shopItemService := service.NewShopItemService(
		repository.NewToolTypeRepository(dependencies.Database),
		repository.NewToolRepository(dependencies.Database),
		dependencies.PlayerStanding,
		dependencies.Levels,
	)
	return &Module{registrars: []httpserver.RouteRegistrar{handler.NewShopItemHandler(shopItemService)}}
}

func (module *Module) RouteRegistrars() []httpserver.RouteRegistrar {
	return module.registrars
}

package story

import (
	"gorm.io/gorm"

	"github.com/farhapartex/nebula-exchange/backend/internal/platform/httpserver"
	"github.com/farhapartex/nebula-exchange/backend/internal/story/handler"
	"github.com/farhapartex/nebula-exchange/backend/internal/story/repository"
	"github.com/farhapartex/nebula-exchange/backend/internal/story/service"
)

type ModuleDependencies struct {
	Database    *gorm.DB
	ImageSigner service.ImageURLSigner
}

type Module struct {
	LevelCatalog service.LevelCatalog
	registrars   []httpserver.RouteRegistrar
}

func NewModule(dependencies ModuleDependencies) *Module {
	levels := repository.NewLevelRepository(dependencies.Database)
	storyService := service.NewStoryService(levels, repository.NewStorySlideRepository(dependencies.Database), dependencies.ImageSigner)
	return &Module{
		LevelCatalog: service.NewLevelCatalog(levels),
		registrars:   []httpserver.RouteRegistrar{handler.NewStoryHandler(storyService)},
	}
}

func (module *Module) RouteRegistrars() []httpserver.RouteRegistrar {
	return module.registrars
}

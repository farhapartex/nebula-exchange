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
	registrars []httpserver.RouteRegistrar
}

func NewModule(dependencies ModuleDependencies) *Module {
	storyService := service.NewStoryService(
		repository.NewLevelRepository(dependencies.Database),
		repository.NewStorySlideRepository(dependencies.Database),
		dependencies.ImageSigner,
	)
	return &Module{registrars: []httpserver.RouteRegistrar{handler.NewStoryHandler(storyService)}}
}

func (module *Module) RouteRegistrars() []httpserver.RouteRegistrar {
	return module.registrars
}

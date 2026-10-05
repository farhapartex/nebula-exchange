package httpserver

import (
	"log/slog"

	"github.com/gin-gonic/gin"

	"github.com/farhapartex/nebula-exchange/backend/internal/platform/apierror"
	"github.com/farhapartex/nebula-exchange/backend/internal/platform/httpserver/middleware"
	"github.com/farhapartex/nebula-exchange/backend/internal/platform/httpserver/request"
	"github.com/farhapartex/nebula-exchange/backend/internal/platform/httpserver/response"
)

const (
	APIBasePath       = "/api/v1"
	webhookPathPrefix = APIBasePath + "/webhooks/"
)

type RouteRegistrar interface {
	RegisterRoutes(router gin.IRouter)
}

type RouterOptions struct {
	Logger         *slog.Logger
	IsProduction   bool
	AllowedOrigins []string
	TrustedProxies []string
}

func NewRouter(options RouterOptions, registrars ...RouteRegistrar) (*gin.Engine, error) {
	if options.IsProduction {
		gin.SetMode(gin.ReleaseMode)
	}
	request.RegisterJSONFieldNames()

	router := gin.New()
	if err := router.SetTrustedProxies(options.TrustedProxies); err != nil {
		return nil, err
	}
	router.HandleMethodNotAllowed = true
	router.NoRoute(func(context *gin.Context) {
		response.WriteError(context, apierror.NotFound("Route not found"))
	})
	router.NoMethod(func(context *gin.Context) {
		response.WriteError(context, apierror.MethodNotAllowed())
	})

	router.Use(
		middleware.RequestID(),
		middleware.RequestLogger(options.Logger),
		middleware.PanicRecovery(options.Logger),
		middleware.SecurityHeaders(options.IsProduction),
	)
	if len(options.AllowedOrigins) > 0 {
		router.Use(middleware.CrossOrigin(options.AllowedOrigins))
	}
	router.Use(middleware.RequireClientIdentification(webhookPathPrefix))

	apiGroup := router.Group(APIBasePath)
	for _, registrar := range registrars {
		registrar.RegisterRoutes(apiGroup)
	}
	return router, nil
}

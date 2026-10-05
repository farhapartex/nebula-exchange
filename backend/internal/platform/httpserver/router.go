package httpserver

import (
	"log/slog"

	"github.com/gin-gonic/gin"

	"github.com/farhapartex/nebula-exchange/backend/internal/platform/apierror"
	"github.com/farhapartex/nebula-exchange/backend/internal/platform/httpserver/authentication"
	"github.com/farhapartex/nebula-exchange/backend/internal/platform/httpserver/middleware"
	"github.com/farhapartex/nebula-exchange/backend/internal/platform/httpserver/request"
	"github.com/farhapartex/nebula-exchange/backend/internal/platform/httpserver/response"
	"github.com/farhapartex/nebula-exchange/backend/internal/platform/idempotency"
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
	AccessTokens   authentication.AccessTokenVerifier
	Idempotency    idempotency.Store
	NonReplayable  []string
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
		router.Use(middleware.CrossOrigin(options.AllowedOrigins, idempotency.ReplayedHeader))
	}
	router.Use(middleware.RequireClientIdentification(webhookPathPrefix))
	if options.AccessTokens != nil {
		router.Use(authentication.IdentifyUser(options.AccessTokens))
	}
	if options.Idempotency != nil {
		router.Use(idempotency.Middleware(options.Idempotency, options.Logger, idempotency.Options{ExemptRoutes: withAPIBasePath(options.NonReplayable)}))
	}

	apiGroup := router.Group(APIBasePath)
	for _, registrar := range registrars {
		registrar.RegisterRoutes(apiGroup)
	}
	return router, nil
}

func withAPIBasePath(routePaths []string) []string {
	fullPaths := make([]string, len(routePaths))
	for pathIndex, routePath := range routePaths {
		fullPaths[pathIndex] = APIBasePath + routePath
	}
	return fullPaths
}

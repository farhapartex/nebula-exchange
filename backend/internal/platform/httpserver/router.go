package httpserver

import (
	"log/slog"
	"net/http"

	"github.com/gin-gonic/gin"

	"nebula-exchange/backend/internal/platform/apierror"
	"nebula-exchange/backend/internal/platform/httpserver/middleware"
	"nebula-exchange/backend/internal/platform/httpserver/request"
	"nebula-exchange/backend/internal/platform/httpserver/response"
	"nebula-exchange/backend/internal/platform/idempotency"
)

const (
	apiBasePath       = "/api/v1"
	webhookPathPrefix = apiBasePath + "/webhooks/"
)

type RouteRegistrar interface {
	RegisterRoutes(router gin.IRouter)
}

type RouterOptions struct {
	Logger           *slog.Logger
	IsProduction     bool
	AllowedOrigins   []string
	TrustedProxies   []string
	IdentifyUser     gin.HandlerFunc
	IdempotencyStore idempotency.Store
}

func NewRouter(options RouterOptions, registrars ...RouteRegistrar) *gin.Engine {
	if options.IsProduction {
		gin.SetMode(gin.ReleaseMode)
	}
	request.RegisterJSONFieldNames()

	router := gin.New()
	if err := router.SetTrustedProxies(options.TrustedProxies); err != nil {
		panic(err)
	}
	router.HandleMethodNotAllowed = true
	router.NoRoute(func(context *gin.Context) {
		response.WriteError(context, apierror.NotFound("Route not found"))
	})
	router.NoMethod(func(context *gin.Context) {
		response.WriteError(context, apierror.New(http.StatusMethodNotAllowed, apierror.CodeMethodNotAllowed, "Method not allowed"))
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
	if options.IdentifyUser != nil {
		router.Use(options.IdentifyUser)
	}
	if options.IdempotencyStore != nil {
		router.Use(idempotency.Middleware(options.IdempotencyStore, options.Logger, idempotency.Options{}))
	}

	apiGroup := router.Group(apiBasePath)
	for _, registrar := range registrars {
		registrar.RegisterRoutes(apiGroup)
	}
	return router
}

package httpserver

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"

	"nebula-exchange/backend/internal/platform/httpserver/middleware"
)

const apiBasePath = "/api/v1"

type RouteRegistrar interface {
	RegisterRoutes(router gin.IRouter)
}

type Server struct {
	httpServer *http.Server
	logger     *slog.Logger
}

func NewRouter(logger *slog.Logger, isProduction bool, registrars ...RouteRegistrar) *gin.Engine {
	if isProduction {
		gin.SetMode(gin.ReleaseMode)
	}
	router := gin.New()
	router.Use(
		middleware.RequestID(),
		middleware.RequestLogger(logger),
		middleware.PanicRecovery(logger),
	)

	apiGroup := router.Group(apiBasePath)
	for _, registrar := range registrars {
		registrar.RegisterRoutes(apiGroup)
	}
	return router
}

func New(address string, handler http.Handler, logger *slog.Logger) *Server {
	return &Server{
		httpServer: &http.Server{
			Addr:              address,
			Handler:           handler,
			ReadHeaderTimeout: 10 * time.Second,
			ReadTimeout:       30 * time.Second,
			WriteTimeout:      30 * time.Second,
			IdleTimeout:       120 * time.Second,
		},
		logger: logger,
	}
}

func (server *Server) Run(shutdownSignal context.Context, shutdownTimeout time.Duration) error {
	listenFailure := make(chan error, 1)
	go func() {
		server.logger.Info("http server listening", slog.String("address", server.httpServer.Addr))
		if err := server.httpServer.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			listenFailure <- err
		}
		close(listenFailure)
	}()

	select {
	case err := <-listenFailure:
		return err
	case <-shutdownSignal.Done():
	}

	server.logger.Info("http server shutting down", slog.Duration("timeout", shutdownTimeout))
	shutdownContext, cancelShutdown := context.WithTimeout(context.Background(), shutdownTimeout)
	defer cancelShutdown()
	if err := server.httpServer.Shutdown(shutdownContext); err != nil {
		return err
	}
	server.logger.Info("http server stopped")
	return nil
}

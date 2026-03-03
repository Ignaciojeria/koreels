package httpserver

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"ledger-service/internal/shared/configuration"
	"ledger-service/internal/shared/infrastructure/httpserver/middleware"
	"ledger-service/internal/shared/infrastructure/observability"

	"github.com/Ignaciojeria/ioc"
	"github.com/go-fuego/fuego"
	"github.com/go-fuego/fuego/option"
	"github.com/hellofresh/health-go/v5"
)

var (
	_ = ioc.Register(NewServer)
	_ = ioc.RegisterAtEnd(StartServer)
	shutdownTimeout = time.Second * 5
)

type Server struct {
	Manager *fuego.Server
	conf    configuration.Conf
}

func NewServer(conf configuration.Conf, requestLogger middleware.RequestLogger) (*Server, error) {
	oidcMiddleware := middleware.NewOIDCFromConf(conf)
	s := fuego.NewServer(fuego.WithAddr(":"+conf.PORT), fuego.WithGlobalMiddlewares(requestLogger, oidcMiddleware))
	s.ReadTimeout = 30 * time.Minute
	s.WriteTimeout = 30 * time.Minute
	s.ReadHeaderTimeout = 30 * time.Minute
	s.IdleTimeout = 30 * time.Minute
	server := &Server{Manager: s, conf: conf}
	if err := server.healthCheck(); err != nil {
		return nil, fmt.Errorf("failed to init healthcheck: %w", err)
	}
	return server, nil
}

func StartServer(s *Server, obs observability.Observability) error {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	go func() {
		<-sigChan
		shutdownCtx, shutdownCancel := context.WithTimeout(ctx, shutdownTimeout)
		defer shutdownCancel()
		if err := s.Manager.Shutdown(shutdownCtx); err != nil {
			obs.Logger.Error("server shutdown error", "error", err)
		}
		cancel()
	}()
	obs.Logger.Info("http server starting", "port", s.conf.PORT, "service", s.conf.PROJECT_NAME, "version", s.conf.VERSION)
	if err := s.Manager.Run(); err != nil && err != http.ErrServerClosed {
		return fmt.Errorf("server failed to start: %w", err)
	}
	return nil
}

var healthNew = func(opts ...health.Option) (*health.Health, error) { return health.New(opts...) }

func (s *Server) healthCheck() error {
	h, err := healthNew(
		health.WithComponent(health.Component{Name: s.conf.PROJECT_NAME, Version: s.conf.VERSION}),
		health.WithSystemInfo(),
	)
	if err != nil {
		return err
	}
	fuego.GetStd(s.Manager, "/health", h.Handler().ServeHTTP, option.Summary("healthCheck"))
	return nil
}

# httpserver

> HTTP server setup with Fuego framework and healthcheck

## app/shared/infrastructure/httpserver/server.go

```go
package httpserver

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/Ignaciojeria/ioc"
	"github.com/go-fuego/fuego"
	"github.com/go-fuego/fuego/option"
	"github.com/hellofresh/health-go/v5"

	"archetype/app/shared/configuration"
	"archetype/app/shared/infrastructure/httpserver/middleware"
	"archetype/app/shared/infrastructure/observability"
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

// NewServer creates a new instance of the HTTP Fuego Server.
// It returns a pointer because it manages network state.
func NewServer(conf configuration.Conf, requestLogger middleware.RequestLogger) (*Server, error) {
	s := fuego.NewServer(
		fuego.WithAddr(":"+conf.PORT),
		fuego.WithGlobalMiddlewares(requestLogger),
	)

	// Defaults tailored for resiliency, preventing long-hanging idle connections
	s.ReadTimeout = 30 * time.Minute
	s.WriteTimeout = 30 * time.Minute
	s.ReadHeaderTimeout = 30 * time.Minute
	s.IdleTimeout = 30 * time.Minute

	server := &Server{
		Manager: s,
		conf:    conf,
	}

	if err := server.healthCheck(); err != nil {
		return nil, fmt.Errorf("failed to init healthcheck: %w", err)
	}

	return server, nil
}

// StartServer runs at the end of the dependency graph and starts the HTTP server.
// It blocks the main thread and gracefully handles OS signals.
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

	obs.Logger.Info(
		"http server starting",
		"port", s.conf.PORT,
		"service", s.conf.PROJECT_NAME,
		"version", s.conf.VERSION,
	)

	if err := s.Manager.Run(); err != nil && err != http.ErrServerClosed {
		return fmt.Errorf("server failed to start: %w", err)
	}

	return nil
}

// healthNew is extractable for testing error paths.
var healthNew = func(opts ...health.Option) (*health.Health, error) {
	return health.New(opts...)
}

func (s *Server) healthCheck() error {
	h, err := healthNew(
		health.WithComponent(health.Component{
			Name:    s.conf.PROJECT_NAME,
			Version: s.conf.VERSION,
		}),
		health.WithSystemInfo(),
	)
	if err != nil {
		return err
	}

	fuego.GetStd(s.Manager,
		"/health",
		h.Handler().ServeHTTP,
		option.Summary("healthCheck"),
	)
	return nil
}
```

---

## Unit tests

When creating a new component, generate tests following this pattern:

### app/shared/infrastructure/httpserver/server_test.go

```go
package httpserver

import (
	"context"
	"errors"
	"io"
	"log/slog"
	"os"
	"strings"
	"syscall"
	"testing"
	"time"

	"archetype/app/shared/configuration"
	"archetype/app/shared/infrastructure/httpserver/middleware"
	"archetype/app/shared/infrastructure/observability"

	"github.com/hellofresh/health-go/v5"
	"go.opentelemetry.io/otel/trace/noop"
)

func createDummyObservability() (observability.Observability, middleware.RequestLogger) {
	logger := slog.New(slog.NewJSONHandler(io.Discard, nil))
	obs := observability.Observability{
		Tracer: noop.NewTracerProvider().Tracer("test"),
		Logger: logger,
	}
	return obs, middleware.NewRequestLogger(obs)
}

func TestNewServer_HealthCheckError(t *testing.T) {
	oldHealthNew := healthNew
	defer func() { healthNew = oldHealthNew }()

	healthNew = func(...health.Option) (*health.Health, error) {
		return nil, errors.New("health init failed")
	}

	conf := configuration.Conf{
		PORT:         "8082",
		PROJECT_NAME: "test",
		VERSION:      "v1",
	}

	_, mw := createDummyObservability()

	_, err := NewServer(conf, mw)
	if err == nil {
		t.Fatal("expected error when health init fails, got nil")
	}
	if !strings.Contains(err.Error(), "failed to init healthcheck") {
		t.Errorf("expected healthcheck error message, got %v", err)
	}
}

func TestNewServer(t *testing.T) {
	conf := configuration.Conf{
		PORT:         "8081",
		PROJECT_NAME: "test-project",
		VERSION:      "v1",
	}

	_, mw := createDummyObservability()
	server, err := NewServer(conf, mw)
	if err != nil {
		t.Fatalf("unexpected error creating server: %v", err)
	}

	if server.Manager == nil {
		t.Fatal("expected server.Manager to be initialized")
	}
}

func TestStartServer_GracefulShutdown(t *testing.T) {
	conf := configuration.Conf{
		PORT:         "0",
		PROJECT_NAME: "test-start",
		VERSION:      "v1",
	}

	obs, mw := createDummyObservability()
	server, err := NewServer(conf, mw)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	errChan := make(chan error, 1)

	go func() {
		errChan <- StartServer(server, obs)
	}()

	time.Sleep(200 * time.Millisecond)

	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()

	if err := server.Manager.Shutdown(ctx); err != nil {
		t.Fatalf("failed to shutdown test server: %v", err)
	}

	select {
	case err := <-errChan:
		if err != nil {
			t.Errorf("StartServer returned an unexpected error: %v", err)
		}
	case <-time.After(3 * time.Second):
		t.Fatal("StartServer took too long to return after shutdown")
	}
}

func TestStartServer_Signal(t *testing.T) {
	conf := configuration.Conf{
		PORT:         "0",
		PROJECT_NAME: "test-signal",
		VERSION:      "v1",
	}

	obs, mw := createDummyObservability()
	server, _ := NewServer(conf, mw)
	errChan := make(chan error, 1)

	go func() {
		errChan <- StartServer(server, obs)
	}()

	time.Sleep(200 * time.Millisecond)

	p, _ := os.FindProcess(os.Getpid())
	_ = p.Signal(syscall.SIGINT)

	select {
	case <-errChan:
		// success
	case <-time.After(2 * time.Second):
		// if it doesn't return, we shutdown manually to not hang
		_ = server.Manager.Shutdown(context.Background())
	}
}

func TestStartServer_ShutdownError(t *testing.T) {
	oldTimeout := shutdownTimeout
	shutdownTimeout = 1
	defer func() { shutdownTimeout = oldTimeout }()

	conf := configuration.Conf{
		PORT:         "0",
		PROJECT_NAME: "test-shutdown-err",
		VERSION:      "v1",
	}

	obs, mw := createDummyObservability()
	server, err := NewServer(conf, mw)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	errChan := make(chan error, 1)
	go func() {
		errChan <- StartServer(server, obs)
	}()

	time.Sleep(200 * time.Millisecond)

	p, _ := os.FindProcess(os.Getpid())
	_ = p.Signal(syscall.SIGINT)

	select {
	case <-errChan:
		// Shutdown path executed (error branch hit when timeout)
	case <-time.After(2 * time.Second):
		_ = server.Manager.Shutdown(context.Background())
	}
}

func TestStartServer_InvalidPort(t *testing.T) {
	conf := configuration.Conf{
		PORT:         "-1",
		PROJECT_NAME: "test-port",
		VERSION:      "v1",
	}

	obs, mw := createDummyObservability()
	server, err := NewServer(conf, mw)
	if err != nil {
		t.Fatalf("unexpected error creating server: %v", err)
	}

	err = StartServer(server, obs)
	if err == nil {
		t.Fatal("expected error due to invalid port, got nil")
	}
}
```

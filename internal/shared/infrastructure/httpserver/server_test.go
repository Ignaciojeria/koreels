package httpserver

import (
	"context"
	"io"
	"log/slog"
	"testing"
	"time"

	"koreels/internal/shared/configuration"
	"koreels/internal/shared/infrastructure/httpserver/middleware"
	"koreels/internal/shared/infrastructure/observability"

	"go.opentelemetry.io/otel/trace/noop"
)

func createDummyObservability() (observability.Observability, middleware.RequestLogger) {
	logger := slog.New(slog.NewJSONHandler(io.Discard, nil))
	obs := observability.Observability{Tracer: noop.NewTracerProvider().Tracer("test"), Logger: logger}
	return obs, middleware.NewRequestLogger(obs)
}

func TestNewServer(t *testing.T) {
	conf := configuration.Conf{PORT: "8081", PROJECT_NAME: "koreels", VERSION: "v1"}
	_, mw := createDummyObservability()
	server, err := NewServer(conf, mw)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if server.Manager == nil {
		t.Fatal("expected server.Manager to be initialized")
	}
}

func TestStartServer_GracefulShutdown(t *testing.T) {
	conf := configuration.Conf{PORT: "0", PROJECT_NAME: "test", VERSION: "v1"}
	obs, mw := createDummyObservability()
	server, err := NewServer(conf, mw)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	errChan := make(chan error, 1)
	go func() { errChan <- StartServer(server, obs) }()
	time.Sleep(200 * time.Millisecond)
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	defer cancel()
	_ = server.Manager.Shutdown(ctx)
	select {
	case <-errChan:
	case <-time.After(3 * time.Second):
		t.Fatal("timeout")
	}
}

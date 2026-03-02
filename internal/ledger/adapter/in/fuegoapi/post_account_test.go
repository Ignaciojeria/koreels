package fuegoapi

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"

	"ledger-service/internal/ledger/application/ports/in"
	"ledger-service/internal/shared/configuration"
	"ledger-service/internal/shared/infrastructure/httpserver"
	"ledger-service/internal/shared/infrastructure/httpserver/middleware"
	"ledger-service/internal/shared/infrastructure/observability"

	"go.opentelemetry.io/otel/trace/noop"
)

func createTestServer(t *testing.T) *httpserver.Server {
	t.Helper()
	obs := observability.Observability{
		Tracer: noop.NewTracerProvider().Tracer("test"),
		Logger: slog.New(slog.NewJSONHandler(io.Discard, nil)),
	}
	mw := middleware.NewRequestLogger(obs)
	conf := configuration.Conf{PORT: "8091", PROJECT_NAME: "test", VERSION: "v1"}
	server, err := httpserver.NewServer(conf, mw)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	return server
}

type mockCreateAccount struct {
	out in.CreateAccountOutput
	err error
}

func (m *mockCreateAccount) Execute(ctx context.Context, req in.CreateAccountInput) (in.CreateAccountOutput, error) {
	if m.err != nil {
		return in.CreateAccountOutput{}, m.err
	}
	return m.out, nil
}

func TestNewPostAccount(t *testing.T) {
	server := createTestServer(t)
	var uc in.CreateAccountExecutor = &mockCreateAccount{
		out: in.CreateAccountOutput{
			AccountId: "550e8400-e29b-41d4-a716-446655440000",
			Balance:   0,
			Currency:  "CREDITS",
			CreatedAt: "2026-03-02T00:00:00Z",
		},
	}
	NewPostAccount(server, uc)

	reqBody, _ := json.Marshal(CreateAccountRequest{
		AccountId:     "550e8400-e29b-41d4-a716-446655440000",
		Type:          "USER",
		Currency:      "CREDITS",
		AllowNegative: false,
	})
	req, _ := http.NewRequest(http.MethodPost, "/accounts", bytes.NewBuffer(reqBody))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	server.Manager.Mux.ServeHTTP(rec, req)

	if rec.Code != http.StatusCreated {
		t.Errorf("got status %v want %v", rec.Code, http.StatusCreated)
	}
}

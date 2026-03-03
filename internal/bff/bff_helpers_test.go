package bff

import (
	"context"
	"io"
	"log/slog"
	"net/http"

	"koreels/internal/ledger/application/ports/in"
	"koreels/internal/shared/configuration"
	"koreels/internal/shared/contextkeys"
	"koreels/internal/shared/infrastructure/httpserver"
	"koreels/internal/shared/infrastructure/httpserver/middleware"
	"koreels/internal/shared/infrastructure/observability"

	"go.opentelemetry.io/otel/trace/noop"
)

type mockGetAccount struct {
	out in.GetAccountOutput
	err error
}

func (m *mockGetAccount) Execute(ctx context.Context, accountId string) (in.GetAccountOutput, error) {
	if m.err != nil {
		return in.GetAccountOutput{}, m.err
	}
	return m.out, nil
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

// createBFFTestServer creates an HTTP server with request logger (no OIDC) for testing.
func createBFFTestServer(t interface{ Fatal(...any) }) *httpserver.Server {
	obs := observability.Observability{
		Tracer: noop.NewTracerProvider().Tracer("test"),
		Logger: slog.New(slog.NewJSONHandler(io.Discard, nil)),
	}
	mw := middleware.NewRequestLogger(obs)
	conf := configuration.Conf{PORT: "8092", PROJECT_NAME: "test-bff", VERSION: "v1"}
	server, err := httpserver.NewServer(conf, mw)
	if err != nil {
		t.Fatal("unexpected error:", err)
	}
	return server
}

// requestWithIdentity returns a new request with identity injected into the context.
func requestWithIdentity(r *http.Request, subject, accountID string) *http.Request {
	id := contextkeys.Identity{Subject: subject, AccountID: accountID}
	if accountID == "" {
		id.AccountID = subject
	}
	return r.WithContext(contextkeys.WithIdentity(r.Context(), id))
}

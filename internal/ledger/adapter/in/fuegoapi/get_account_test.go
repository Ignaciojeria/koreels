package fuegoapi

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"koreels/internal/ledger/application/ports/in"
	domainerrors "koreels/internal/ledger/domain/errors"
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

func TestNewGetAccount(t *testing.T) {
	server := createTestServer(t)
	var uc in.GetAccountExecutor = &mockGetAccount{
		out: in.GetAccountOutput{
			AccountId:        "550e8400-e29b-41d4-a716-446655440000",
			Balance:          1200,
			AvailableBalance: 1200,
			Currency:         "CREDITS",
			UpdatedAt:        "2026-03-02T00:00:00Z",
		},
	}
	NewGetAccount(server, uc)

	req, _ := http.NewRequest(http.MethodGet, "/accounts/550e8400-e29b-41d4-a716-446655440000", nil)
	rec := httptest.NewRecorder()
	server.Manager.Mux.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("got status %v want %v", rec.Code, http.StatusOK)
	}
}

func TestNewGetAccount_NotFound_Returns404(t *testing.T) {
	server := createTestServer(t)
	var uc in.GetAccountExecutor = &mockGetAccount{err: domainerrors.ErrAccountNotFound}
	NewGetAccount(server, uc)

	req, _ := http.NewRequest(http.MethodGet, "/accounts/550e8400-e29b-41d4-a716-446655440000", nil)
	rec := httptest.NewRecorder()
	server.Manager.Mux.ServeHTTP(rec, req)

	if rec.Code != http.StatusNotFound {
		t.Errorf("got status %v want %v", rec.Code, http.StatusNotFound)
	}
}

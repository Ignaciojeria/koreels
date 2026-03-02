package fuegoapi

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"ledger-service/internal/ledger/application/ports/in"
)

type mockGetAccountLedger struct {
	out in.GetAccountLedgerOutput
	err error
}

func (m *mockGetAccountLedger) Execute(ctx context.Context, accountId string, limit int, cursor string) (in.GetAccountLedgerOutput, error) {
	if m.err != nil {
		return in.GetAccountLedgerOutput{}, m.err
	}
	return m.out, nil
}

func TestNewGetAccountLedger(t *testing.T) {
	server := createTestServer(t)
	var uc in.GetAccountLedgerExecutor = &mockGetAccountLedger{out: in.GetAccountLedgerOutput{AccountId: "550e8400-e29b-41d4-a716-446655440000"}}
	NewGetAccountLedger(server, uc)

	req, _ := http.NewRequest(http.MethodGet, "/accounts/550e8400-e29b-41d4-a716-446655440000/ledger", nil)
	rec := httptest.NewRecorder()
	server.Manager.Mux.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("got status %v want %v", rec.Code, http.StatusOK)
	}
}

package fuegoapi

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"koreels/internal/ledger/application/ports/in"
	domainerrors "koreels/internal/ledger/domain/errors"
)

type mockGetTransaction struct {
	out in.GetTransactionOutput
	err error
}

func (m *mockGetTransaction) Execute(ctx context.Context, transactionId string) (in.GetTransactionOutput, error) {
	if m.err != nil {
		return in.GetTransactionOutput{}, m.err
	}
	return m.out, nil
}

func TestNewGetTransaction(t *testing.T) {
	server := createTestServer(t)
	var uc in.GetTransactionExecutor = &mockGetTransaction{
		out: in.GetTransactionOutput{
			TransactionId: "660e8400-e29b-41d4-a716-446655440001",
			Entries:       []in.TransactionEntryOutput{{AccountId: "a", Amount: -100, BalanceAfter: 900}},
		},
	}
	NewGetTransaction(server, uc)

	req, _ := http.NewRequest(http.MethodGet, "/transactions/660e8400-e29b-41d4-a716-446655440001", nil)
	rec := httptest.NewRecorder()
	server.Manager.Mux.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("got status %v want %v", rec.Code, http.StatusOK)
	}
}

func TestNewGetTransaction_NotFound_Returns404(t *testing.T) {
	server := createTestServer(t)
	var uc in.GetTransactionExecutor = &mockGetTransaction{err: domainerrors.ErrTransactionNotFound}
	NewGetTransaction(server, uc)

	req, _ := http.NewRequest(http.MethodGet, "/transactions/660e8400-e29b-41d4-a716-446655440001", nil)
	rec := httptest.NewRecorder()
	server.Manager.Mux.ServeHTTP(rec, req)

	if rec.Code != http.StatusNotFound {
		t.Errorf("got status %v want %v", rec.Code, http.StatusNotFound)
	}
}

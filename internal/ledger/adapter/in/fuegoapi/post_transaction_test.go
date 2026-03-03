package fuegoapi

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"koreels/internal/ledger/application/ports/in"
)

type mockCreateTransaction struct {
	out in.CreateTransactionOutput
	err error
}

func (m *mockCreateTransaction) Execute(ctx context.Context, req in.CreateTransactionInput) (in.CreateTransactionOutput, error) {
	if m.err != nil {
		return in.CreateTransactionOutput{}, m.err
	}
	return m.out, nil
}

func TestNewPostTransaction(t *testing.T) {
	server := createTestServer(t)
	var uc in.CreateTransactionExecutor = &mockCreateTransaction{
		out: in.CreateTransactionOutput{
			TransactionId: "660e8400-e29b-41d4-a716-446655440001",
			Status:        "COMMITTED",
			CreatedAt:     "2026-03-02T00:00:00Z",
		},
	}
	NewPostTransaction(server, uc)

	reqBody, _ := json.Marshal(CreateTransactionRequest{
		TransactionId: "660e8400-e29b-41d4-a716-446655440001",
		Entries: []TransactionEntryInputReq{
			{AccountId: "550e8400-e29b-41d4-a716-446655440000", Amount: -100},
			{AccountId: "550e8400-e29b-41d4-a716-446655440001", Amount: 100},
		},
	})
	req, _ := http.NewRequest(http.MethodPost, "/transactions", bytes.NewBuffer(reqBody))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Idempotency-Key", "test-key")
	rec := httptest.NewRecorder()
	server.Manager.Mux.ServeHTTP(rec, req)

	if rec.Code != http.StatusCreated {
		t.Errorf("got status %v want %v", rec.Code, http.StatusCreated)
	}
}

func TestNewPostTransaction_MissingIdempotencyKey_Returns400(t *testing.T) {
	server := createTestServer(t)
	var uc in.CreateTransactionExecutor = &mockCreateTransaction{}
	NewPostTransaction(server, uc)

	reqBody, _ := json.Marshal(CreateTransactionRequest{
		TransactionId: "660e8400-e29b-41d4-a716-446655440001",
		Entries: []TransactionEntryInputReq{
			{AccountId: "550e8400-e29b-41d4-a716-446655440000", Amount: -100},
			{AccountId: "550e8400-e29b-41d4-a716-446655440001", Amount: 100},
		},
	})
	req, _ := http.NewRequest(http.MethodPost, "/transactions", bytes.NewBuffer(reqBody))
	req.Header.Set("Content-Type", "application/json")
	// No Idempotency-Key header
	rec := httptest.NewRecorder()
	server.Manager.Mux.ServeHTTP(rec, req)

	if rec.Code != http.StatusBadRequest {
		t.Errorf("got status %v want %v", rec.Code, http.StatusBadRequest)
	}
}

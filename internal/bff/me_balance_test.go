package bff

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"ledger-service/internal/ledger/application/ports/in"
)

func TestNewMeBalanceRoutes_WithIdentity_Returns200(t *testing.T) {
	server := createBFFTestServer(t)
	uc := &mockGetAccount{
		out: in.GetAccountOutput{
			AccountId:        "user-123",
			Balance:          1000,
			AvailableBalance: 1000,
			Currency:         "CREDITS",
			UpdatedAt:        "2026-03-02T00:00:00Z",
		},
	}
	_, err := NewMeBalanceRoutes(server, uc)
	if err != nil {
		t.Fatal(err)
	}

	req, _ := http.NewRequest(http.MethodGet, "/me/balance", nil)
	req = requestWithIdentity(req, "sub-123", "user-123")
	rec := httptest.NewRecorder()
	server.Manager.Mux.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Errorf("got status %v want %v", rec.Code, http.StatusOK)
	}
}

func TestNewMeBalanceRoutes_WithoutIdentity_Returns401(t *testing.T) {
	server := createBFFTestServer(t)
	uc := &mockGetAccount{}
	_, err := NewMeBalanceRoutes(server, uc)
	if err != nil {
		t.Fatal(err)
	}

	req, _ := http.NewRequest(http.MethodGet, "/me/balance", nil)
	rec := httptest.NewRecorder()
	server.Manager.Mux.ServeHTTP(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Errorf("got status %v want %v", rec.Code, http.StatusUnauthorized)
	}
}

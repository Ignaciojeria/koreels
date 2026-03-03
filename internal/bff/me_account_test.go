package bff

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"koreels/internal/ledger/application/ports/in"
	domainerrors "koreels/internal/ledger/domain/errors"
)

func TestNewMeAccountRoutes_WithIdentity_CreatesAccount_Returns201(t *testing.T) {
	server := createBFFTestServer(t)
	getUC := &mockGetAccount{}
	createUC := &mockCreateAccount{
		out: in.CreateAccountOutput{
			AccountId: "user-456",
			Balance:   0,
			Currency:  "CREDITS",
			CreatedAt: "2026-03-02T00:00:00Z",
		},
	}
	_, err := NewMeAccountRoutes(server, getUC, createUC)
	if err != nil {
		t.Fatal(err)
	}

	body := CreateMyAccountRequest{AllowNegative: false, Metadata: nil}
	bodyBytes, _ := json.Marshal(body)
	req, _ := http.NewRequest(http.MethodPost, "/me/account", bytes.NewBuffer(bodyBytes))
	req.Header.Set("Content-Type", "application/json")
	req = requestWithIdentity(req, "sub-456", "user-456")
	rec := httptest.NewRecorder()
	server.Manager.Mux.ServeHTTP(rec, req)

	if rec.Code != http.StatusCreated {
		t.Errorf("got status %v want %v", rec.Code, http.StatusCreated)
	}
}

func TestNewMeAccountRoutes_WhenAccountAlreadyExists_ReturnsSuccessWithExistingAccount(t *testing.T) {
	server := createBFFTestServer(t)
	createUC := &mockCreateAccount{err: domainerrors.ErrAccountAlreadyExists}
	getUC := &mockGetAccount{
		out: in.GetAccountOutput{
			AccountId:        "user-789",
			Balance:          500,
			AvailableBalance: 500,
			Currency:         "CREDITS",
			UpdatedAt:        "2026-03-02T00:00:00Z",
		},
	}
	_, err := NewMeAccountRoutes(server, getUC, createUC)
	if err != nil {
		t.Fatal(err)
	}

	req, _ := http.NewRequest(http.MethodPost, "/me/account", bytes.NewBuffer([]byte("{}")))
	req.Header.Set("Content-Type", "application/json")
	req = requestWithIdentity(req, "sub-789", "user-789")
	rec := httptest.NewRecorder()
	server.Manager.Mux.ServeHTTP(rec, req)

	// Idempotent: route uses DefaultStatusCode(201), so we get 2xx with existing account data
	if rec.Code < 200 || rec.Code >= 300 {
		t.Errorf("got status %v want 2xx (idempotent existing account)", rec.Code)
	}
}

func TestNewMeAccountRoutes_WithoutIdentity_Returns401(t *testing.T) {
	server := createBFFTestServer(t)
	getUC := &mockGetAccount{}
	createUC := &mockCreateAccount{}
	_, err := NewMeAccountRoutes(server, getUC, createUC)
	if err != nil {
		t.Fatal(err)
	}

	req, _ := http.NewRequest(http.MethodPost, "/me/account", bytes.NewBuffer([]byte("{}")))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	server.Manager.Mux.ServeHTTP(rec, req)

	if rec.Code != http.StatusUnauthorized {
		t.Errorf("got status %v want %v", rec.Code, http.StatusUnauthorized)
	}
}

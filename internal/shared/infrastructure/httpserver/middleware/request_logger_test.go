package middleware

import (
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"

	"ledger-service/internal/shared/infrastructure/observability"

	"github.com/stretchr/testify/assert"
	"go.opentelemetry.io/otel/trace/noop"
)

func TestRequestLogger(t *testing.T) {
	logger := slog.New(slog.NewJSONHandler(io.Discard, nil))
	obs := observability.Observability{Tracer: noop.NewTracerProvider().Tracer("test"), Logger: logger}
	mw := NewRequestLogger(obs)
	handler := mw(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusCreated)
		w.Write([]byte("ok"))
	}))
	req := httptest.NewRequest("GET", "/test", nil)
	w := httptest.NewRecorder()
	handler.ServeHTTP(w, req)
	assert.Equal(t, http.StatusCreated, w.Code)
	assert.Equal(t, "ok", w.Body.String())
}

# request-logger-middleware

> HTTP request logger middleware

## app/shared/infrastructure/httpserver/middleware/request_logger.go

```go
package middleware

import (
	"net"
	"net/http"
	"strings"
	"time"

	"archetype/app/shared/infrastructure/observability"

	"github.com/Ignaciojeria/ioc"
)

var _ = ioc.Register(NewRequestLogger)

type RequestLogger func(http.Handler) http.Handler

type statusWriter struct {
	http.ResponseWriter
	status int
}

func (w *statusWriter) WriteHeader(code int) {
	w.status = code
	w.ResponseWriter.WriteHeader(code)
}

func NewRequestLogger(obs observability.Observability) RequestLogger {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()

			ctx, span := obs.Tracer.Start(r.Context(), "http.request")
			defer span.End()

			sw := &statusWriter{
				ResponseWriter: w,
				status:         http.StatusOK, // default to 200
			}

			defer func() {
				fields := []any{
					"method", r.Method,
					"path", r.URL.Path,
					"remote_ip", clientIP(r),
					"status", sw.status,
					"duration_ms", time.Since(start).Milliseconds(),
				}

				if sw.status >= 500 {
					obs.Logger.ErrorContext(ctx, "http_request_failed", fields...)
				} else {
					obs.Logger.InfoContext(ctx, "http_request", fields...)
				}
			}()

			next.ServeHTTP(sw, r.WithContext(ctx))
		})
	}
}

func clientIP(r *http.Request) string {
	if fwd := r.Header.Get("Forwarded"); fwd != "" {
		return fwd
	}

	if xff := r.Header.Get("X-Forwarded-For"); xff != "" {
		parts := strings.Split(xff, ",")
		return strings.TrimSpace(parts[0])
	}

	if xrip := r.Header.Get("X-Real-IP"); xrip != "" {
		return xrip
	}

	host, _, err := net.SplitHostPort(r.RemoteAddr)
	if err == nil {
		return host
	}

	return r.RemoteAddr
}
```

---

## Unit tests

When creating a new component, generate tests following this pattern:

### app/shared/infrastructure/httpserver/middleware/request_logger_test.go

```go
package middleware

import (
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"

	"archetype/app/shared/infrastructure/observability"

	"github.com/stretchr/testify/assert"
	"go.opentelemetry.io/otel/trace/noop"
)

func TestRequestLogger(t *testing.T) {
	// Create dummy observability
	logger := slog.New(slog.NewJSONHandler(io.Discard, nil))

	obs := observability.Observability{
		Tracer: noop.NewTracerProvider().Tracer("test"),
		Logger: logger,
	}

	// Because we can't easily capture default slog output without redirecting stdout during tests,
	// we just ensure the middleware doesn't panic and executes the next handler.
	middleware := NewRequestLogger(obs)

	handler := middleware(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusCreated)
		w.Write([]byte("ok"))
	}))

	req := httptest.NewRequest("GET", "/test", nil)
	req.Header.Set("X-Forwarded-For", "192.168.1.1, 10.0.0.1")
	w := httptest.NewRecorder()

	handler.ServeHTTP(w, req)

	assert.Equal(t, http.StatusCreated, w.Code)
	assert.Equal(t, "ok", w.Body.String())
}

func TestClientIP(t *testing.T) {
	tests := []struct {
		name     string
		headers  map[string]string
		remote   string
		expected string
	}{
		{
			name:     "Forwarded Header",
			headers:  map[string]string{"Forwarded": "for=192.0.2.60"},
			remote:   "127.0.0.1:8080",
			expected: "for=192.0.2.60",
		},
		{
			name:     "X-Forwarded-For",
			headers:  map[string]string{"X-Forwarded-For": "203.0.113.195, 70.41.3.18"},
			remote:   "127.0.0.1:8080",
			expected: "203.0.113.195",
		},
		{
			name:     "X-Real-IP",
			headers:  map[string]string{"X-Real-IP": "198.51.100.1"},
			remote:   "127.0.0.1:8080",
			expected: "198.51.100.1",
		},
		{
			name:     "Fallback RemoteAddr with Port",
			headers:  map[string]string{},
			remote:   "192.168.0.1:1234",
			expected: "192.168.0.1",
		},
		{
			name:     "Fallback RemoteAddr no Port",
			headers:  map[string]string{},
			remote:   "10.0.0.1",
			expected: "10.0.0.1",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := httptest.NewRequest("GET", "/", nil)
			for k, v := range tt.headers {
				req.Header.Set(k, v)
			}
			req.RemoteAddr = tt.remote

			ip := clientIP(req)
			assert.Equal(t, tt.expected, ip)
		})
	}
}
```

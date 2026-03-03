package middleware

import (
	"net"
	"net/http"
	"strings"
	"time"

	"koreels/internal/shared/infrastructure/observability"

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
			sw := &statusWriter{ResponseWriter: w, status: http.StatusOK}
			defer func() {
				fields := []any{"method", r.Method, "path", r.URL.Path, "remote_ip", clientIP(r), "status", sw.status, "duration_ms", time.Since(start).Milliseconds()}
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
		return strings.TrimSpace(strings.Split(xff, ",")[0])
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

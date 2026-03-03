package middleware

import (
	"context"
	"net/http"
	"strings"
	"sync"

	"ledger-service/internal/shared/contextkeys"
	"ledger-service/internal/shared/configuration"

	"github.com/coreos/go-oidc/v3/oidc"
)

// OIDCConfig is the configuration for the OIDC middleware.
// When Issuer is empty the middleware is a no-op.
type OIDCConfig struct {
	Issuer    string // OIDC_ISSUER
	ClientID  string // OIDC_CLIENT_ID
	Audience  string // OIDC_AUDIENCE; if empty, ClientID is used for aud validation
	SkipPaths []string
}

// OIDC returns a middleware that validates JWT Bearer tokens and injects Identity into context.
// If config.Issuer is empty it passes through (no-op). Public paths (e.g. /health, /metrics) are skipped when listed in SkipPaths.
func OIDC(config OIDCConfig) func(http.Handler) http.Handler {
	if config.Issuer == "" {
		return func(next http.Handler) http.Handler {
			return next
		}
	}
	skip := make(map[string]struct{})
	for _, p := range config.SkipPaths {
		skip[p] = struct{}{}
	}
	audience := config.Audience
	if audience == "" {
		audience = config.ClientID
	}
	var (
		initOnce sync.Once
		provider *oidc.Provider
		verifier *oidc.IDTokenVerifier
		initErr  error
	)
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if _, ok := skip[r.URL.Path]; ok {
				next.ServeHTTP(w, r)
				return
			}
			initOnce.Do(func() {
				provider, initErr = oidc.NewProvider(context.Background(), config.Issuer)
				if initErr != nil {
					return
				}
				verifier = provider.Verifier(&oidc.Config{
					ClientID: audience,
				})
			})
			if initErr != nil {
				http.Error(w, "oidc provider unavailable", http.StatusInternalServerError)
				return
			}
			raw, ok := bearerToken(r)
			if !ok {
				http.Error(w, "missing or invalid authorization", http.StatusUnauthorized)
				return
			}
			idToken, err := verifier.Verify(r.Context(), raw)
			if err != nil {
				http.Error(w, "invalid token", http.StatusUnauthorized)
				return
			}
			identity := contextkeys.Identity{
				Subject: idToken.Subject,
			}
			var custom struct {
				AccountID string `json:"account_id"`
			}
			if _ = idToken.Claims(&custom); custom.AccountID != "" {
				identity.AccountID = custom.AccountID
			}
			ctx := contextkeys.WithIdentity(r.Context(), identity)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

// NewOIDCFromConf builds OIDC middleware from shared Conf. No-op when Conf.OIDC_ISSUER is empty.
func NewOIDCFromConf(conf configuration.Conf) func(http.Handler) http.Handler {
	return OIDC(OIDCConfig{
		Issuer:    conf.OIDC_ISSUER,
		ClientID:  conf.OIDC_CLIENT_ID,
		Audience:  conf.OIDC_AUDIENCE,
		SkipPaths: []string{"/health", "/metrics"},
	})
}

func bearerToken(r *http.Request) (string, bool) {
	const prefix = "Bearer "
	h := r.Header.Get("Authorization")
	if !strings.HasPrefix(h, prefix) {
		return "", false
	}
	return strings.TrimSpace(h[len(prefix):]), true
}

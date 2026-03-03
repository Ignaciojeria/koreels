package contextkeys

import "context"

type contextKey string

const (
	IdempotencyKey contextKey = "idempotency_key"
	IdentityKey    contextKey = "identity"
)

// Identity holds the authenticated user identity from OIDC (immutable, by value).
// Subject is the IdP subject (sub claim); AccountID is the business account id (claim or resolved from mapping).
type Identity struct {
	Subject   string
	AccountID string
}

// WithIdempotencyKey injects the idempotency key into context.
func WithIdempotencyKey(ctx context.Context, key string) context.Context {
	return context.WithValue(ctx, IdempotencyKey, key)
}

// GetIdempotencyKey extracts the idempotency key from context.
func GetIdempotencyKey(ctx context.Context) (string, bool) {
	v := ctx.Value(IdempotencyKey)
	if v == nil {
		return "", false
	}
	s, ok := v.(string)
	return s, ok
}

// WithIdentity injects the identity into context (by value, immutable).
func WithIdentity(ctx context.Context, id Identity) context.Context {
	return context.WithValue(ctx, IdentityKey, id)
}

// GetIdentity extracts the identity from context.
func GetIdentity(ctx context.Context) (Identity, bool) {
	v := ctx.Value(IdentityKey)
	if v == nil {
		return Identity{}, false
	}
	id, ok := v.(Identity)
	return id, ok
}

// GetSubject returns the OIDC sub claim from context, or "" if not set.
func GetSubject(ctx context.Context) (string, bool) {
	id, ok := GetIdentity(ctx)
	if !ok {
		return "", false
	}
	return id.Subject, id.Subject != ""
}

// GetAccountID returns the account id for the ledger from context (claim or mapping), or "" if not set.
func GetAccountID(ctx context.Context) (string, bool) {
	id, ok := GetIdentity(ctx)
	if !ok {
		return "", false
	}
	return id.AccountID, id.AccountID != ""
}

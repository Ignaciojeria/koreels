package contextkeys

import "context"

type contextKey string

const (
	IdempotencyKey contextKey = "idempotency_key"
)

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

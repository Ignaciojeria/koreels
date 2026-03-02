package out

import (
	"context"

	"ledger-service/internal/ledger/domain/entity"
)

type IdempotencyStore interface {
	FindByKey(ctx context.Context, key string) (*entity.IdempotencyRecord, error)
	Store(ctx context.Context, record entity.IdempotencyRecord) error
}

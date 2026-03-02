package out

import (
	"context"

	"ledger-service/internal/ledger/domain/entity"
)

type AccountRepository interface {
	Create(ctx context.Context, account entity.Account) error
	FindByID(ctx context.Context, accountId string) (*entity.Account, error)
}

package out

import (
	"context"

	"koreels/internal/ledger/domain/entity"
)

type TransactionRepository interface {
	Create(ctx context.Context, tx entity.Transaction, entries []entity.TransactionEntry) error
	FindByID(ctx context.Context, transactionId string) (*entity.Transaction, []entity.TransactionEntry, error)
}

package out

import (
	"context"

	"ledger-service/internal/ledger/domain/entity"
)

type TransactionExecutor interface {
	Execute(ctx context.Context, tx entity.Transaction, entries []TransactionEntryInput) ([]entity.TransactionEntry, error)
}

type TransactionEntryInput struct {
	AccountId string
	Amount    int64
}

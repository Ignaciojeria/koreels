package out

import (
	"context"

	"koreels/internal/ledger/domain/entity"
)

type LedgerEntryRepository interface {
	ListByAccount(ctx context.Context, accountId string, limit int, cursor string) ([]entity.LedgerEntry, string, error)
}

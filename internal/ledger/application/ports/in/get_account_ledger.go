package in

import "context"

type GetAccountLedgerExecutor interface {
	Execute(ctx context.Context, accountId string, limit int, cursor string) (GetAccountLedgerOutput, error)
}

type GetAccountLedgerOutput struct {
	AccountId  string
	Entries    []LedgerEntryOutput
	NextCursor string
}

type LedgerEntryOutput struct {
	TransactionId string
	Amount        int64
	BalanceAfter  int64
	CreatedAt     string
}

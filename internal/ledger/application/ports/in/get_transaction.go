package in

import "context"

type GetTransactionExecutor interface {
	Execute(ctx context.Context, transactionId string) (GetTransactionOutput, error)
}

type GetTransactionOutput struct {
	TransactionId string
	Entries       []TransactionEntryOutput
	Metadata      map[string]string
	CreatedAt     string
}

type TransactionEntryOutput struct {
	AccountId    string
	Amount       int64
	BalanceAfter int64
}

package in

import "context"

type CreateTransactionExecutor interface {
	Execute(ctx context.Context, req CreateTransactionInput) (CreateTransactionOutput, error)
}

type CreateTransactionInput struct {
	TransactionId string
	Entries       []TransactionEntryInput
	Metadata      map[string]string
}

type TransactionEntryInput struct {
	AccountId string
	Amount    int64
}

type CreateTransactionOutput struct {
	TransactionId string
	Status        string
	CreatedAt     string
}

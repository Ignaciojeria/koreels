package in

import "context"

type CreateAccountExecutor interface {
	Execute(ctx context.Context, req CreateAccountInput) (CreateAccountOutput, error)
}

type CreateAccountInput struct {
	AccountId     string
	Type          string
	Currency      string
	AllowNegative bool
	Metadata      map[string]string
}

type CreateAccountOutput struct {
	AccountId string
	Balance   int64
	Currency  string
	CreatedAt string
}

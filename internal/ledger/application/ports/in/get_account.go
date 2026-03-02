package in

import "context"

type GetAccountExecutor interface {
	Execute(ctx context.Context, accountId string) (GetAccountOutput, error)
}

type GetAccountOutput struct {
	AccountId        string
	Balance          int64
	AvailableBalance int64
	Currency         string
	UpdatedAt        string
}

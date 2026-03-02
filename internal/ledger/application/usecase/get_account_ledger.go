package usecase

import (
	"context"

	"ledger-service/internal/ledger/application/ports/in"
	"ledger-service/internal/ledger/application/ports/out"

	"github.com/Ignaciojeria/ioc"
)

var _ = ioc.Register(NewGetAccountLedgerUseCase)

type getAccountLedgerUseCase struct {
	repo out.LedgerEntryRepository
}

func NewGetAccountLedgerUseCase(repo out.LedgerEntryRepository) (in.GetAccountLedgerExecutor, error) {
	return &getAccountLedgerUseCase{repo: repo}, nil
}

func (uc *getAccountLedgerUseCase) Execute(ctx context.Context, accountId string, limit int, cursor string) (in.GetAccountLedgerOutput, error) {
	entries, nextCursor, err := uc.repo.ListByAccount(ctx, accountId, limit, cursor)
	if err != nil {
		return in.GetAccountLedgerOutput{}, err
	}
	result := make([]in.LedgerEntryOutput, len(entries))
	for i, e := range entries {
		result[i] = in.LedgerEntryOutput{
			TransactionId: e.TransactionId,
			Amount:        e.Amount,
			BalanceAfter:  e.BalanceAfter,
			CreatedAt:     e.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
		}
	}
	return in.GetAccountLedgerOutput{AccountId: accountId, Entries: result, NextCursor: nextCursor}, nil
}

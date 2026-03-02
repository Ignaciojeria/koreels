package usecase

import (
	"context"

	"ledger-service/internal/ledger/application/ports/in"
	"ledger-service/internal/ledger/application/ports/out"
	domainerrors "ledger-service/internal/ledger/domain/errors"

	"github.com/Ignaciojeria/ioc"
)

var _ = ioc.Register(NewGetTransactionUseCase)

type getTransactionUseCase struct {
	repo out.TransactionRepository
}

func NewGetTransactionUseCase(repo out.TransactionRepository) (in.GetTransactionExecutor, error) {
	return &getTransactionUseCase{repo: repo}, nil
}

func (uc *getTransactionUseCase) Execute(ctx context.Context, transactionId string) (in.GetTransactionOutput, error) {
	tx, entries, err := uc.repo.FindByID(ctx, transactionId)
	if err != nil {
		return in.GetTransactionOutput{}, err
	}
	if tx == nil {
		return in.GetTransactionOutput{}, domainerrors.ErrTransactionNotFound
	}
	result := make([]in.TransactionEntryOutput, len(entries))
	for i, e := range entries {
		result[i] = in.TransactionEntryOutput{
			AccountId:    e.AccountId,
			Amount:       e.Amount,
			BalanceAfter: e.BalanceAfter,
		}
	}
	return in.GetTransactionOutput{
		TransactionId: tx.TransactionId,
		Entries:       result,
		Metadata:      tx.Metadata,
		CreatedAt:     tx.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
	}, nil
}

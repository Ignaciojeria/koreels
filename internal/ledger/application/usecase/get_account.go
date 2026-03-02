package usecase

import (
	"context"

	"ledger-service/internal/ledger/application/ports/in"
	"ledger-service/internal/ledger/application/ports/out"
	domainerrors "ledger-service/internal/ledger/domain/errors"

	"github.com/Ignaciojeria/ioc"
)

var _ = ioc.Register(NewGetAccountUseCase)

type getAccountUseCase struct {
	repo out.AccountRepository
}

func NewGetAccountUseCase(repo out.AccountRepository) (in.GetAccountExecutor, error) {
	return &getAccountUseCase{repo: repo}, nil
}

func (uc *getAccountUseCase) Execute(ctx context.Context, accountId string) (in.GetAccountOutput, error) {
	acc, err := uc.repo.FindByID(ctx, accountId)
	if err != nil {
		if err == domainerrors.ErrAccountNotFound {
			return in.GetAccountOutput{}, err
		}
		return in.GetAccountOutput{}, err
	}
	return in.GetAccountOutput{
		AccountId:        acc.AccountId,
		Balance:          acc.Balance,
		AvailableBalance: acc.Balance,
		Currency:         acc.Currency,
		UpdatedAt:        acc.UpdatedAt.Format("2006-01-02T15:04:05Z07:00"),
	}, nil
}

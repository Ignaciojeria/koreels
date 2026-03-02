package usecase

import (
	"context"
	"time"

	"ledger-service/internal/ledger/application/ports/in"
	"ledger-service/internal/ledger/application/ports/out"
	"ledger-service/internal/ledger/domain/entity"

	"github.com/Ignaciojeria/ioc"
)

var _ = ioc.Register(NewCreateAccountUseCase)

type createAccountUseCase struct {
	repo out.AccountRepository
}

func NewCreateAccountUseCase(repo out.AccountRepository) (in.CreateAccountExecutor, error) {
	return &createAccountUseCase{repo: repo}, nil
}

func (uc *createAccountUseCase) Execute(ctx context.Context, req in.CreateAccountInput) (in.CreateAccountOutput, error) {
	now := time.Now()
	acc := entity.Account{
		AccountId:     req.AccountId,
		Type:          entity.AccountType(req.Type),
		Currency:      req.Currency,
		AllowNegative: req.AllowNegative,
		Metadata:      req.Metadata,
		CreatedAt:     now,
		UpdatedAt:     now,
	}
	if err := uc.repo.Create(ctx, acc); err != nil {
		return in.CreateAccountOutput{}, err
	}
	return in.CreateAccountOutput{
		AccountId: req.AccountId,
		Balance:   0,
		Currency:  req.Currency,
		CreatedAt: now.Format(time.RFC3339),
	}, nil
}

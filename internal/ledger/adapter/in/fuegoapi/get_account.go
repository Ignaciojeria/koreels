package fuegoapi

import (
	"net/http"

	"koreels/internal/ledger/application/ports/in"
	domainerrors "koreels/internal/ledger/domain/errors"
	"koreels/internal/shared/infrastructure/httpserver"

	"github.com/Ignaciojeria/ioc"
	fuegofw "github.com/go-fuego/fuego"
	"github.com/go-fuego/fuego/option"
)

var _ = ioc.Register(NewGetAccount)

type GetAccountResponse struct {
	AccountId        string `json:"accountId"`
	Balance          int64  `json:"balance"`
	AvailableBalance int64  `json:"availableBalance"`
	Currency         string `json:"currency"`
	UpdatedAt        string `json:"updatedAt"`
}

func NewGetAccount(s *httpserver.Server, uc in.GetAccountExecutor) {
	fuegofw.Get(s.Manager, "/accounts/{accountId}",
		func(c fuegofw.ContextNoBody) (GetAccountResponse, error) {
			accountId := c.PathParam("accountId")
			out, err := uc.Execute(c.Context(), accountId)
			if err != nil {
				if err == domainerrors.ErrAccountNotFound {
					return GetAccountResponse{}, fuegofw.HTTPError{Err: err, Status: http.StatusNotFound, Detail: err.Error()}
				}
				return GetAccountResponse{}, err
			}
			return GetAccountResponse{
				AccountId:        out.AccountId,
				Balance:          out.Balance,
				AvailableBalance: out.AvailableBalance,
				Currency:         out.Currency,
				UpdatedAt:        out.UpdatedAt,
			}, nil
		},
		option.Summary("getAccount"),
	)
}

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

var _ = ioc.Register(NewGetTransaction)

type GetTransactionResponse struct {
	TransactionId string                `json:"transactionId"`
	Entries       []TransactionEntryDTO `json:"entries"`
	Metadata      map[string]string     `json:"metadata"`
	CreatedAt     string                `json:"createdAt"`
}

type TransactionEntryDTO struct {
	AccountId    string `json:"accountId"`
	Amount       int64  `json:"amount"`
	BalanceAfter int64  `json:"balanceAfter"`
}

func NewGetTransaction(s *httpserver.Server, uc in.GetTransactionExecutor) {
	fuegofw.Get(s.Manager, "/transactions/{transactionId}",
		func(c fuegofw.ContextNoBody) (GetTransactionResponse, error) {
			transactionId := c.PathParam("transactionId")
			out, err := uc.Execute(c.Context(), transactionId)
			if err != nil {
				if err == domainerrors.ErrTransactionNotFound {
					return GetTransactionResponse{}, fuegofw.HTTPError{Err: err, Status: http.StatusNotFound, Detail: err.Error()}
				}
				return GetTransactionResponse{}, err
			}
			entries := make([]TransactionEntryDTO, len(out.Entries))
			for i, e := range out.Entries {
				entries[i] = TransactionEntryDTO{
					AccountId:    e.AccountId,
					Amount:       e.Amount,
					BalanceAfter: e.BalanceAfter,
				}
			}
			return GetTransactionResponse{
				TransactionId: out.TransactionId,
				Entries:       entries,
				Metadata:      out.Metadata,
				CreatedAt:     out.CreatedAt,
			}, nil
		},
		option.Summary("getTransaction"),
	)
}

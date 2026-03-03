package fuegoapi

import (
	"strconv"

	"koreels/internal/ledger/application/ports/in"
	"koreels/internal/shared/infrastructure/httpserver"

	"github.com/Ignaciojeria/ioc"
	fuegofw "github.com/go-fuego/fuego"
	"github.com/go-fuego/fuego/option"
	"github.com/go-fuego/fuego/param"
)

var _ = ioc.Register(NewGetAccountLedger)

type GetAccountLedgerResponse struct {
	AccountId  string           `json:"accountId"`
	Entries    []LedgerEntryDTO `json:"entries"`
	NextCursor string           `json:"nextCursor,omitempty"`
}

type LedgerEntryDTO struct {
	TransactionId string `json:"transactionId"`
	Amount        int64  `json:"amount"`
	BalanceAfter  int64  `json:"balanceAfter"`
	CreatedAt     string `json:"createdAt"`
}

func NewGetAccountLedger(s *httpserver.Server, uc in.GetAccountLedgerExecutor) {
	fuegofw.Get(s.Manager, "/accounts/{accountId}/ledger",
		func(c fuegofw.ContextNoBody) (GetAccountLedgerResponse, error) {
			accountId := c.PathParam("accountId")
			limit := 50
			if l := c.QueryParam("limit"); l != "" {
				if n, err := strconv.Atoi(l); err == nil && n > 0 {
					limit = n
				}
			}
			cursor := c.QueryParam("cursor")

			out, err := uc.Execute(c.Context(), accountId, limit, cursor)
			if err != nil {
				return GetAccountLedgerResponse{}, err
			}

			entries := make([]LedgerEntryDTO, len(out.Entries))
			for i, e := range out.Entries {
				entries[i] = LedgerEntryDTO{
					TransactionId: e.TransactionId,
					Amount:        e.Amount,
					BalanceAfter:  e.BalanceAfter,
					CreatedAt:     e.CreatedAt,
				}
			}
			return GetAccountLedgerResponse{
				AccountId:  out.AccountId,
				Entries:    entries,
				NextCursor: out.NextCursor,
			}, nil
		},
		option.Summary("getAccountLedger"),
		option.QueryInt("limit", "Number of entries", param.Default(50)),
		option.Query("cursor", "Pagination cursor"),
	)
}

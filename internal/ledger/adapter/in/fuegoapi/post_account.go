package fuegoapi

import (
	"net/http"

	"ledger-service/internal/ledger/application/ports/in"
	"ledger-service/internal/shared/infrastructure/httpserver"

	"github.com/Ignaciojeria/ioc"
	fuegofw "github.com/go-fuego/fuego"
	"github.com/go-fuego/fuego/option"
)

var _ = ioc.Register(NewPostAccount)

type CreateAccountRequest struct {
	AccountId     string            `json:"accountId" validate:"required,uuid"`
	Type          string            `json:"type" validate:"required,oneof=USER SYSTEM PROMOTION"`
	Currency      string            `json:"currency" validate:"required,oneof=CREDITS"`
	AllowNegative bool              `json:"allowNegative"`
	Metadata      map[string]string `json:"metadata"`
}

type CreateAccountResponse struct {
	AccountId string `json:"accountId"`
	Balance   int64  `json:"balance"`
	Currency  string `json:"currency"`
	CreatedAt string `json:"createdAt"`
}

func NewPostAccount(s *httpserver.Server, uc in.CreateAccountExecutor) {
	fuegofw.Post(s.Manager, "/accounts",
		func(c fuegofw.ContextWithBody[CreateAccountRequest]) (CreateAccountResponse, error) {
			body, err := c.Body()
			if err != nil {
				return CreateAccountResponse{}, err
			}
			out, err := uc.Execute(c.Context(), in.CreateAccountInput{
				AccountId:     body.AccountId,
				Type:          body.Type,
				Currency:      body.Currency,
				AllowNegative: body.AllowNegative,
				Metadata:      body.Metadata,
			})
			if err != nil {
				return CreateAccountResponse{}, err
			}
			return CreateAccountResponse{
				AccountId: out.AccountId,
				Balance:   out.Balance,
				Currency:  out.Currency,
				CreatedAt: out.CreatedAt,
			}, nil
		},
		option.Summary("createAccount"),
		option.DefaultStatusCode(http.StatusCreated),
	)
}

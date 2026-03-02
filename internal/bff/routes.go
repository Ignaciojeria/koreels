package bff

import (
	"net/http"

	"ledger-service/internal/ledger/application/ports/in"
	"ledger-service/internal/shared/infrastructure/httpserver"

	"github.com/Ignaciojeria/ioc"
	fuegofw "github.com/go-fuego/fuego"
	"github.com/go-fuego/fuego/option"
)

var _ = ioc.Register(NewBFFRoutes)

// NewBFFRoutes registra solo rutas BFF (ej. /me/balance). No expone el API del ledger.
// El ledger se usa en memoria vía los ports in (GetAccountExecutor, etc.).
func NewBFFRoutes(s *httpserver.Server, getAccount in.GetAccountExecutor) (*BFFRoutes, error) {
	fuegofw.Get(s.Manager, "/me/balance",
		func(c fuegofw.ContextNoBody) (BalanceResponse, error) {
			accountId := c.QueryParam("accountId")
			if accountId == "" {
				return BalanceResponse{}, fuegofw.HTTPError{
					Err:    errMissingAccountId,
					Status: http.StatusBadRequest,
					Detail: "query param accountId is required",
				}
			}
			out, err := getAccount.Execute(c.Context(), accountId)
			if err != nil {
				return BalanceResponse{}, err
			}
			return BalanceResponse{
				AccountId: out.AccountId,
				Balance:   out.Balance,
				Currency:  out.Currency,
			}, nil
		},
		option.Summary("getMyBalance"),
		option.Query("accountId", "Account ID (e.g. from auth)"),
	)
	return &BFFRoutes{}, nil
}

type BFFRoutes struct{}

type BalanceResponse struct {
	AccountId string `json:"accountId"`
	Balance   int64  `json:"balance"`
	Currency  string `json:"currency"`
}

var errMissingAccountId = &errMsg{"accountId required"}

type errMsg struct{ msg string }

func (e *errMsg) Error() string { return e.msg }

package bff

import (
	"net/http"

	"ledger-service/internal/ledger/application/ports/in"
	"ledger-service/internal/shared/contextkeys"
	"ledger-service/internal/shared/infrastructure/httpserver"

	"github.com/Ignaciojeria/ioc"
	fuegofw "github.com/go-fuego/fuego"
	"github.com/go-fuego/fuego/option"
)

var _ = ioc.Register(NewMeBalanceRoutes)

// NewMeBalanceRoutes registra GET /me/balance. No expone el API del ledger.
// La identidad (accountId) se lee del context inyectado por el middleware OIDC.
func NewMeBalanceRoutes(s *httpserver.Server, getAccount in.GetAccountExecutor) (*MeBalanceRoutes, error) {
	fuegofw.Get(s.Manager, "/me/balance",
		func(c fuegofw.ContextNoBody) (BalanceResponse, error) {
			accountId, ok := contextkeys.GetAccountID(c.Context())
			if !ok || accountId == "" {
				accountId, _ = contextkeys.GetSubject(c.Context())
			}
			if accountId == "" {
				return BalanceResponse{}, fuegofw.HTTPError{
					Err:    errMissingIdentity,
					Status: http.StatusUnauthorized,
					Detail: "authentication required",
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
	)
	return &MeBalanceRoutes{}, nil
}

// MeBalanceRoutes agrupa la ruta GET /me/balance (registrada por NewMeBalanceRoutes).
type MeBalanceRoutes struct{}

type BalanceResponse struct {
	AccountId string `json:"accountId"`
	Balance   int64  `json:"balance"`
	Currency  string `json:"currency"`
}

var errMissingIdentity = &errMsg{"authentication required"}

type errMsg struct{ msg string }

func (e *errMsg) Error() string { return e.msg }

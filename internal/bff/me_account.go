package bff

import (
	"net/http"

	"koreels/internal/ledger/application/ports/in"
	domainerrors "koreels/internal/ledger/domain/errors"
	"koreels/internal/shared/contextkeys"
	"koreels/internal/shared/infrastructure/httpserver"

	"github.com/Ignaciojeria/ioc"
	fuegofw "github.com/go-fuego/fuego"
	"github.com/go-fuego/fuego/option"
)

const (
	defaultAccountType     = "USER"
	defaultAccountCurrency = "CREDITS"
)

// CreateMyAccountRequest is the optional body for POST /me/account. Empty body is valid.
type CreateMyAccountRequest struct {
	AllowNegative bool              `json:"allowNegative"`
	Metadata      map[string]string `json:"metadata"`
}

// CreateMyAccountResponse is returned by POST /me/account (201 created or 200 already existed).
type CreateMyAccountResponse struct {
	AccountId string `json:"accountId"`
	Balance   int64  `json:"balance"`
	Currency  string `json:"currency"`
	CreatedAt string `json:"createdAt"`
}

var _ = ioc.Register(NewMeAccountRoutes)

// NewMeAccountRoutes registra POST /me/account (crear/asegurar cuenta del usuario autenticado).
func NewMeAccountRoutes(s *httpserver.Server, getAccount in.GetAccountExecutor, createAccount in.CreateAccountExecutor) (*MeAccountRoutes, error) {
	fuegofw.Post(s.Manager, "/me/account",
		func(c fuegofw.ContextWithBody[*CreateMyAccountRequest]) (CreateMyAccountResponse, error) {
			accountId, ok := contextkeys.GetAccountID(c.Context())
			if !ok || accountId == "" {
				accountId, _ = contextkeys.GetSubject(c.Context())
			}
			if accountId == "" {
				return CreateMyAccountResponse{}, fuegofw.HTTPError{
					Err:    errMissingIdentity,
					Status: http.StatusUnauthorized,
					Detail: "authentication required",
				}
			}
			body, _ := c.Body()
			if body == nil {
				body = &CreateMyAccountRequest{}
			}
			req := in.CreateAccountInput{
				AccountId:     accountId,
				Type:          defaultAccountType,
				Currency:      defaultAccountCurrency,
				AllowNegative: body.AllowNegative,
				Metadata:      body.Metadata,
			}
			out, err := createAccount.Execute(c.Context(), req)
			if err != nil {
				if err == domainerrors.ErrAccountAlreadyExists {
					existing, getErr := getAccount.Execute(c.Context(), accountId)
					if getErr != nil {
						return CreateMyAccountResponse{}, getErr
					}
					return CreateMyAccountResponse{
						AccountId: existing.AccountId,
						Balance:   existing.Balance,
						Currency:  existing.Currency,
						CreatedAt: existing.UpdatedAt,
					}, nil
				}
				return CreateMyAccountResponse{}, err
			}
			return CreateMyAccountResponse{
				AccountId: out.AccountId,
				Balance:   out.Balance,
				Currency:  out.Currency,
				CreatedAt: out.CreatedAt,
			}, nil
		},
		option.Summary("createMyAccount"),
		option.DefaultStatusCode(http.StatusCreated),
	)
	return &MeAccountRoutes{}, nil
}

// MeAccountRoutes agrupa la ruta POST /me/account (registrada por NewMeAccountRoutes).
type MeAccountRoutes struct{}

package fuegoapi

import (
	"errors"
	"net/http"

	"koreels/internal/ledger/application/ports/in"
	"koreels/internal/ledger/application/usecase"
	domainerrors "koreels/internal/ledger/domain/errors"
	"koreels/internal/shared/contextkeys"
	"koreels/internal/shared/infrastructure/httpserver"

	"github.com/Ignaciojeria/ioc"
	fuegofw "github.com/go-fuego/fuego"
	"github.com/go-fuego/fuego/option"
	"github.com/go-fuego/fuego/param"
)

var _ = ioc.Register(NewPostTransaction)

const idempotencyKeyHeader = "Idempotency-Key"

type CreateTransactionRequest struct {
	TransactionId string                     `json:"transactionId" validate:"required,uuid"`
	Entries       []TransactionEntryInputReq `json:"entries" validate:"required,min=2,dive"`
	Metadata      map[string]string          `json:"metadata"`
}

type TransactionEntryInputReq struct {
	AccountId string `json:"accountId" validate:"required,uuid"`
	Amount    int64  `json:"amount"`
}

type CreateTransactionResponse struct {
	TransactionId string `json:"transactionId"`
	Status        string `json:"status"`
	CreatedAt     string `json:"createdAt"`
}

func NewPostTransaction(s *httpserver.Server, uc in.CreateTransactionExecutor) {
		fuegofw.Post(s.Manager, "/transactions",
		func(c fuegofw.ContextWithBody[CreateTransactionRequest]) (CreateTransactionResponse, error) {
			idempotencyKey := c.Header(idempotencyKeyHeader)
			if idempotencyKey == "" {
				return CreateTransactionResponse{}, fuegofw.HTTPError{Err: usecase.ErrMissingIdempotencyKey, Status: http.StatusBadRequest, Detail: usecase.ErrMissingIdempotencyKey.Error()}
			}
			ctx := contextkeys.WithIdempotencyKey(c.Context(), idempotencyKey)

			body, err := c.Body()
			if err != nil {
				return CreateTransactionResponse{}, err
			}
			entries := make([]in.TransactionEntryInput, len(body.Entries))
			for i, e := range body.Entries {
				entries[i] = in.TransactionEntryInput{AccountId: e.AccountId, Amount: e.Amount}
			}
			req := in.CreateTransactionInput{
				TransactionId: body.TransactionId,
				Entries:       entries,
				Metadata:      body.Metadata,
			}
			out, err := uc.Execute(ctx, req)
			if err != nil {
				if errors.Is(err, usecase.ErrMissingIdempotencyKey) {
					return CreateTransactionResponse{}, fuegofw.HTTPError{Err: err, Status: http.StatusBadRequest, Detail: err.Error()}
				}
				if err == domainerrors.ErrUnbalancedTransaction {
					return CreateTransactionResponse{}, fuegofw.HTTPError{Err: err, Status: http.StatusBadRequest, Detail: err.Error()}
				}
				if err == domainerrors.ErrIdempotencyConflict {
					return CreateTransactionResponse{}, fuegofw.HTTPError{Err: err, Status: http.StatusConflict, Detail: err.Error()}
				}
				if err == domainerrors.ErrInsufficientBalance || err == domainerrors.ErrAccountNotFound {
					return CreateTransactionResponse{}, fuegofw.HTTPError{Err: err, Status: http.StatusUnprocessableEntity, Detail: err.Error()}
				}
				return CreateTransactionResponse{}, err
			}
			return CreateTransactionResponse{
				TransactionId: out.TransactionId,
				Status:        out.Status,
				CreatedAt:     out.CreatedAt,
			}, nil
		},
		option.Summary("createTransaction"),
		option.Header(idempotencyKeyHeader, "Idempotency key for safe retries", param.Required()),
		option.DefaultStatusCode(http.StatusCreated),
	)
}

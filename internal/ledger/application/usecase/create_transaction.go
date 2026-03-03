package usecase

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"time"

	"koreels/internal/ledger/application/ports/in"
	"koreels/internal/ledger/application/ports/out"
	"koreels/internal/ledger/domain/entity"
	domainerrors "koreels/internal/ledger/domain/errors"
	"koreels/internal/shared/contextkeys"

	"github.com/Ignaciojeria/ioc"
)

var _ = ioc.Register(NewCreateTransactionUseCase)

type createTransactionUseCase struct {
	executor    out.TransactionExecutor
	idempotency out.IdempotencyStore
	txRepo      out.TransactionRepository
}

func NewCreateTransactionUseCase(executor out.TransactionExecutor, idempotency out.IdempotencyStore, txRepo out.TransactionRepository) (in.CreateTransactionExecutor, error) {
	return &createTransactionUseCase{executor: executor, idempotency: idempotency, txRepo: txRepo}, nil
}

// ErrMissingIdempotencyKey is returned when the Idempotency-Key header is missing.
var ErrMissingIdempotencyKey = errors.New("idempotency key is required (Idempotency-Key header)")

func (uc *createTransactionUseCase) Execute(ctx context.Context, req in.CreateTransactionInput) (in.CreateTransactionOutput, error) {
	idempotencyKey, ok := contextkeys.GetIdempotencyKey(ctx)
	if !ok || idempotencyKey == "" {
		return in.CreateTransactionOutput{}, ErrMissingIdempotencyKey
	}

	hash := computeRequestHash(req)

	existing, err := uc.idempotency.FindByKey(ctx, idempotencyKey)
	if err != nil {
		return in.CreateTransactionOutput{}, err
	}
	if existing != nil {
		if existing.RequestHash != hash {
			return in.CreateTransactionOutput{}, domainerrors.ErrIdempotencyConflict
		}
		if len(existing.ResponsePayload) > 0 {
			var out in.CreateTransactionOutput
			if err := json.Unmarshal(existing.ResponsePayload, &out); err == nil {
				return out, nil
			}
		}
		tx, _, _ := uc.txRepo.FindByID(ctx, existing.TransactionId)
		if tx != nil {
			return in.CreateTransactionOutput{
				TransactionId: tx.TransactionId,
				Status:        string(tx.Status),
				CreatedAt:     tx.CreatedAt.Format(time.RFC3339),
			}, nil
		}
	}

	var total int64
	for _, e := range req.Entries {
		total += e.Amount
	}
	if total != 0 {
		return in.CreateTransactionOutput{}, domainerrors.ErrUnbalancedTransaction
	}

	entries := make([]out.TransactionEntryInput, len(req.Entries))
	for i, e := range req.Entries {
		entries[i] = out.TransactionEntryInput{AccountId: e.AccountId, Amount: e.Amount}
	}

	txn := entity.Transaction{
		TransactionId: req.TransactionId,
		Status:        entity.TransactionStatusCommitted,
		Metadata:      req.Metadata,
	}

	_, err = uc.executor.Execute(ctx, txn, entries)
	if err != nil {
		return in.CreateTransactionOutput{}, err
	}

	output := in.CreateTransactionOutput{
		TransactionId: req.TransactionId,
		Status:        string(entity.TransactionStatusCommitted),
		CreatedAt:     time.Now().Format(time.RFC3339),
	}
	responsePayload, _ := json.Marshal(output)

	if err := uc.idempotency.Store(ctx, entity.IdempotencyRecord{
		IdempotencyKey:   idempotencyKey,
		TransactionId:    req.TransactionId,
		RequestHash:      hash,
		ResponsePayload:  responsePayload,
	}); err != nil {
		return in.CreateTransactionOutput{}, err
	}

	return output, nil
}

func computeRequestHash(req in.CreateTransactionInput) string {
	body := struct {
		TransactionId string                     `json:"transactionId"`
		Entries       []in.TransactionEntryInput `json:"entries"`
		Metadata      map[string]string          `json:"metadata"`
	}{TransactionId: req.TransactionId, Entries: req.Entries, Metadata: req.Metadata}
	b, _ := json.Marshal(body)
	h := sha256.Sum256(b)
	return hex.EncodeToString(h[:])
}

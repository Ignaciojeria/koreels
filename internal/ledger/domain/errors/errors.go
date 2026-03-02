package errors

import "errors"

var (
	ErrUnbalancedTransaction = errors.New("transaction entries must sum to zero")
	ErrAccountNotFound       = errors.New("account not found")
	ErrTransactionNotFound   = errors.New("transaction not found")
	ErrInsufficientBalance   = errors.New("insufficient balance")
	ErrIdempotencyConflict   = errors.New("idempotency key already used with different payload")
)

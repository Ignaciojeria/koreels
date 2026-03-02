package entity

import "time"

type TransactionStatus string

const (
	TransactionStatusCommitted TransactionStatus = "COMMITTED"
)

type Transaction struct {
	TransactionId string
	Status        TransactionStatus
	Metadata      map[string]string
	CreatedAt     time.Time
}

type TransactionEntry struct {
	TransactionId string
	AccountId     string
	Amount        int64
	BalanceAfter  int64
}

package entity

import "time"

type LedgerEntry struct {
	TransactionId string
	AccountId     string
	Amount        int64
	BalanceAfter  int64
	CreatedAt     time.Time
}

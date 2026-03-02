package entity

import "time"

type AccountType string

const (
	AccountTypeUser      AccountType = "USER"
	AccountTypeSystem    AccountType = "SYSTEM"
	AccountTypePromotion AccountType = "PROMOTION"
)

type Account struct {
	AccountId     string
	Type          AccountType
	Currency      string
	AllowNegative bool
	Balance       int64
	Metadata      map[string]string
	CreatedAt     time.Time
	UpdatedAt     time.Time
}

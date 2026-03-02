package postgres

import (
	"context"
	"strconv"
	"time"

	"ledger-service/internal/ledger/application/ports/out"
	"ledger-service/internal/ledger/domain/entity"

	"github.com/Ignaciojeria/ioc"
	"github.com/jmoiron/sqlx"
)

var _ = ioc.Register(NewLedgerEntryRepository)

type ledgerEntryDB struct {
	ID            int64     `db:"id"`
	TransactionId string    `db:"transaction_id"`
	AccountId     string    `db:"account_id"`
	Amount        int64     `db:"amount"`
	BalanceAfter  int64     `db:"balance_after"`
	CreatedAt     time.Time `db:"created_at"`
}

func toLedgerEntryDomain(m ledgerEntryDB) entity.LedgerEntry {
	return entity.LedgerEntry{
		TransactionId: m.TransactionId,
		AccountId:     m.AccountId,
		Amount:        m.Amount,
		BalanceAfter:  m.BalanceAfter,
		CreatedAt:     m.CreatedAt,
	}
}

type ledgerEntryRepository struct {
	db *sqlx.DB
}

func NewLedgerEntryRepository(db *sqlx.DB) (out.LedgerEntryRepository, error) {
	return &ledgerEntryRepository{db: db}, nil
}

func (r *ledgerEntryRepository) ListByAccount(ctx context.Context, accountId string, limit int, cursor string) ([]entity.LedgerEntry, string, error) {
	if limit <= 0 {
		limit = 50
	}
	if limit > 100 {
		limit = 100
	}

	query := `SELECT id, transaction_id, account_id, amount, balance_after, created_at FROM ledger_entries WHERE account_id = $1`
	args := []interface{}{accountId}
	argIdx := 2

	if cursor != "" {
		cursorID, err := strconv.ParseInt(cursor, 10, 64)
		if err == nil {
			query += ` AND id < $` + strconv.Itoa(argIdx)
			args = append(args, cursorID)
			argIdx++
		}
	}

	query += ` ORDER BY id DESC LIMIT $` + strconv.Itoa(argIdx)
	args = append(args, limit+1)

	var rows []ledgerEntryDB
	err := r.db.SelectContext(ctx, &rows, query, args...)
	if err != nil {
		return nil, "", err
	}

	nextCursor := ""
	toReturn := len(rows)
	if len(rows) > limit {
		toReturn = limit
		nextCursor = strconv.FormatInt(rows[limit].ID, 10)
	}

	result := make([]entity.LedgerEntry, toReturn)
	for i := 0; i < toReturn; i++ {
		result[i] = toLedgerEntryDomain(rows[i])
	}
	return result, nextCursor, nil
}

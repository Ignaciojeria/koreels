package postgres

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"time"

	"koreels/internal/ledger/application/ports/out"
	"koreels/internal/ledger/domain/entity"

	"github.com/Ignaciojeria/ioc"
	"github.com/jmoiron/sqlx"
)

var _ = ioc.Register(NewTransactionRepository)

type transactionDB struct {
	TransactionId string          `db:"transaction_id"`
	Status        string          `db:"status"`
	Metadata      json.RawMessage `db:"metadata"`
	CreatedAt     time.Time       `db:"created_at"`
}

type transactionEntryDB struct {
	AccountId    string `db:"account_id"`
	Amount       int64  `db:"amount"`
	BalanceAfter int64  `db:"balance_after"`
}

func toTransactionDomain(tx transactionDB, entries []transactionEntryDB) (*entity.Transaction, []entity.TransactionEntry) {
	metadata := make(map[string]string)
	if len(tx.Metadata) > 0 {
		_ = json.Unmarshal(tx.Metadata, &metadata)
	}
	txn := &entity.Transaction{
		TransactionId: tx.TransactionId,
		Status:        entity.TransactionStatus(tx.Status),
		Metadata:      metadata,
		CreatedAt:     tx.CreatedAt,
	}
	ents := make([]entity.TransactionEntry, len(entries))
	for i, e := range entries {
		ents[i] = entity.TransactionEntry{
			TransactionId: tx.TransactionId,
			AccountId:     e.AccountId,
			Amount:        e.Amount,
			BalanceAfter:  e.BalanceAfter,
		}
	}
	return txn, ents
}

func fromTransactionDomain(tx entity.Transaction) transactionDB {
	metadata, _ := json.Marshal(tx.Metadata)
	return transactionDB{
		TransactionId: tx.TransactionId,
		Status:        string(tx.Status),
		Metadata:      metadata,
	}
}

type transactionRepository struct {
	db *sqlx.DB
}

func NewTransactionRepository(db *sqlx.DB) (out.TransactionRepository, error) {
	return &transactionRepository{db: db}, nil
}

func (r *transactionRepository) Create(ctx context.Context, tx entity.Transaction, entries []entity.TransactionEntry) error {
	txx, err := r.db.BeginTxx(ctx, nil)
	if err != nil {
		return err
	}
	defer txx.Rollback()

	m := fromTransactionDomain(tx)
	_, err = txx.ExecContext(ctx, `INSERT INTO transactions (transaction_id, status, metadata) VALUES ($1, $2, $3)`, m.TransactionId, m.Status, m.Metadata)
	if err != nil {
		return err
	}

	for _, e := range entries {
		_, err = txx.ExecContext(ctx, `INSERT INTO ledger_entries (transaction_id, account_id, amount, balance_after) VALUES ($1, $2, $3, $4)`, e.TransactionId, e.AccountId, e.Amount, e.BalanceAfter)
		if err != nil {
			return err
		}
	}

	return txx.Commit()
}

func (r *transactionRepository) FindByID(ctx context.Context, transactionId string) (*entity.Transaction, []entity.TransactionEntry, error) {
	var tx transactionDB
	err := r.db.GetContext(ctx, &tx, `SELECT transaction_id, status, metadata, created_at FROM transactions WHERE transaction_id = $1`, transactionId)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil, nil
		}
		return nil, nil, err
	}

	var entries []transactionEntryDB
	err = r.db.SelectContext(ctx, &entries, `SELECT account_id, amount, balance_after FROM ledger_entries WHERE transaction_id = $1 ORDER BY id`, transactionId)
	if err != nil {
		return nil, nil, err
	}

	txn, ents := toTransactionDomain(tx, entries)
	return txn, ents, nil
}

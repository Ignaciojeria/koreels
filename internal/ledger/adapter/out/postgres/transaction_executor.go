package postgres

import (
	"context"
	"encoding/json"
	"sort"
	"strings"

	"koreels/internal/ledger/application/ports/out"
	"koreels/internal/ledger/domain/entity"
	domainerrors "koreels/internal/ledger/domain/errors"

	"github.com/Ignaciojeria/ioc"
	"github.com/jmoiron/sqlx"
)

var _ = ioc.Register(NewTransactionExecutor)

type transactionExecutor struct {
	db *sqlx.DB
}

func NewTransactionExecutor(db *sqlx.DB) (out.TransactionExecutor, error) {
	return &transactionExecutor{db: db}, nil
}

func (e *transactionExecutor) Execute(ctx context.Context, tx entity.Transaction, entries []out.TransactionEntryInput) ([]entity.TransactionEntry, error) {
	if len(entries) < 2 {
		return nil, domainerrors.ErrUnbalancedTransaction
	}

	type acc struct {
		id     string
		amount int64
	}
	sorted := make([]acc, len(entries))
	for i, en := range entries {
		sorted[i] = acc{en.AccountId, en.Amount}
	}
	sort.Slice(sorted, func(i, j int) bool { return sorted[i].id < sorted[j].id })

	txx, err := e.db.BeginTxx(ctx, nil)
	if err != nil {
		return nil, err
	}
	defer txx.Rollback()

	accountIDs := make([]string, 0, len(entries))
	seen := make(map[string]bool)
	for _, en := range sorted {
		if !seen[en.id] {
			seen[en.id] = true
			accountIDs = append(accountIDs, en.id)
		}
	}

	type accRow struct {
		AccountId     string `db:"account_id"`
		Balance       int64  `db:"balance"`
		AllowNegative bool   `db:"allow_negative"`
	}
	var accRows []accRow
	query, args, _ := sqlx.In(`SELECT account_id, balance, allow_negative FROM accounts WHERE account_id IN (?) ORDER BY account_id FOR UPDATE`, accountIDs)
	query = txx.Rebind(query)
	if err := txx.SelectContext(ctx, &accRows, query, args...); err != nil {
		return nil, err
	}
	if len(accRows) != len(accountIDs) {
		return nil, domainerrors.ErrAccountNotFound
	}

	balances := make(map[string]int64)
	for _, r := range accRows {
		balances[r.AccountId] = r.Balance
	}
	allowNegative := make(map[string]bool)
	for _, r := range accRows {
		allowNegative[r.AccountId] = r.AllowNegative
	}

	result := make([]entity.TransactionEntry, len(sorted))
	for i, en := range sorted {
		newBalance := balances[en.id] + en.amount
		if !allowNegative[en.id] && newBalance < 0 {
			return nil, domainerrors.ErrInsufficientBalance
		}
		balances[en.id] = newBalance

		result[i] = entity.TransactionEntry{
			TransactionId: tx.TransactionId,
			AccountId:     en.id,
			Amount:        en.amount,
			BalanceAfter:  newBalance,
		}
	}

	metadata, _ := json.Marshal(tx.Metadata)
	_, err = txx.ExecContext(ctx, `INSERT INTO transactions (transaction_id, status, metadata) VALUES ($1, $2, $3)`, tx.TransactionId, tx.Status, metadata)
	if err != nil {
		return nil, err
	}

	for _, r := range result {
		_, err = txx.ExecContext(ctx, `INSERT INTO ledger_entries (transaction_id, account_id, amount, balance_after) VALUES ($1, $2, $3, $4)`, r.TransactionId, r.AccountId, r.Amount, r.BalanceAfter)
		if err != nil {
			return nil, mapTriggerError(err)
		}
	}

	if err := txx.Commit(); err != nil {
		return nil, mapTriggerError(err)
	}

	return result, nil
}

func mapTriggerError(err error) error {
	if err == nil {
		return nil
	}
	s := err.Error()
	if strings.Contains(s, "must sum to 0") {
		return domainerrors.ErrUnbalancedTransaction
	}
	if strings.Contains(s, "does not allow negative balance") {
		return domainerrors.ErrInsufficientBalance
	}
	return err
}

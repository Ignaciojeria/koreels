package postgres

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"time"

	"ledger-service/internal/ledger/application/ports/out"
	"ledger-service/internal/ledger/domain/entity"
	domainerrors "ledger-service/internal/ledger/domain/errors"

	"github.com/Ignaciojeria/ioc"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jmoiron/sqlx"
)

var _ = ioc.Register(NewAccountRepository)

type accountDB struct {
	AccountId     string          `db:"account_id"`
	Type          string          `db:"type"`
	Currency      string          `db:"currency"`
	AllowNegative bool            `db:"allow_negative"`
	Balance       int64           `db:"balance"`
	Metadata      json.RawMessage `db:"metadata"`
	CreatedAt     time.Time       `db:"created_at"`
	UpdatedAt     time.Time       `db:"updated_at"`
}

func toAccountDomain(m accountDB) *entity.Account {
	metadata := make(map[string]string)
	if len(m.Metadata) > 0 {
		_ = json.Unmarshal(m.Metadata, &metadata)
	}
	return &entity.Account{
		AccountId:     m.AccountId,
		Type:          entity.AccountType(m.Type),
		Currency:      m.Currency,
		AllowNegative: m.AllowNegative,
		Balance:       m.Balance,
		Metadata:      metadata,
		CreatedAt:     m.CreatedAt,
		UpdatedAt:     m.UpdatedAt,
	}
}

func fromAccountDomain(e entity.Account) accountDB {
	metadata, _ := json.Marshal(e.Metadata)
	return accountDB{
		AccountId:     e.AccountId,
		Type:          string(e.Type),
		Currency:      e.Currency,
		AllowNegative: e.AllowNegative,
		Metadata:      metadata,
	}
}

type accountRepository struct {
	db *sqlx.DB
}

func NewAccountRepository(db *sqlx.DB) (out.AccountRepository, error) {
	return &accountRepository{db: db}, nil
}

func (r *accountRepository) Create(ctx context.Context, account entity.Account) error {
	m := fromAccountDomain(account)
	_, err := r.db.ExecContext(ctx, `
		INSERT INTO accounts (account_id, type, currency, allow_negative, balance, metadata)
		VALUES ($1, $2, $3, $4, 0, $5)
	`, m.AccountId, m.Type, m.Currency, m.AllowNegative, m.Metadata)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			return domainerrors.ErrAccountAlreadyExists
		}
		return err
	}
	return nil
}

func (r *accountRepository) FindByID(ctx context.Context, accountId string) (*entity.Account, error) {
	var m accountDB
	err := r.db.GetContext(ctx, &m, `
		SELECT account_id, type, currency, allow_negative, balance, metadata, created_at, updated_at
		FROM accounts WHERE account_id = $1
	`, accountId)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, domainerrors.ErrAccountNotFound
		}
		return nil, err
	}
	return toAccountDomain(m), nil
}

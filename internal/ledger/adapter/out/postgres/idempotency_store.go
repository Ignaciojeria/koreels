package postgres

import (
	"context"
	"database/sql"
	"errors"

	"koreels/internal/ledger/application/ports/out"
	"koreels/internal/ledger/domain/entity"

	"github.com/Ignaciojeria/ioc"
	"github.com/jmoiron/sqlx"
)

var _ = ioc.Register(NewIdempotencyStore)

type idempotencyStore struct {
	db *sqlx.DB
}

func NewIdempotencyStore(db *sqlx.DB) (out.IdempotencyStore, error) {
	return &idempotencyStore{db: db}, nil
}

func (s *idempotencyStore) FindByKey(ctx context.Context, key string) (*entity.IdempotencyRecord, error) {
	var rec struct {
		IdempotencyKey  string          `db:"idempotency_key"`
		TransactionId   string          `db:"transaction_id"`
		RequestHash     string          `db:"request_hash"`
		ResponsePayload []byte          `db:"response_payload"`
	}
	err := s.db.GetContext(ctx, &rec, `SELECT idempotency_key, transaction_id, request_hash, response_payload FROM idempotency_records WHERE idempotency_key = $1`, key)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, nil
		}
		return nil, err
	}
	return &entity.IdempotencyRecord{
		IdempotencyKey:  rec.IdempotencyKey,
		TransactionId:  rec.TransactionId,
		RequestHash:    rec.RequestHash,
		ResponsePayload: rec.ResponsePayload,
	}, nil
}

func (s *idempotencyStore) Store(ctx context.Context, record entity.IdempotencyRecord) error {
	_, err := s.db.ExecContext(ctx, `INSERT INTO idempotency_records (idempotency_key, transaction_id, request_hash, response_payload) VALUES ($1, $2, $3, $4)`, record.IdempotencyKey, record.TransactionId, record.RequestHash, record.ResponsePayload)
	return err
}

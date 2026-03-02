CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

CREATE TABLE accounts (
    account_id UUID PRIMARY KEY,
    type VARCHAR(50) NOT NULL,
    currency VARCHAR(20) NOT NULL,
    allow_negative BOOLEAN NOT NULL DEFAULT false,
    balance BIGINT NOT NULL DEFAULT 0,
    metadata JSONB,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE transactions (
    transaction_id UUID PRIMARY KEY,
    status VARCHAR(50) NOT NULL DEFAULT 'COMMITTED',
    metadata JSONB,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE TABLE ledger_entries (
    id BIGSERIAL PRIMARY KEY,
    transaction_id UUID NOT NULL REFERENCES transactions(transaction_id),
    account_id UUID NOT NULL REFERENCES accounts(account_id),
    amount BIGINT NOT NULL,
    balance_after BIGINT NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
) WITH (fillfactor = 100);

CREATE INDEX idx_ledger_entries_account ON ledger_entries(account_id);
CREATE INDEX idx_ledger_entries_transaction ON ledger_entries(transaction_id);
CREATE INDEX idx_ledger_entries_account_id ON ledger_entries(account_id, id DESC);

CREATE OR REPLACE FUNCTION check_ledger_double_entry()
RETURNS TRIGGER AS $$
DECLARE
  tx_id UUID;
  total BIGINT;
BEGIN
  tx_id := COALESCE(NEW.transaction_id, OLD.transaction_id);
  SELECT COALESCE(SUM(amount), 0) INTO total
  FROM ledger_entries
  WHERE transaction_id = tx_id;

  IF total != 0 THEN
    RAISE EXCEPTION 'ledger entries for transaction % must sum to 0, got %', tx_id, total;
  END IF;
  RETURN COALESCE(NEW, OLD);
END;
$$ LANGUAGE plpgsql;

CREATE CONSTRAINT TRIGGER ledger_double_entry_check
  AFTER INSERT OR UPDATE OR DELETE ON ledger_entries
  DEFERRABLE INITIALLY DEFERRED
  FOR EACH ROW
  EXECUTE FUNCTION check_ledger_double_entry();

CREATE OR REPLACE FUNCTION update_account_balance()
RETURNS TRIGGER AS $$
BEGIN
  UPDATE accounts
  SET balance = balance + NEW.amount,
      updated_at = NOW()
  WHERE account_id = NEW.account_id;

  IF EXISTS (
    SELECT 1 FROM accounts
    WHERE account_id = NEW.account_id AND NOT allow_negative AND balance < 0
  ) THEN
    RAISE EXCEPTION 'Account % does not allow negative balance', NEW.account_id;
  END IF;

  RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER trigger_update_balance
  AFTER INSERT ON ledger_entries
  FOR EACH ROW
  EXECUTE FUNCTION update_account_balance();

CREATE TABLE idempotency_records (
    idempotency_key VARCHAR(255) PRIMARY KEY,
    transaction_id UUID NOT NULL,
    request_hash VARCHAR(64) NOT NULL,
    response_payload JSONB,
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

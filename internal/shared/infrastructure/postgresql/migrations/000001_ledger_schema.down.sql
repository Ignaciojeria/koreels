DROP TRIGGER IF EXISTS trigger_update_balance ON ledger_entries;
DROP FUNCTION IF EXISTS update_account_balance();

DROP TRIGGER IF EXISTS ledger_double_entry_check ON ledger_entries;
DROP FUNCTION IF EXISTS check_ledger_double_entry();

DROP TABLE IF EXISTS idempotency_records;
DROP TABLE IF EXISTS ledger_entries;
DROP TABLE IF EXISTS transactions;
DROP TABLE IF EXISTS accounts;

-- Composite index for the LATERAL balance subquery in account searches.
-- Covers: WHERE e.account_id = a.id, then JOIN on t.id with type/cleared_at filtering.
CREATE INDEX IF NOT EXISTS idx_txn_entries_account_transaction
    ON transaction_entries (account_id, transaction_id);

-- Covering index for the transaction join in balance computation.
-- Allows index-only scans for transaction_type and cleared_at lookups.
CREATE INDEX IF NOT EXISTS idx_transactions_type_cleared
    ON transactions (id, transaction_type, cleared_at)
    WHERE deleted_at IS NULL;

-- Partial index for soft-delete filtering on transactions.
-- Speeds up the deleted_at IS NULL predicate in balance queries.
CREATE INDEX IF NOT EXISTS idx_transactions_not_deleted
    ON transactions (id)
    WHERE deleted_at IS NULL;

--- Copyright 2023-2026 Ant Investor Ltd
---
--- Licensed under the Apache License, Version 2.0 (the "License");
--- you may not use this file except in compliance with the License.
--- You may obtain a copy of the License at
---
---      http://www.apache.org/licenses/LICENSE-2.0
---
--- Unless required by applicable law or agreed to in writing, software
--- distributed under the License is distributed on an "AS IS" BASIS,
--- WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
--- See the License for the specific language governing permissions and
--- limitations under the License.

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

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

-- Persist the entry-level currency that was previously gorm:"-" (transient).
-- Without this, every reload lost per-entry currency, ToAPI returned a Money
-- with empty currency_code, and currency-aware zero-sum integrity checks
-- could not be re-evaluated against persisted data.
ALTER TABLE transaction_entries
    ADD COLUMN IF NOT EXISTS currency VARCHAR(10);

-- Backfill from the parent transaction's currency. Validate() enforces that
-- every entry's account currency matches the transaction currency at write
-- time, so this is the authoritative source for existing rows.
UPDATE transaction_entries e
SET currency = t.currency
FROM transactions t
WHERE e.transaction_id = t.id
  AND e.currency IS NULL;

-- Lock the column down so future inserts cannot omit it.
ALTER TABLE transaction_entries
    ALTER COLUMN currency SET NOT NULL;

-- Currency-scoped queries (cash book, per-currency trial balance) need this.
CREATE INDEX IF NOT EXISTS idx_transaction_entries_currency
    ON transaction_entries (currency);

-- Drop the per-entry balance snapshot. It stored the pre-transaction account
-- balance with a name that implied post-transaction, was overwritten on every
-- UpdateTransaction (destroying the historical snapshot), and was never
-- exposed via ToAPI in the first place. Running balances belong in derived
-- statement queries, not denormalised on the entry row.
ALTER TABLE transaction_entries
    DROP COLUMN IF EXISTS balance;

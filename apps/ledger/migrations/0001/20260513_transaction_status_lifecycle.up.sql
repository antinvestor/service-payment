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

-- Status: explicit lifecycle for double-entry transactions.
--   draft    — created but not yet submitted; no balance impact
--   pending  — submitted, awaiting confirmation (async settlement);
--              contributes to un_cleared_balance
--   posted   — confirmed/settled; contributes to balance; immutable
--              except for the auto-transition to 'reversed' when a
--              REVERSAL transaction lands
--   reversed — was posted, then offset by a REVERSAL transaction;
--              entries remain in the books, balance impact cancels
--   voided   — was draft or pending, then cancelled; no balance impact
--   failed   — was pending, then external system rejected; no balance impact
ALTER TABLE transactions
    ADD COLUMN IF NOT EXISTS status VARCHAR(20),
    ADD COLUMN IF NOT EXISTS posted_at TIMESTAMP,
    ADD COLUMN IF NOT EXISTS voided_at TIMESTAMP,
    ADD COLUMN IF NOT EXISTS reversed_transaction_id VARCHAR(50);

-- Backfill: any pre-existing row with a non-zero cleared_at is treated as
-- posted (it was contributing to balance); rows with zero/null cleared_at
-- are treated as pending. posted_at mirrors cleared_at so callers reading
-- the new column see the same moment.
UPDATE transactions
SET status = CASE
        WHEN cleared_at IS NOT NULL AND cleared_at != '0001-01-01 00:00:00'
            THEN 'posted'
        ELSE 'pending'
    END
WHERE status IS NULL;

UPDATE transactions
SET posted_at = cleared_at
WHERE posted_at IS NULL
  AND cleared_at IS NOT NULL
  AND cleared_at != '0001-01-01 00:00:00';

ALTER TABLE transactions
    ALTER COLUMN status SET NOT NULL;

-- Enum guard at the DB layer: prevents application bugs from writing
-- unknown statuses. Names must stay in sync with models.TransactionStatus*.
ALTER TABLE transactions
    ADD CONSTRAINT transactions_status_check
    CHECK (status IN ('draft', 'pending', 'posted', 'reversed', 'voided', 'failed'));

-- Self-FK for reversal lineage. ON DELETE SET NULL keeps the column safe
-- if an original is hard-deleted (shouldn't happen — we use soft delete).
ALTER TABLE transactions
    ADD CONSTRAINT transactions_reversed_transaction_id_fkey
    FOREIGN KEY (reversed_transaction_id)
    REFERENCES transactions (id)
    ON DELETE SET NULL;

-- Status is the new primary filter for balance queries and reports.
CREATE INDEX IF NOT EXISTS idx_transactions_status
    ON transactions (status)
    WHERE deleted_at IS NULL;

-- Reversal lookup ("show me everything reversing transaction X").
CREATE INDEX IF NOT EXISTS idx_transactions_reversed_transaction_id
    ON transactions (reversed_transaction_id)
    WHERE reversed_transaction_id IS NOT NULL;

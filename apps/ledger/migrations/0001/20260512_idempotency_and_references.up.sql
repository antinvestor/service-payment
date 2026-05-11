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

-- IdempotencyKey: caller-supplied dedup token (e.g. "webhook:mtn:abc123").
-- Indexed UNIQUE so concurrent retries with the same key collide at the DB
-- layer, eliminating the need for application-level mutexes or pre-insert
-- lookups. Partial WHERE excludes empty strings so legacy and key-less
-- inserts coexist.
ALTER TABLE transactions
    ADD COLUMN IF NOT EXISTS idempotency_key VARCHAR(120);

-- ExternalRef: provider/external system reference (mobile money transaction
-- id, bank statement id, processor reference). Indexed but not unique —
-- providers can legitimately reissue the same reference for legitimate
-- distinct transactions (refunds vs originals, etc.).
ALTER TABLE transactions
    ADD COLUMN IF NOT EXISTS external_ref VARCHAR(120);

-- Source: classifies the origin of the posting (e.g. "mtn", "airtel",
-- "bank_stanbic", "manual_admin", "internal_billing"). Drives reconciliation
-- queries and source-scoped reports.
ALTER TABLE transactions
    ADD COLUMN IF NOT EXISTS source VARCHAR(50);

-- Partial UNIQUE so multiple rows can carry NULL/empty idempotency_key
-- (PostgreSQL would otherwise treat NULLs as distinct, but the empty-string
-- case still needs guarding).
CREATE UNIQUE INDEX IF NOT EXISTS idx_transactions_idempotency_key
    ON transactions (idempotency_key)
    WHERE idempotency_key IS NOT NULL AND idempotency_key != '';

CREATE INDEX IF NOT EXISTS idx_transactions_external_ref
    ON transactions (external_ref)
    WHERE external_ref IS NOT NULL AND external_ref != '';

CREATE INDEX IF NOT EXISTS idx_transactions_source
    ON transactions (source)
    WHERE source IS NOT NULL AND source != '';

-- Chronological/journal-book queries scan by transacted_at. Partial on
-- deleted_at IS NULL keeps the index focused on live rows. DESC supports
-- the typical "most recent first" report order with no extra sort step.
CREATE INDEX IF NOT EXISTS idx_transactions_transacted_at_desc
    ON transactions (transacted_at DESC)
    WHERE deleted_at IS NULL;

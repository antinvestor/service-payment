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

-- Per-account semantics, finer-grained than LedgerType. AccountType lets
-- contra-, clearing-, suspense- and memo- accounts live under their natural
-- parent ledgers without overloading the type column. NormalBalance keeps
-- DEADCLIC sign rules deterministic per row, ready for accounts whose
-- normal side cannot be inferred from the parent ledger alone (e.g. an
-- allowance for doubtful debts under an asset ledger has credit-normal).
ALTER TABLE accounts
    ADD COLUMN IF NOT EXISTS account_type VARCHAR(50),
    ADD COLUMN IF NOT EXISTS normal_balance VARCHAR(10);

-- Backfill from the existing LedgerType. Note: capital → equity by convention.
UPDATE accounts SET
    account_type = CASE ledger_type
        WHEN 'ASSET'     THEN 'asset'
        WHEN 'LIABILITY' THEN 'liability'
        WHEN 'INCOME'    THEN 'income'
        WHEN 'EXPENSE'   THEN 'expense'
        WHEN 'CAPITAL'   THEN 'equity'
        ELSE NULL
    END,
    normal_balance = CASE ledger_type
        WHEN 'ASSET'     THEN 'debit'
        WHEN 'EXPENSE'   THEN 'debit'
        WHEN 'LIABILITY' THEN 'credit'
        WHEN 'INCOME'    THEN 'credit'
        WHEN 'CAPITAL'   THEN 'credit'
        ELSE NULL
    END
WHERE account_type IS NULL OR normal_balance IS NULL;

ALTER TABLE accounts
    ALTER COLUMN account_type SET NOT NULL,
    ALTER COLUMN normal_balance SET NOT NULL;

ALTER TABLE accounts
    ADD CONSTRAINT accounts_account_type_check
    CHECK (account_type IN (
        'asset', 'liability', 'equity', 'income', 'expense',
        'contra_asset', 'contra_liability', 'contra_income', 'contra_expense',
        'clearing', 'suspense', 'memo'
    ));

ALTER TABLE accounts
    ADD CONSTRAINT accounts_normal_balance_check
    CHECK (normal_balance IN ('debit', 'credit', 'none'));

-- Account-type-scoped report queries (trial balance by classification,
-- contra rollups, clearing exposure) need this.
CREATE INDEX IF NOT EXISTS idx_accounts_account_type
    ON accounts (account_type)
    WHERE deleted_at IS NULL;

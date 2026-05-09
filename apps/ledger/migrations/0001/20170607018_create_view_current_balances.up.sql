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


CREATE MATERIALIZED VIEW account_balances_view AS
WITH balance_summary AS (
    SELECT
        e.account_id,
        t.currency,
        COALESCE(SUM(CASE WHEN t.transaction_type IN ('NORMAL', 'REVERSAL') AND t.cleared_at IS NOT NULL AND t.cleared_at != '0001-01-01 00:00:00' THEN e.amount ELSE 0 END), 0) AS balance,
        COALESCE(SUM(CASE WHEN t.transaction_type IN ('NORMAL', 'REVERSAL') AND (t.cleared_at IS NULL OR t.cleared_at = '0001-01-01 00:00:00') THEN e.amount ELSE 0 END), 0) AS uncleared_balance,
        COALESCE(SUM(CASE WHEN t.transaction_type = 'RESERVATION' THEN e.amount ELSE 0 END), 0) AS reserved_balance
    FROM transaction_entries e
    LEFT JOIN transactions t ON e.transaction_id = t.id
    GROUP BY e.account_id, t.currency
)
SELECT
    a.id as account_id,
    a.currency,
    a.data,
    COALESCE(bs.balance, 0) AS balance,
    COALESCE(bs.uncleared_balance, 0) AS uncleared_balance,
    COALESCE(bs.reserved_balance, 0) AS reserved_balance,
    a.ledger_id,
    a.ledger_type,
    a.created_at,
    a.modified_at,
    a.version,
    a.tenant_id,
    a.partition_id,
    a.access_id,
    a.deleted_at
FROM accounts a
         LEFT JOIN balance_summary bs ON a.id = bs.account_id AND a.currency = bs.currency
    WITH DATA;

CREATE UNIQUE INDEX idx_account_balances_view_unique ON account_balances_view (account_id);
CREATE INDEX idx_account_balances_view_account_id_currency ON account_balances_view (account_id, currency);

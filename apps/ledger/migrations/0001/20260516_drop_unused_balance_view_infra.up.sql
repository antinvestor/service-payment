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

-- Drop the unused materialized view stack. account_balances_view was
-- defined alongside a pg_cron refresh schedule that was always commented
-- out, and no code path reads from it — the LATERAL subquery in
-- accounts.go constAccountQuery is the canonical balance source and now
-- uses Status-based filtering instead of the cleared_at sentinel this
-- view encoded. Removing the dead objects keeps the schema honest and
-- the audit-flagged surface area smaller.
DROP MATERIALIZED VIEW IF EXISTS account_balances_view CASCADE;
DROP FUNCTION IF EXISTS refresh_account_balances_view();
DROP TABLE IF EXISTS account_balances_view_refresh_log;

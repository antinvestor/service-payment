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

-- The forward migration drops infrastructure that was never read by any
-- code path; the down migration recreates the empty refresh-log table
-- and the refresh function shell so the schema shape is restored. The
-- materialized view itself is intentionally not recreated — its
-- definition encoded the cleared_at sentinel that the Status lifecycle
-- has superseded, and resurrecting stale logic would mislead operators.
CREATE TABLE IF NOT EXISTS account_balances_view_refresh_log (
    last_refresh TIMESTAMP
);

CREATE OR REPLACE FUNCTION refresh_account_balances_view() RETURNS void AS $$
BEGIN
    -- Materialized view no longer exists; this function is a no-op shell
    -- restored only so the down migration is reversible at the schema
    -- shape level.
    RETURN;
END;
$$ LANGUAGE plpgsql;

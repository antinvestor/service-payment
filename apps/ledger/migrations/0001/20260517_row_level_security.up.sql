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

-- Row-Level Security: enforce tenancy at the database layer so every
-- query — GORM-built, Raw SQL, ad-hoc psql, admin tool — is automatically
-- scoped to the caller's tenant + partition. Application code stops
-- referencing tenant_id / partition_id manually in WHERE clauses; the
-- repository sets two session variables (app.tenant_id / app.partition_id)
-- at the start of each request, and Postgres policies do the filtering.
--
-- Behaviour:
--   * Settings unset (e.g. migrations, system services): policies fall
--     through and the query sees every row. Required so this migration
--     itself, and any maintenance jobs that legitimately span tenants,
--     keep working.
--   * Settings set: the row's tenant_id / partition_id must match.
--
-- FORCE ROW LEVEL SECURITY applies the policies even when the app user
-- is the table owner — without it Postgres exempts owners from RLS.

CREATE OR REPLACE FUNCTION app_tenancy_matches(
    row_tenant_id text,
    row_partition_id text
) RETURNS boolean AS $$
BEGIN
    RETURN (
        current_setting('app.tenant_id', true) IS NULL
        OR current_setting('app.tenant_id', true) = ''
        OR row_tenant_id = current_setting('app.tenant_id', true)
    ) AND (
        current_setting('app.partition_id', true) IS NULL
        OR current_setting('app.partition_id', true) = ''
        OR row_partition_id = current_setting('app.partition_id', true)
    );
END;
$$ LANGUAGE plpgsql STABLE;

DO $$
DECLARE
    t text;
BEGIN
    FOR t IN SELECT unnest(ARRAY[
        'books',
        'ledgers',
        'accounts',
        'transactions',
        'transaction_entries'
    ])
    LOOP
        EXECUTE format('ALTER TABLE %I ENABLE ROW LEVEL SECURITY', t);
        EXECUTE format('ALTER TABLE %I FORCE ROW LEVEL SECURITY', t);
        EXECUTE format('DROP POLICY IF EXISTS app_tenancy_isolation ON %I', t);
        EXECUTE format(
            'CREATE POLICY app_tenancy_isolation ON %I FOR ALL '
            'USING (app_tenancy_matches(tenant_id, partition_id)) '
            'WITH CHECK (app_tenancy_matches(tenant_id, partition_id))',
            t
        );
    END LOOP;
END$$;

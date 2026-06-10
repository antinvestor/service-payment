// Copyright 2023-2026 Ant Investor Ltd
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

// Package rlsadmin complements frame's frametests/rlstest helper for
// test suites whose tests legitimately run DDL under the unprivileged
// RLS role — re-running migrations, partition retention maintenance or
// fault-injection ALTERs. Plain table privileges (rlstest.GrantAll) are
// not enough for those: Postgres requires table ownership for DDL.
//
// Transferring ownership to the test role is safe for the isolation
// guarantees because frame installs its tenancy policies with FORCE
// ROW LEVEL SECURITY, which applies RLS to the table owner as well.
package rlsadmin

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/pitabwire/frame/frametests/rlstest"

	// Pull in the pgx stdlib driver so database/sql can dial postgres.
	_ "github.com/jackc/pgx/v5/stdlib"
)

// GrantOwnership makes the rlstest role the owner of every table,
// sequence and non-extension function in the public schema and allows
// it to create new relations. Function ownership is needed because
// re-running migrations re-issues CREATE OR REPLACE FUNCTION for
// frame's app_tenancy_matches helper. Must be invoked AFTER migration
// (alongside rlstest.GrantAll) so all objects exist. Idempotent.
func GrantOwnership(ctx context.Context, dsn string) error {
	adminDB, err := sql.Open("pgx", dsn)
	if err != nil {
		return fmt.Errorf("rlsadmin: open admin db: %w", err)
	}
	defer adminDB.Close()

	stmts := []string{
		`GRANT CREATE ON SCHEMA public TO ` + rlstest.Role,
		`DO $$
		DECLARE r record;
		BEGIN
			FOR r IN SELECT tablename AS name FROM pg_tables WHERE schemaname = 'public' LOOP
				EXECUTE format('ALTER TABLE public.%I OWNER TO ` + rlstest.Role + `', r.name);
			END LOOP;
			FOR r IN SELECT sequencename AS name FROM pg_sequences WHERE schemaname = 'public' LOOP
				EXECUTE format('ALTER SEQUENCE public.%I OWNER TO ` + rlstest.Role + `', r.name);
			END LOOP;
			FOR r IN
				SELECT p.oid::regprocedure AS name
				FROM pg_proc p
				JOIN pg_namespace n ON p.pronamespace = n.oid
				WHERE n.nspname = 'public'
				AND NOT EXISTS (
					SELECT 1 FROM pg_depend d
					WHERE d.objid = p.oid AND d.deptype = 'e'
				)
			LOOP
				EXECUTE format('ALTER FUNCTION %s OWNER TO ` + rlstest.Role + `', r.name);
			END LOOP;
		END $$`,
	}
	for _, stmt := range stmts {
		if _, err = adminDB.ExecContext(ctx, stmt); err != nil {
			return fmt.Errorf("rlsadmin: %s: %w", stmt, err)
		}
	}
	return nil
}

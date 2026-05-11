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

-- A Book is an independent accounting scope: one entity's complete set of
-- financial records. Examples: the platform's own book, one book per
-- savings group, one per merchant, one per agent. Each book has its own
-- chart of accounts, its own trial balance and its own balance sheet —
-- and entries from one book must never cross-contaminate another. Without
-- this entity the system has no structural way to enforce that boundary.
CREATE TABLE IF NOT EXISTS books (
    id                  VARCHAR(50) PRIMARY KEY,
    created_at          TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    modified_at         TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    created_by          VARCHAR(50),
    modified_by         VARCHAR(50),
    version             BIGINT DEFAULT 0,
    tenant_id           VARCHAR(50),
    partition_id        VARCHAR(50),
    access_id           VARCHAR(50),
    deleted_at          TIMESTAMPTZ,
    name                VARCHAR(100) NOT NULL,
    type                VARCHAR(50)  NOT NULL,
    -- parent_id supports hierarchical books — an organization can hold many
    -- group books (one per chama / SACCO / branch), each group can hold many
    -- individual member books. Consolidated reports roll an entity's
    -- descendants into a single trial balance / income statement while
    -- per-book reports continue to view a single node. Cross-book POSTING
    -- remains strict: entries cannot cross book boundaries even when the
    -- books share a parent — settlements between scopes are modeled as
    -- two separate transactions linked by external_ref.
    parent_id           VARCHAR(50),
    currency            VARCHAR(10),
    data                JSONB
);

CREATE INDEX IF NOT EXISTS idx_books_base_tenancy
    ON books (tenant_id, partition_id, access_id);

CREATE INDEX IF NOT EXISTS idx_books_data
    ON books USING GIN (data jsonb_path_ops);

CREATE INDEX IF NOT EXISTS idx_books_parent_id
    ON books (parent_id)
    WHERE parent_id IS NOT NULL;

ALTER TABLE books
    ADD CONSTRAINT books_parent_id_fkey
    FOREIGN KEY (parent_id) REFERENCES books (id) ON DELETE SET NULL;

-- book_id columns are nullable for backward compatibility — pre-existing
-- rows without a book continue to work, and the cross-book validation in
-- TransactionBusiness.Validate only triggers when BookID is set on a
-- posting. FK ON DELETE SET NULL prevents orphan references if a book is
-- ever hard-deleted (soft-delete is the standard path).
ALTER TABLE ledgers
    ADD COLUMN IF NOT EXISTS book_id VARCHAR(50);

ALTER TABLE accounts
    ADD COLUMN IF NOT EXISTS book_id VARCHAR(50);

ALTER TABLE transactions
    ADD COLUMN IF NOT EXISTS book_id VARCHAR(50);

ALTER TABLE ledgers
    ADD CONSTRAINT ledgers_book_id_fkey
    FOREIGN KEY (book_id) REFERENCES books (id) ON DELETE SET NULL;

ALTER TABLE accounts
    ADD CONSTRAINT accounts_book_id_fkey
    FOREIGN KEY (book_id) REFERENCES books (id) ON DELETE SET NULL;

ALTER TABLE transactions
    ADD CONSTRAINT transactions_book_id_fkey
    FOREIGN KEY (book_id) REFERENCES books (id) ON DELETE SET NULL;

CREATE INDEX IF NOT EXISTS idx_ledgers_book_id
    ON ledgers (book_id)
    WHERE book_id IS NOT NULL;

CREATE INDEX IF NOT EXISTS idx_accounts_book_id
    ON accounts (book_id)
    WHERE book_id IS NOT NULL;

CREATE INDEX IF NOT EXISTS idx_transactions_book_id
    ON transactions (book_id)
    WHERE book_id IS NOT NULL;

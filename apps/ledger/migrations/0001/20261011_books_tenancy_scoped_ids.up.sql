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

-- A book's id is unique within its tenancy (tenant + partition), not
-- globally. Callers name their books (a savings group's book is named by the
-- group's id), so a global key would let one tenant claim an id first and
-- block another, and would answer "taken" for ids another tenant holds.
--
-- The global primary key is replaced by a unique key over the tenancy and
-- the id. It is a unique constraint rather than a primary key because a
-- primary key forces tenant_id and partition_id NOT NULL, which AutoMigrate
-- relaxes again on every run. Rows always carry both: BaseModel writes them
-- as strings, never NULL.
--
-- References to a book carry the referrer's own tenancy, so a ledger,
-- account or transaction can only name a book of its own tenant and
-- partition.

ALTER TABLE transactions DROP CONSTRAINT IF EXISTS transactions_book_id_fkey;
ALTER TABLE accounts     DROP CONSTRAINT IF EXISTS accounts_book_id_fkey;
ALTER TABLE ledgers      DROP CONSTRAINT IF EXISTS ledgers_book_id_fkey;
ALTER TABLE books        DROP CONSTRAINT IF EXISTS books_parent_id_fkey;

ALTER TABLE books DROP CONSTRAINT IF EXISTS books_pkey;

ALTER TABLE books
    ADD CONSTRAINT books_tenancy_id_key UNIQUE (tenant_id, partition_id, id);

CREATE INDEX IF NOT EXISTS idx_books_id ON books (id);

ALTER TABLE books
    ADD CONSTRAINT books_parent_id_fkey
    FOREIGN KEY (tenant_id, partition_id, parent_id)
    REFERENCES books (tenant_id, partition_id, id) ON DELETE SET NULL (parent_id);

ALTER TABLE ledgers
    ADD CONSTRAINT ledgers_book_id_fkey
    FOREIGN KEY (tenant_id, partition_id, book_id)
    REFERENCES books (tenant_id, partition_id, id) ON DELETE SET NULL (book_id);

ALTER TABLE accounts
    ADD CONSTRAINT accounts_book_id_fkey
    FOREIGN KEY (tenant_id, partition_id, book_id)
    REFERENCES books (tenant_id, partition_id, id) ON DELETE SET NULL (book_id);

ALTER TABLE transactions
    ADD CONSTRAINT transactions_book_id_fkey
    FOREIGN KEY (tenant_id, partition_id, book_id)
    REFERENCES books (tenant_id, partition_id, id) ON DELETE SET NULL (book_id);

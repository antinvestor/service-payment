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

ALTER TABLE transactions DROP CONSTRAINT IF EXISTS transactions_book_id_fkey;
ALTER TABLE accounts     DROP CONSTRAINT IF EXISTS accounts_book_id_fkey;
ALTER TABLE ledgers      DROP CONSTRAINT IF EXISTS ledgers_book_id_fkey;
ALTER TABLE books        DROP CONSTRAINT IF EXISTS books_parent_id_fkey;

DROP INDEX IF EXISTS idx_books_id;
ALTER TABLE books DROP CONSTRAINT IF EXISTS books_tenancy_id_key;
ALTER TABLE books ADD CONSTRAINT books_pkey PRIMARY KEY (id);

ALTER TABLE books
    ADD CONSTRAINT books_parent_id_fkey
    FOREIGN KEY (parent_id) REFERENCES books (id) ON DELETE SET NULL;

ALTER TABLE ledgers
    ADD CONSTRAINT ledgers_book_id_fkey
    FOREIGN KEY (book_id) REFERENCES books (id) ON DELETE SET NULL;

ALTER TABLE accounts
    ADD CONSTRAINT accounts_book_id_fkey
    FOREIGN KEY (book_id) REFERENCES books (id) ON DELETE SET NULL;

ALTER TABLE transactions
    ADD CONSTRAINT transactions_book_id_fkey
    FOREIGN KEY (book_id) REFERENCES books (id) ON DELETE SET NULL;

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

DROP INDEX IF EXISTS idx_transactions_book_id;
DROP INDEX IF EXISTS idx_accounts_book_id;
DROP INDEX IF EXISTS idx_ledgers_book_id;

ALTER TABLE transactions DROP CONSTRAINT IF EXISTS transactions_book_id_fkey;
ALTER TABLE accounts     DROP CONSTRAINT IF EXISTS accounts_book_id_fkey;
ALTER TABLE ledgers      DROP CONSTRAINT IF EXISTS ledgers_book_id_fkey;

ALTER TABLE transactions DROP COLUMN IF EXISTS book_id;
ALTER TABLE accounts     DROP COLUMN IF EXISTS book_id;
ALTER TABLE ledgers      DROP COLUMN IF EXISTS book_id;

DROP INDEX IF EXISTS idx_books_data;
DROP INDEX IF EXISTS idx_books_base_tenancy;

DROP TABLE IF EXISTS books;

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

-- Re-add the dropped balance column. Historical pre-transaction snapshot
-- values cannot be recovered; the column is restored as nullable.
ALTER TABLE transaction_entries
    ADD COLUMN IF NOT EXISTS balance NUMERIC(29,9);

DROP INDEX IF EXISTS idx_transaction_entries_currency;

ALTER TABLE transaction_entries
    DROP COLUMN IF EXISTS currency;

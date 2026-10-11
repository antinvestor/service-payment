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

DROP INDEX IF EXISTS idx_transactions_reversed_transaction_id;
DROP INDEX IF EXISTS idx_transactions_status;

ALTER TABLE transactions
    DROP CONSTRAINT IF EXISTS transactions_reversed_transaction_id_fkey,
    DROP CONSTRAINT IF EXISTS transactions_status_check;

ALTER TABLE transactions
    DROP COLUMN IF EXISTS reversed_transaction_id,
    DROP COLUMN IF EXISTS voided_at,
    DROP COLUMN IF EXISTS posted_at,
    DROP COLUMN IF EXISTS status;

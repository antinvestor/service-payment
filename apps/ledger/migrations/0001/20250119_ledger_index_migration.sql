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


-- Recreate the tsvector column and handle empty jsonb_to_tsv properly
ALTER TABLE ledgers
    ADD COLUMN search_properties tsvector GENERATED ALWAYS AS (
        jsonb_to_tsv(COALESCE(data, '{}'::jsonb))
        ) STORED;

-- Recreate the GIN index
CREATE INDEX idx_ledgers_search_properties ON ledgers USING GIN (search_properties);

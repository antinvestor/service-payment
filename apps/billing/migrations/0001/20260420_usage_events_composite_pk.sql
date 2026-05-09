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

-- Copyright 2023-2026 Ant Investor Ltd
--
-- Licensed under the Apache License, Version 2.0 (the "License").

-- usage_events is promoted to a TimescaleDB hypertable. TimescaleDB
-- requires the time-partition column to participate in every UNIQUE/PRIMARY
-- constraint, so replace the BaseModel-default PK (id) with a composite
-- (id, true_created_at). SDK-batched metering events can arrive late, so
-- the time partition is keyed off the client-captured timestamp.

ALTER TABLE usage_events DROP CONSTRAINT IF EXISTS usage_events_pkey;
ALTER TABLE usage_events ADD PRIMARY KEY (id, true_created_at);

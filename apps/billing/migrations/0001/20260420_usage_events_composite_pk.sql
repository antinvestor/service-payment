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

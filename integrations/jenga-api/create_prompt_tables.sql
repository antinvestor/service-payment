-- Create prompt_statuses table
CREATE TABLE IF NOT EXISTS prompt_statuses (
  id VARCHAR(50) PRIMARY KEY,
  created_at TIMESTAMP,
  modified_at TIMESTAMP,
  version INTEGER,
  tenant_id VARCHAR(50),
  partition_id VARCHAR(50),
  access_id VARCHAR(50),
  deleted_at TIMESTAMP,
  prompt_id VARCHAR(50),
  extra JSONB,
  state INTEGER,
  status INTEGER
);

-- Create prompts table
CREATE TABLE IF NOT EXISTS prompts (
  id VARCHAR(50) PRIMARY KEY,
  created_at TIMESTAMP,
  modified_at TIMESTAMP,
  version INTEGER,
  tenant_id VARCHAR(50),
  partition_id VARCHAR(50),
  access_id VARCHAR(50),
  deleted_at TIMESTAMP,
  source_id VARCHAR(50),
  source_profile_type VARCHAR(50),
  source_contact_id VARCHAR(50),
  recipient_id VARCHAR(50),
  recipient_profile_type VARCHAR(50),
  recipient_contact_id VARCHAR(50),
  amount JSONB,
  date_created VARCHAR(50),
  device_id VARCHAR(50),
  state INTEGER,
  status INTEGER,
  route VARCHAR(50),
  account_id VARCHAR(50),
  extra JSONB
);

-- +migrate Up
CREATE TABLE IF NOT EXISTS user_configs (
    id SERIAL PRIMARY KEY,
    user_id UUID NOT NULL UNIQUE REFERENCES users(id) ON DELETE CASCADE,
    api_endpoint TEXT,
    model_name TEXT,
    api_key BYTEA, -- Store encrypted API key as bytes
    created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

-- Optional: Add index for faster lookups by user_id if needed, although UNIQUE constraint already creates one.
-- CREATE INDEX IF NOT EXISTS idx_user_configs_user_id ON user_configs(user_id);

-- Trigger to update updated_at timestamp
CREATE OR REPLACE FUNCTION update_updated_at_column()
RETURNS TRIGGER AS $$
BEGIN
   NEW.updated_at = NOW();
   RETURN NEW;
END;
$$ language 'plpgsql';

CREATE TRIGGER update_user_configs_updated_at
BEFORE UPDATE ON user_configs
FOR EACH ROW
EXECUTE FUNCTION update_updated_at_column();
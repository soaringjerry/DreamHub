-- Enable UUID generation if not already enabled
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

-- Create the structured_memories table
CREATE TABLE structured_memories (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    key VARCHAR(255) NOT NULL,
    value TEXT NOT NULL,
    created_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP,
    -- Ensure that each user can only have one entry per key
    UNIQUE (user_id, key)
);

-- Create indexes for faster lookups
CREATE INDEX idx_structured_memories_user_id ON structured_memories(user_id);
CREATE INDEX idx_structured_memories_user_id_key ON structured_memories(user_id, key);

-- Optional: Trigger to update updated_at timestamp on row update
CREATE OR REPLACE FUNCTION update_updated_at_column()
RETURNS TRIGGER AS $$
BEGIN
   NEW.updated_at = NOW();
   RETURN NEW;
END;
$$ language 'plpgsql';

CREATE TRIGGER update_structured_memories_updated_at
BEFORE UPDATE ON structured_memories
FOR EACH ROW
EXECUTE FUNCTION update_updated_at_column();
-- Add sync-related fields to existing tables

-- Add version and sync timestamp to conversations
ALTER TABLE conversations 
ADD COLUMN IF NOT EXISTS version INTEGER DEFAULT 1,
ADD COLUMN IF NOT EXISTS sync_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
ADD COLUMN IF NOT EXISTS deleted_at TIMESTAMP WITH TIME ZONE;

-- Add version and sync timestamp to conversation_history (messages)
ALTER TABLE conversation_history
ADD COLUMN IF NOT EXISTS version INTEGER DEFAULT 1,
ADD COLUMN IF NOT EXISTS sync_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
ADD COLUMN IF NOT EXISTS deleted_at TIMESTAMP WITH TIME ZONE;

-- Add version and sync timestamp to structured_memories
ALTER TABLE structured_memories
ADD COLUMN IF NOT EXISTS version INTEGER DEFAULT 1,
ADD COLUMN IF NOT EXISTS sync_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
ADD COLUMN IF NOT EXISTS deleted_at TIMESTAMP WITH TIME ZONE;

-- Add version and sync timestamp to user_configs
ALTER TABLE user_configs
ADD COLUMN IF NOT EXISTS version INTEGER DEFAULT 1,
ADD COLUMN IF NOT EXISTS sync_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP;

-- Create sync_status table to track sync state per device
CREATE TABLE IF NOT EXISTS sync_status (
    id UUID DEFAULT gen_random_uuid() PRIMARY KEY,
    user_id VARCHAR(255) NOT NULL,
    device_id VARCHAR(255) NOT NULL,
    last_sync_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    sync_version BIGINT DEFAULT 0,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(user_id, device_id)
);

-- Create index for sync queries
CREATE INDEX IF NOT EXISTS idx_conversations_sync ON conversations(user_id, sync_at);
CREATE INDEX IF NOT EXISTS idx_conversation_history_sync ON conversation_history(user_id, sync_at);
CREATE INDEX IF NOT EXISTS idx_structured_memories_sync ON structured_memories(user_id, sync_at);
CREATE INDEX IF NOT EXISTS idx_user_configs_sync ON user_configs(user_id, sync_at);

-- Create trigger to update sync_at on modifications
CREATE OR REPLACE FUNCTION update_sync_at()
RETURNS TRIGGER AS $$
BEGIN
    NEW.sync_at = CURRENT_TIMESTAMP;
    NEW.version = COALESCE(OLD.version, 0) + 1;
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

-- Apply trigger to tables
DROP TRIGGER IF EXISTS update_conversations_sync_at ON conversations;
CREATE TRIGGER update_conversations_sync_at
BEFORE UPDATE ON conversations
FOR EACH ROW EXECUTE FUNCTION update_sync_at();

DROP TRIGGER IF EXISTS update_conversation_history_sync_at ON conversation_history;
CREATE TRIGGER update_conversation_history_sync_at
BEFORE UPDATE ON conversation_history
FOR EACH ROW EXECUTE FUNCTION update_sync_at();

DROP TRIGGER IF EXISTS update_structured_memories_sync_at ON structured_memories;
CREATE TRIGGER update_structured_memories_sync_at
BEFORE UPDATE ON structured_memories
FOR EACH ROW EXECUTE FUNCTION update_sync_at();

DROP TRIGGER IF EXISTS update_user_configs_sync_at ON user_configs;
CREATE TRIGGER update_user_configs_sync_at
BEFORE UPDATE ON user_configs
FOR EACH ROW EXECUTE FUNCTION update_sync_at();
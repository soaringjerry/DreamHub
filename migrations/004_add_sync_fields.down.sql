-- Remove sync-related fields and objects

-- Drop triggers
DROP TRIGGER IF EXISTS update_conversations_sync_at ON conversations;
DROP TRIGGER IF EXISTS update_conversation_history_sync_at ON conversation_history;
DROP TRIGGER IF EXISTS update_structured_memories_sync_at ON structured_memories;
DROP TRIGGER IF EXISTS update_user_configs_sync_at ON user_configs;

-- Drop function
DROP FUNCTION IF EXISTS update_sync_at();

-- Drop indexes
DROP INDEX IF EXISTS idx_conversations_sync;
DROP INDEX IF EXISTS idx_conversation_history_sync;
DROP INDEX IF EXISTS idx_structured_memories_sync;
DROP INDEX IF EXISTS idx_user_configs_sync;

-- Drop sync_status table
DROP TABLE IF EXISTS sync_status;

-- Remove columns from tables
ALTER TABLE conversations 
DROP COLUMN IF EXISTS version,
DROP COLUMN IF EXISTS sync_at,
DROP COLUMN IF EXISTS deleted_at;

ALTER TABLE conversation_history
DROP COLUMN IF EXISTS version,
DROP COLUMN IF EXISTS sync_at,
DROP COLUMN IF EXISTS deleted_at;

ALTER TABLE structured_memories
DROP COLUMN IF EXISTS version,
DROP COLUMN IF EXISTS sync_at,
DROP COLUMN IF EXISTS deleted_at;

ALTER TABLE user_configs
DROP COLUMN IF EXISTS version,
DROP COLUMN IF EXISTS sync_at;
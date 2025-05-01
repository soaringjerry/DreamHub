-- Drop the trigger and function first
DROP TRIGGER IF EXISTS update_structured_memories_updated_at ON structured_memories;
DROP FUNCTION IF EXISTS update_updated_at_column();

-- Drop indexes
DROP INDEX IF EXISTS idx_structured_memories_user_id_key;
DROP INDEX IF EXISTS idx_structured_memories_user_id;

-- Drop the table
DROP TABLE IF EXISTS structured_memories;

-- Note: We don't drop the uuid-ossp extension here as other tables might use it.
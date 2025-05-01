-- migrations/001_create_users_table.down.sql

-- Drop the trigger first
DROP TRIGGER IF EXISTS update_users_updated_at ON users;

-- Drop the trigger function
DROP FUNCTION IF EXISTS update_updated_at_column();

-- Drop the index
DROP INDEX IF EXISTS idx_users_username;

-- Drop the users table
DROP TABLE IF EXISTS users;
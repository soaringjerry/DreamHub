-- +migrate Down
DROP TRIGGER IF EXISTS update_user_configs_updated_at ON user_configs;
DROP FUNCTION IF EXISTS update_updated_at_column();
DROP TABLE IF EXISTS user_configs;
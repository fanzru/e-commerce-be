-- Drop indexes
DROP INDEX IF EXISTS idx_refresh_tokens_token;
DROP INDEX IF EXISTS idx_refresh_tokens_user_id;
DROP INDEX IF EXISTS idx_carts_user_id;

-- Remove foreign keys and columns
ALTER TABLE carts DROP COLUMN IF EXISTS user_id;
ALTER TABLE checkouts DROP COLUMN IF EXISTS user_id;

-- Drop tables
DROP TABLE IF EXISTS refresh_tokens;
DROP TABLE IF EXISTS users; 
-- Drop the index on user_id
DROP INDEX IF EXISTS idx_carts_user_id;

-- Drop the foreign key constraint
ALTER TABLE carts DROP CONSTRAINT IF EXISTS fk_carts_user_id;

-- Drop the user_id column
ALTER TABLE carts DROP COLUMN IF EXISTS user_id; 
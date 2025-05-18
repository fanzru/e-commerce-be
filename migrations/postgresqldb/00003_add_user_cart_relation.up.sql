-- Add user_id column to carts table
ALTER TABLE carts ADD COLUMN user_id UUID NULL;

-- Add foreign key constraint to link carts to users
ALTER TABLE carts ADD CONSTRAINT fk_carts_user_id
    FOREIGN KEY (user_id) REFERENCES users (id) ON DELETE SET NULL;

-- Add index on user_id for faster lookup
CREATE INDEX idx_carts_user_id ON carts (user_id) WHERE deleted_at IS NULL;

-- Update existing carts to have NULL user_id
UPDATE carts SET user_id = NULL;

COMMENT ON COLUMN carts.user_id IS 'Reference to the user who owns the cart'; 
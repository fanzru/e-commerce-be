-- Restore cart domain to original state with carts table
BEGIN;

-- 1. Recreate the carts table
CREATE TABLE carts (
    id uuid DEFAULT gen_random_uuid() NOT NULL,
    created_at timestamptz DEFAULT CURRENT_TIMESTAMP NULL,
    updated_at timestamptz DEFAULT CURRENT_TIMESTAMP NULL,
    deleted_at timestamptz NULL,
    user_id uuid NULL,
    CONSTRAINT carts_pkey PRIMARY KEY (id)
);

-- 2. Add foreign key constraints to carts
ALTER TABLE carts ADD CONSTRAINT carts_user_id_fkey FOREIGN KEY (user_id) REFERENCES users(id);
ALTER TABLE carts ADD CONSTRAINT fk_carts_user_id FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE SET NULL;

-- 3. Add index on user_id
CREATE INDEX idx_carts_user_id ON carts USING btree (user_id);

-- 4. Add comment
COMMENT ON COLUMN carts.user_id IS 'Reference to the user who owns the cart';

-- 5. Rename session_id back to cart_id in cart_items
ALTER TABLE cart_items RENAME COLUMN session_id TO cart_id;

-- 6. Recreate unique constraint on cart_id and product_id
ALTER TABLE cart_items DROP CONSTRAINT IF EXISTS cart_items_user_id_product_id_key;
ALTER TABLE cart_items ADD CONSTRAINT cart_items_cart_id_product_id_key UNIQUE (cart_id, product_id);

-- 7. Add foreign key from cart_items to carts
ALTER TABLE cart_items ADD CONSTRAINT cart_items_cart_id_fkey FOREIGN KEY (cart_id) REFERENCES carts(id) ON DELETE CASCADE;

-- 8. Rename session_id back to cart_id in checkouts
ALTER TABLE checkouts RENAME COLUMN session_id TO cart_id;

-- 9. Recreate constraints on checkouts
ALTER TABLE checkouts ADD CONSTRAINT checkouts_cart_id_key UNIQUE (cart_id);
ALTER TABLE checkouts ADD CONSTRAINT checkouts_cart_id_fkey FOREIGN KEY (cart_id) REFERENCES carts(id);

-- 10. Insert cart records for each distinct cart_id in cart_items
INSERT INTO carts (id, user_id, created_at, updated_at)
SELECT DISTINCT ci.cart_id, ci.user_id, CURRENT_TIMESTAMP, CURRENT_TIMESTAMP
FROM cart_items ci;

-- 11. Drop the index on cart_items.user_id
DROP INDEX IF EXISTS idx_cart_items_user_id;

-- 12. Drop user_id column from cart_items
ALTER TABLE cart_items DROP COLUMN user_id;

-- 13. Remove comments
COMMENT ON TABLE cart_items IS NULL;
COMMENT ON COLUMN cart_items.cart_id IS NULL;

COMMIT; 
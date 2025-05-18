-- Refactor cart domain to remove carts table and use direct user_id relationship
BEGIN;

-- 1. First create new user_id column in cart_items
ALTER TABLE cart_items 
ADD COLUMN user_id UUID REFERENCES users(id) ON DELETE CASCADE;

-- 2. Update cart_items to set user_id from carts
UPDATE cart_items ci
SET user_id = c.user_id
FROM carts c
WHERE ci.cart_id = c.id AND c.user_id IS NOT NULL;

-- 3. Create index on user_id in cart_items
CREATE INDEX idx_cart_items_user_id ON cart_items(user_id);

-- 4. Modify checkout table to remove cart_id dependency
-- First add a temporary session_id column to group cart items that were 
-- previously associated with the same cart
ALTER TABLE checkouts ADD COLUMN session_id UUID;

-- Copy cart_id to session_id
UPDATE checkouts SET session_id = cart_id;

-- 5. Drop cart_id foreign key constraint in checkouts
ALTER TABLE checkouts DROP CONSTRAINT checkouts_cart_id_fkey;

-- 6. Drop cart_id unique constraint in checkouts
ALTER TABLE checkouts DROP CONSTRAINT checkouts_cart_id_key;

-- 7. Rename the cart_id column to session_id in the constraint
ALTER TABLE checkouts RENAME COLUMN cart_id TO session_id;

-- 8. Add index on session_id
CREATE INDEX idx_checkouts_session_id ON checkouts(session_id);

-- 9. Drop foreign key constraints on cart_items
ALTER TABLE cart_items DROP CONSTRAINT cart_items_cart_id_fkey;
ALTER TABLE cart_items DROP CONSTRAINT cart_items_cart_id_product_id_key;

-- 10. Create new unique constraint for user_id and product_id
ALTER TABLE cart_items ADD CONSTRAINT cart_items_user_id_product_id_key UNIQUE (user_id, product_id);

-- 11. Rename cart_id to session_id to maintain data structure
ALTER TABLE cart_items RENAME COLUMN cart_id TO session_id;

-- 12. Finally, drop the carts table
DROP TABLE carts;

-- 13. Add comment explaining the change
COMMENT ON TABLE cart_items IS 'Stores items in a user''s cart with direct user_id relationship';
COMMENT ON COLUMN cart_items.session_id IS 'Legacy identifier from cart, now used as session grouping';
COMMENT ON COLUMN cart_items.user_id IS 'User who owns this cart item';

COMMIT; 
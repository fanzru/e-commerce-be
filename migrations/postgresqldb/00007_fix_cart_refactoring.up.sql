-- Migration to fix cart refactoring completely
BEGIN;

-- 1. Add user_id column to cart_items if it doesn't exist
DO $$
BEGIN
    IF NOT EXISTS (SELECT FROM information_schema.columns 
                  WHERE table_name = 'cart_items' 
                  AND column_name = 'user_id') THEN
        
        -- Add user_id column to cart_items
        ALTER TABLE cart_items 
        ADD COLUMN user_id UUID REFERENCES users(id) ON DELETE CASCADE;
        
        -- Update cart_items to set user_id from carts
        UPDATE cart_items ci
        SET user_id = c.user_id
        FROM carts c
        WHERE ci.cart_id = c.id AND c.user_id IS NOT NULL;
        
        -- Create index on user_id in cart_items
        CREATE INDEX idx_cart_items_user_id ON cart_items(user_id);
    END IF;
END $$;

-- 2. Modify checkouts table to work without cart_id
-- First ensure we have a temporary column for migration
DO $$
BEGIN
    IF EXISTS (SELECT FROM information_schema.columns 
              WHERE table_name = 'checkouts' 
              AND column_name = 'cart_id') 
    AND NOT EXISTS (SELECT FROM information_schema.columns 
                   WHERE table_name = 'checkouts' 
                   AND column_name = 'session_id') THEN
        
        -- First add session_id column
        ALTER TABLE checkouts ADD COLUMN session_id UUID;
        
        -- Copy cart_id to session_id (for temporary reference)
        UPDATE checkouts SET session_id = cart_id;
        
        -- Create index on session_id (if needed for transition)
        CREATE INDEX idx_checkouts_session_id ON checkouts(session_id);
        
        -- Now we can safely remove the cart dependencies
        
        -- Drop cart_id foreign key constraint in checkouts
        ALTER TABLE checkouts DROP CONSTRAINT IF EXISTS checkouts_cart_id_fkey;
        
        -- Drop cart_id unique constraint in checkouts
        ALTER TABLE checkouts DROP CONSTRAINT IF EXISTS checkouts_cart_id_key;
        
        -- Drop cart_id column as it's no longer needed
        ALTER TABLE checkouts DROP COLUMN cart_id;
    END IF;
END $$;

-- 3. Update cart_items constraints to use user_id directly
DO $$
BEGIN
    -- Create new unique constraint for user_id and product_id if it doesn't exist
    IF NOT EXISTS (SELECT 1 FROM pg_constraint WHERE conname = 'cart_items_user_id_product_id_key') 
       AND EXISTS (SELECT FROM information_schema.columns 
                  WHERE table_name = 'cart_items' 
                  AND column_name = 'user_id') THEN
        
        -- Drop existing cart_id constraints
        ALTER TABLE cart_items DROP CONSTRAINT IF EXISTS cart_items_cart_id_fkey;
        ALTER TABLE cart_items DROP CONSTRAINT IF EXISTS cart_items_cart_id_product_id_key;
        
        -- Add new constraint
        ALTER TABLE cart_items ADD CONSTRAINT cart_items_user_id_product_id_key UNIQUE (user_id, product_id);
    END IF;
END $$;

-- 4. Finally, fully remove cart_id column from cart_items if we have user_id
DO $$
BEGIN
    IF EXISTS (SELECT FROM information_schema.columns 
              WHERE table_name = 'cart_items' 
              AND column_name = 'cart_id')
       AND EXISTS (SELECT FROM information_schema.columns 
                  WHERE table_name = 'cart_items' 
                  AND column_name = 'user_id') THEN
                  
        -- Rename cart_id to session_id as a first step (if not already done)
        ALTER TABLE cart_items RENAME COLUMN cart_id TO session_id;
        
        -- Then remove session_id
        ALTER TABLE cart_items DROP COLUMN session_id;
    END IF;
END $$;

-- 5. Drop the carts table if it still exists
DROP TABLE IF EXISTS carts;

-- 6. Clean up any remaining session_id references
DO $$
BEGIN
    -- Remove session_id from checkout if it exists
    IF EXISTS (SELECT FROM information_schema.columns 
              WHERE table_name = 'checkouts' 
              AND column_name = 'session_id') THEN
        
        -- Drop index on session_id if it exists
        IF EXISTS (SELECT 1 FROM pg_indexes WHERE indexname = 'idx_checkouts_session_id') THEN
            DROP INDEX idx_checkouts_session_id;
        END IF;
        
        -- Remove session_id column from checkouts
        ALTER TABLE checkouts DROP COLUMN session_id;
    END IF;
END $$;

-- 7. Update comments to reflect the new structure
COMMENT ON TABLE cart_items IS 'Stores items in a user''s cart with direct user_id relationship';
COMMENT ON COLUMN cart_items.user_id IS 'User who owns this cart item';

-- 8. Make sure we have proper indexes
DO $$
BEGIN
    -- Create index on user_id in checkouts if it doesn't exist
    IF NOT EXISTS (SELECT 1 FROM pg_indexes WHERE indexname = 'idx_checkouts_user_id') THEN
        CREATE INDEX idx_checkouts_user_id ON checkouts(user_id);
    END IF;
END $$;

COMMIT; 
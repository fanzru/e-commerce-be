-- Migration to completely remove session_id from cart system
BEGIN;

-- 1. First check if the column exists
DO $$
BEGIN
    IF EXISTS (SELECT FROM information_schema.columns 
                WHERE table_name = 'cart_items' 
                AND column_name = 'session_id') THEN
        
        -- Remove session_id column from cart_items as it's no longer needed
        ALTER TABLE cart_items DROP COLUMN session_id;
        
    END IF;
END $$;

-- 2. Update comments to reflect the simplified structure
COMMENT ON TABLE cart_items IS 'Stores items in a user''s cart with direct user_id relationship';
COMMENT ON COLUMN cart_items.user_id IS 'User who owns this cart item';

-- 3. Check if we need to modify the checkouts table
DO $$
BEGIN
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

-- 4. Create new index on user_id in checkouts if it doesn't exist
DO $$
BEGIN
    IF NOT EXISTS (SELECT 1 FROM pg_indexes WHERE indexname = 'idx_checkouts_user_id') THEN
        CREATE INDEX idx_checkouts_user_id ON checkouts(user_id);
    END IF;
END $$;

COMMIT; 
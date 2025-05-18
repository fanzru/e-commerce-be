-- Migration to restore session_id to cart system
BEGIN;

-- 1. Add session_id to cart_items if it doesn't exist
DO $$
BEGIN
    IF NOT EXISTS (SELECT FROM information_schema.columns 
                WHERE table_name = 'cart_items' 
                AND column_name = 'session_id') THEN
        
        -- Add session_id column to cart_items
        ALTER TABLE cart_items ADD COLUMN session_id UUID;
        
    END IF;
END $$;

-- 2. Add session_id to checkouts if it doesn't exist
DO $$
BEGIN
    IF NOT EXISTS (SELECT FROM information_schema.columns 
                WHERE table_name = 'checkouts' 
                AND column_name = 'session_id') THEN
        
        -- Add session_id column to checkouts
        ALTER TABLE checkouts ADD COLUMN session_id UUID;
        
        -- Create index on session_id
        CREATE INDEX idx_checkouts_session_id ON checkouts(session_id);
        
    END IF;
END $$;

-- 3. Update comments to reflect the restored structure
COMMENT ON TABLE cart_items IS 'Stores items in a user''s cart with direct user_id relationship';
COMMENT ON COLUMN cart_items.session_id IS 'Legacy identifier from cart, now used as session grouping';
COMMENT ON COLUMN cart_items.user_id IS 'User who owns this cart item';

COMMIT; 
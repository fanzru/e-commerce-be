-- Migration to revert cart refactoring and restore the original cart system
BEGIN;

-- 1. Recreate the carts table if it doesn't exist
DO $$
BEGIN
    IF NOT EXISTS (SELECT FROM information_schema.tables 
                   WHERE table_name = 'carts') THEN
        
        -- Recreate the carts table
        CREATE TABLE carts (
            id uuid DEFAULT gen_random_uuid() NOT NULL,
            created_at timestamptz DEFAULT CURRENT_TIMESTAMP NULL,
            updated_at timestamptz DEFAULT CURRENT_TIMESTAMP NULL,
            deleted_at timestamptz NULL,
            user_id uuid NULL,
            CONSTRAINT carts_pkey PRIMARY KEY (id)
        );
        
        -- Add foreign key constraints to carts
        ALTER TABLE carts ADD CONSTRAINT carts_user_id_fkey FOREIGN KEY (user_id) REFERENCES users(id);
        ALTER TABLE carts ADD CONSTRAINT fk_carts_user_id FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE SET NULL;
        
        -- Add index on user_id
        CREATE INDEX idx_carts_user_id ON carts USING btree (user_id);
        
        -- Add comment
        COMMENT ON COLUMN carts.user_id IS 'Reference to the user who owns the cart';
    END IF;
END $$;

-- 2. Add cart_id column to cart_items if it doesn't exist
DO $$
BEGIN
    IF NOT EXISTS (SELECT FROM information_schema.columns 
                   WHERE table_name = 'cart_items' 
                   AND column_name = 'cart_id') THEN
        
        -- Add cart_id column
        ALTER TABLE cart_items ADD COLUMN cart_id uuid;
        
        -- For existing cart items, create one cart per user and assign items to it
        WITH new_carts AS (
            INSERT INTO carts (user_id)
            SELECT DISTINCT user_id FROM cart_items
            WHERE user_id IS NOT NULL
            RETURNING id, user_id
        )
        UPDATE cart_items ci
        SET cart_id = nc.id
        FROM new_carts nc
        WHERE ci.user_id = nc.user_id;
        
        -- Make cart_id not null after migration
        ALTER TABLE cart_items ALTER COLUMN cart_id SET NOT NULL;
        
        -- Create constraint for cart_id and product_id
        ALTER TABLE cart_items ADD CONSTRAINT cart_items_cart_id_product_id_key UNIQUE (cart_id, product_id);
        
        -- Add foreign key from cart_items to carts
        ALTER TABLE cart_items ADD CONSTRAINT cart_items_cart_id_fkey FOREIGN KEY (cart_id) REFERENCES carts(id) ON DELETE CASCADE;
    END IF;
END $$;

-- 3. Modify checkouts table to use cart_id
DO $$
BEGIN
    IF NOT EXISTS (SELECT FROM information_schema.columns 
                  WHERE table_name = 'checkouts' 
                  AND column_name = 'cart_id') THEN
        
        -- Add cart_id column to checkouts
        ALTER TABLE checkouts ADD COLUMN cart_id uuid;
        
        -- For each existing checkout, create a cart and link it
        WITH checkout_carts AS (
            INSERT INTO carts (user_id)
            SELECT user_id FROM checkouts
            WHERE user_id IS NOT NULL
            RETURNING id, user_id
        )
        UPDATE checkouts c
        SET cart_id = cc.id
        FROM checkout_carts cc
        WHERE c.user_id = cc.user_id;
        
        -- Make cart_id not null
        ALTER TABLE checkouts ALTER COLUMN cart_id SET NOT NULL;
        
        -- Add constraints for cart_id
        ALTER TABLE checkouts ADD CONSTRAINT checkouts_cart_id_key UNIQUE (cart_id);
        ALTER TABLE checkouts ADD CONSTRAINT checkouts_cart_id_fkey FOREIGN KEY (cart_id) REFERENCES carts(id);
    END IF;
END $$;

-- 4. Update comments to reflect the original structure
COMMENT ON TABLE cart_items IS NULL;
COMMENT ON COLUMN cart_items.cart_id IS NULL;

COMMIT; 
-- Migration to remove product detail columns from cart_items
BEGIN;

-- First check if the columns exist
DO $$
BEGIN
    -- Drop product_sku column if it exists
    IF EXISTS (SELECT FROM information_schema.columns 
               WHERE table_name = 'cart_items' 
               AND column_name = 'product_sku') THEN
        ALTER TABLE cart_items DROP COLUMN product_sku;
    END IF;

    -- Drop product_name column if it exists
    IF EXISTS (SELECT FROM information_schema.columns 
               WHERE table_name = 'cart_items' 
               AND column_name = 'product_name') THEN
        ALTER TABLE cart_items DROP COLUMN product_name;
    END IF;

    -- Drop unit_price column if it exists
    IF EXISTS (SELECT FROM information_schema.columns 
               WHERE table_name = 'cart_items' 
               AND column_name = 'unit_price') THEN
        ALTER TABLE cart_items DROP COLUMN unit_price;
    END IF;
END $$;

-- Add foreign key constraint to products if it doesn't exist already
DO $$
BEGIN
    IF NOT EXISTS (SELECT 1 FROM pg_constraint WHERE conname = 'cart_items_product_id_fkey') THEN
        ALTER TABLE cart_items 
        ADD CONSTRAINT cart_items_product_id_fkey 
        FOREIGN KEY (product_id) 
        REFERENCES products(id);
    END IF;
END $$;

-- Update comment to reflect the change
COMMENT ON TABLE cart_items IS 'Stores items in a user''s cart with direct reference to products';

COMMIT; 
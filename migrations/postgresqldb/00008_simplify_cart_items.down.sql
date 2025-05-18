-- Reverse migration to add back product detail columns to cart_items
BEGIN;

-- Add columns if they don't exist
DO $$
BEGIN
    -- Add product_sku column if it doesn't exist
    IF NOT EXISTS (SELECT FROM information_schema.columns 
                  WHERE table_name = 'cart_items' 
                  AND column_name = 'product_sku') THEN
        ALTER TABLE cart_items ADD COLUMN product_sku VARCHAR;
    END IF;

    -- Add product_name column if it doesn't exist
    IF NOT EXISTS (SELECT FROM information_schema.columns 
                  WHERE table_name = 'cart_items' 
                  AND column_name = 'product_name') THEN
        ALTER TABLE cart_items ADD COLUMN product_name VARCHAR;
    END IF;

    -- Add unit_price column if it doesn't exist
    IF NOT EXISTS (SELECT FROM information_schema.columns 
                  WHERE table_name = 'cart_items' 
                  AND column_name = 'unit_price') THEN
        ALTER TABLE cart_items ADD COLUMN unit_price DECIMAL(10, 2);
    END IF;
END $$;

-- Populate the columns with data from products table
UPDATE cart_items ci
SET 
    product_sku = p.sku,
    product_name = p.name,
    unit_price = p.price
FROM products p
WHERE ci.product_id = p.id AND ci.deleted_at IS NULL;

-- Remove foreign key constraint if it exists
ALTER TABLE cart_items DROP CONSTRAINT IF EXISTS cart_items_product_id_fkey;

-- Update comment
COMMENT ON TABLE cart_items IS 'Stores items in a user''s cart with product details';

COMMIT; 
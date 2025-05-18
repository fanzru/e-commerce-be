-- Refactor checkout to order system
BEGIN;

-- First, add missing columns to checkouts table
ALTER TABLE checkouts 
ADD COLUMN IF NOT EXISTS user_id UUID REFERENCES users(id),
ADD COLUMN IF NOT EXISTS payment_status VARCHAR(50) DEFAULT 'PENDING' NOT NULL,
ADD COLUMN IF NOT EXISTS payment_method VARCHAR(50),
ADD COLUMN IF NOT EXISTS payment_reference VARCHAR(255),
ADD COLUMN IF NOT EXISTS notes TEXT,
ADD COLUMN IF NOT EXISTS status VARCHAR(50) DEFAULT 'CREATED' NOT NULL,
ADD COLUMN IF NOT EXISTS completed_at TIMESTAMP WITH TIME ZONE;

-- Create index on user_id for faster lookups
CREATE INDEX IF NOT EXISTS idx_checkouts_user_id ON checkouts(user_id);
CREATE INDEX IF NOT EXISTS idx_checkouts_payment_status ON checkouts(payment_status);
CREATE INDEX IF NOT EXISTS idx_checkouts_status ON checkouts(status);

-- Update existing checkouts to set user_id from carts table
UPDATE checkouts c
SET user_id = (SELECT user_id FROM carts WHERE id = c.cart_id)
WHERE c.user_id IS NULL AND EXISTS (SELECT 1 FROM carts WHERE id = c.cart_id AND user_id IS NOT NULL);

-- Create a view for order history to make querying easier
CREATE OR REPLACE VIEW order_history AS
SELECT 
    c.id as order_id,
    c.user_id,
    c.subtotal,
    c.total_discount,
    c.total,
    c.payment_status,
    c.payment_method,
    c.status,
    c.created_at as order_date,
    c.completed_at,
    COUNT(ci.id) as item_count,
    COALESCE(pa.promotion_count, 0) as promotion_count
FROM 
    checkouts c
LEFT JOIN 
    checkout_items ci ON c.id = ci.checkout_id
LEFT JOIN 
    (SELECT checkout_id, COUNT(*) as promotion_count FROM promotion_applied GROUP BY checkout_id) pa 
    ON c.id = pa.checkout_id
GROUP BY 
    c.id, c.user_id, c.subtotal, c.total_discount, c.total, 
    c.payment_status, c.payment_method, c.status, c.created_at, c.completed_at, pa.promotion_count;

-- Add comment for better documentation
COMMENT ON TABLE checkouts IS 'Stores order information, representing completed checkouts with payment status';
COMMENT ON COLUMN checkouts.payment_status IS 'Payment status: PENDING, PAID, FAILED, REFUNDED';
COMMENT ON COLUMN checkouts.status IS 'Order status: CREATED, PROCESSING, SHIPPED, DELIVERED, CANCELLED';

COMMIT; 
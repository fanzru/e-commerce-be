-- Setup database tables
CREATE EXTENSION IF NOT EXISTS "pgcrypto";

-- Products table
CREATE TABLE IF NOT EXISTS products (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    sku VARCHAR(50) NOT NULL UNIQUE,
    name VARCHAR(255) NOT NULL,
    price DECIMAL(10, 2) NOT NULL,
    inventory INT NOT NULL DEFAULT 0,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP WITH TIME ZONE NULL
);

-- Carts table
CREATE TABLE IF NOT EXISTS carts (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP WITH TIME ZONE NULL
);

-- Cart items table
CREATE TABLE IF NOT EXISTS cart_items (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    cart_id UUID NOT NULL,
    product_id UUID NOT NULL,
    product_sku VARCHAR(50) NOT NULL,
    product_name VARCHAR(255) NOT NULL,
    unit_price DECIMAL(10, 2) NOT NULL,
    quantity INT NOT NULL DEFAULT 1,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP WITH TIME ZONE NULL,
    FOREIGN KEY (cart_id) REFERENCES carts(id) ON DELETE CASCADE,
    FOREIGN KEY (product_id) REFERENCES products(id) ON DELETE RESTRICT,
    UNIQUE(cart_id, product_id)
);

-- Promotions table
CREATE TABLE IF NOT EXISTS promotions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    type VARCHAR(50) NOT NULL,
    description TEXT NOT NULL,
    rule JSONB NOT NULL,
    active BOOLEAN NOT NULL DEFAULT TRUE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP WITH TIME ZONE NULL
);

-- Checkouts table
CREATE TABLE IF NOT EXISTS checkouts (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    cart_id UUID UNIQUE NOT NULL REFERENCES carts(id),
    subtotal DECIMAL(10, 2) NOT NULL,
    total_discount DECIMAL(10, 2) NOT NULL,
    total DECIMAL(10, 2) NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- Checkout items table
CREATE TABLE IF NOT EXISTS checkout_items (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    checkout_id UUID NOT NULL REFERENCES checkouts(id),
    product_id UUID NOT NULL REFERENCES products(id),
    product_sku VARCHAR(50) NOT NULL,
    product_name VARCHAR(255) NOT NULL,
    quantity INT NOT NULL,
    unit_price DECIMAL(10, 2) NOT NULL,
    subtotal DECIMAL(10, 2) NOT NULL,
    discount DECIMAL(10, 2) NOT NULL,
    total DECIMAL(10, 2) NOT NULL,
    FOREIGN KEY (checkout_id) REFERENCES checkouts(id) ON DELETE CASCADE
);

-- Promotion applied table
CREATE TABLE IF NOT EXISTS promotion_applied (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    checkout_id UUID NOT NULL REFERENCES checkouts(id),
    promotion_id UUID NOT NULL REFERENCES promotions(id),
    description TEXT NOT NULL,
    discount DECIMAL(10, 2) NOT NULL,
    FOREIGN KEY (checkout_id) REFERENCES checkouts(id) ON DELETE CASCADE
);

-- Insert seed data

-- Seed products
INSERT INTO products (sku, name, price, inventory) VALUES
    ('120P90', 'Google Home', 49.99, 10),
    ('43N23P', 'MacBook Pro', 5399.99, 5),
    ('A304SD', 'Alexa Speaker', 109.50, 10),
    ('234234', 'Raspberry Pi B', 30.00, 2)
ON CONFLICT (sku) DO NOTHING;

-- Seed promotions
INSERT INTO promotions (type, description, rule) VALUES
    ('BUY_ONE_GET_ONE_FREE', 'Each MacBook Pro purchase comes with a free Raspberry Pi B', 
    '{"trigger_sku": "43N23P", "free_sku": "234234", "trigger_quantity": 1, "free_quantity": 1}'::JSONB),
    
    ('BUY_3_PAY_2', 'When you buy 3 Google Home devices, you only pay for 2',
    '{"sku": "120P90", "min_quantity": 3, "paid_quantity_divisor": 2, "free_quantity_divisor": 1}'::JSONB),
    
    ('BULK_DISCOUNT', 'Buying more than 3 Alexa Speakers gives a 10% discount on all Alexa Speakers',
    '{"sku": "A304SD", "min_quantity": 4, "discount_percentage": 10}'::JSONB)
ON CONFLICT DO NOTHING; 
-- Create users table
CREATE TABLE IF NOT EXISTS users (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    email VARCHAR(255) NOT NULL UNIQUE,
    password VARCHAR(255) NOT NULL,
    name VARCHAR(255) NOT NULL,
    role VARCHAR(20) NOT NULL DEFAULT 'customer',
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP WITH TIME ZONE NULL
);

-- Add user_id column to carts table
ALTER TABLE carts 
ADD COLUMN user_id UUID NULL REFERENCES users(id);

-- Add index for faster lookup of user's carts
CREATE INDEX idx_carts_user_id ON carts(user_id);

-- Add user_id to checkouts table for direct user reference
ALTER TABLE checkouts
ADD COLUMN user_id UUID NULL REFERENCES users(id);

-- Create refresh tokens table for authentication
CREATE TABLE IF NOT EXISTS refresh_tokens (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    token VARCHAR(255) NOT NULL UNIQUE,
    expires_at TIMESTAMP WITH TIME ZONE NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP
);

-- Add index for faster token lookups
CREATE INDEX idx_refresh_tokens_token ON refresh_tokens(token);
CREATE INDEX idx_refresh_tokens_user_id ON refresh_tokens(user_id);

-- Insert sample users (password is 'password' hashed)
INSERT INTO users (email, password, name, role) VALUES
    ('admin@fanzru.devdev', '$2a$10$3QxDjD1ylgPnRgQLhBrTaOGzfLiXYee6cI8FOuuLrmZlHIcaZnQFi', 'Admin User', 'admin'),
    ('customer@fanzru.com', '$2a$10$3QxDjD1ylgPnRgQLhBrTaOGzfLiXYee6cI8FOuuLrmZlHIcaZnQFi', 'Regular User', 'customer')
ON CONFLICT (email) DO NOTHING; 
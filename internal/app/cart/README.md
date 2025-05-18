# Cart Domain

## Overview

The Cart domain provides shopping cart functionality for the e-commerce application. It allows users to:

- Add products to their cart
- Update quantities of products in their cart
- Remove items from their cart
- View their cart contents
- Clear their entire cart

## Domain Model

The cart domain uses a simple model:

- **Cart**: A virtual collection of cart items belonging to a user
- **CartItem**: A single item in a user's cart with product details and quantity

## Implementation Details

### Simplified Data Model

The cart system uses a simplified data model that directly associates cart items with users through the `user_id` field. This approach eliminates the need for a separate `carts` table, reducing complexity and improving performance.

### Database Schema

The main table is `cart_items`:

```sql
CREATE TABLE cart_items (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    product_id UUID NOT NULL REFERENCES products(id) ON DELETE RESTRICT,
    quantity INT NOT NULL DEFAULT 1,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP WITH TIME ZONE NULL,
    CONSTRAINT cart_items_user_id_product_id_key UNIQUE (user_id, product_id)
);
```

## Refactoring History

The cart domain has undergone significant refactoring:

1. **Initial Design**: Used a separate `carts` table with `cart_id` references in `cart_items`
2. **First Refactoring**: Eliminated the `carts` table but retained `session_id` (renamed from `cart_id`) for backward compatibility
3. **Current Design**: Completely removed `session_id`, using only `user_id` to associate cart items with users

## API Endpoints

The cart functionality is exposed through REST API endpoints:

- `GET /api/v1/carts/me` - Get the current user's cart
- `POST /api/v1/carts/me` - Add an item to the user's cart
- `PUT /api/v1/carts/me/items/{itemId}` - Update item quantity
- `DELETE /api/v1/carts/me/items/{itemId}` - Remove an item from the cart
- `DELETE /api/v1/carts/me/clear` - Clear the user's cart

## Architecture

The cart domain follows clean architecture principles:

- **Entity Layer**: Contains the domain models (`Cart` and `CartItem`)
- **Repository Layer**: Provides data access abstraction
- **Use Case Layer**: Implements business logic
- **HTTP Layer**: Exposes REST API endpoints

## Dependencies

The cart domain depends on:

- Product domain (for product information)
- User domain (for user authentication and identification)

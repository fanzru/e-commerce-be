-- Drop tables in reverse order of creation to handle dependencies

DROP TABLE IF EXISTS promotion_applied;
DROP TABLE IF EXISTS checkout_items;
DROP TABLE IF EXISTS checkouts;
DROP TABLE IF EXISTS cart_items;
DROP TABLE IF EXISTS promotions;
DROP TABLE IF EXISTS carts;
DROP TABLE IF EXISTS products; 
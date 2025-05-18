# E-Commerce Backend

A Go-based backend service for an e-commerce platform using Domain-Driven Design (DDD) principles. This service handles products, shopping carts, promotions, and checkout processes.

## Architecture

The project follows a clean architecture approach with Domain-Driven Design:

- `api/` - OpenAPI specifications
- `cmd/` - Application entry points
- `internal/` - Internal application code, organized by domains:
  - `app/` - Domain logic
    - `product/` - Product domain
    - `cart/` - Cart domain
    - `promotion/` - Promotion domain
    - `checkout/` - Checkout domain
  - `common/` - Common utilities
  - `infrastructure/` - Infrastructure concerns
  - `server/` - HTTP server setup
- `migrations/` - Database migrations
- `docs/` - API documentation

## Features

- Product management
- Cart management with add/remove/update items
- Promotion rules:
  - Buy one get one free (MacBook Pro comes with a free Raspberry Pi B)
  - Buy 3 pay for 2 (3 Google Home devices for the price of 2)
  - Bulk discounts (10% off when buying more than 3 Alexa Speakers)
- Checkout process that applies promotions

## Database Migrations

Database setup and seed data are handled by a single migration:

- **00001_setup_first_app.up.sql** - Creates all tables and inserts seed data
- **00001_setup_first_app.down.sql** - Drops all tables in reverse order

## Getting Started

### Prerequisites

- Go 1.22 or higher
- PostgreSQL 15 or higher
- [golang-migrate](https://github.com/golang-migrate/migrate) for database migrations

### Setup

1. Clone the repository
2. Create a `.env` file based on `.env.sample`
3. Create the database:

   ```sql
   CREATE DATABASE ecommerce;
   ```

4. Run database migrations:

   ```bash
   make migrate-up
   ```

5. Start the server:

   ```bash
   make run
   ```

6. Access the Swagger UI at http://localhost:8080/swagger/

### Development Commands

```bash
# Run the application
make run

# Generate HTTP handlers from OpenAPI specs
make gen-http

# Generate Swagger documentation
make gen-swagger

# Run database migrations
make migrate-up

# Rollback database migrations
make migrate-down

# Create a new migration
make migrate-create

# Run tests
make test

# See all available commands
make help
```

## API Documentation

API documentation is available via Swagger UI when the application is running, or you can view the OpenAPI specs in the `api/` directory.

## License

MIT

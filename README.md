# FORTEPAY E-Commerce

A comprehensive e-commerce platform built with Go backend and vanilla JavaScript frontend, implementing Domain-Driven Design (DDD) principles and clean architecture.

## Table of Contents

- [Features](#features)
- [Architecture](#architecture)
  - [Backend Architecture](#backend-architecture)
  - [Frontend Architecture](#frontend-architecture)
- [Web UI](#web-ui)
  - [Web Structure](#web-structure)
  - [Web Assets](#web-assets)
  - [Web Features](#web-features)
  - [Web Authentication](#web-authentication)
  - [Web Development](#web-development)
- [Middleware](#middleware)
  - [Authentication Middleware](#authentication-middleware)
  - [Handler-Level Authentication](#handler-level-authentication)
  - [Request ID Middleware](#request-id-middleware)
  - [Logging Middleware](#logging-middleware)
  - [Middleware Factory](#middleware-factory)
  - [Tracing Middleware](#tracing-middleware)
- [Database Schema](#database-schema)
  - [Products Table](#products-table)
  - [Users Table](#users-table)
  - [Cart Items Table](#cart-items-table)
  - [Promotions Table](#promotions-table)
  - [Checkouts Table](#checkouts-table)
  - [Checkout Items Table](#checkout-items-table)
  - [Promotion Applied Table](#promotion-applied-table)
  - [Refresh Tokens Table](#refresh-tokens-table)
- [Promotion System](#promotion-system)
- [Frontend Implementation](#frontend-implementation)
- [Getting Started](#getting-started)
  - [Prerequisites](#prerequisites)
  - [Setup](#setup)
  - [Development Commands](#development-commands)
- [Running with Docker](#running-with-docker)
  - [Using Docker Compose](#method-1-using-docker-compose)
  - [Using Helper Scripts](#method-2-using-helper-scripts)
- [Environment Variables](#environment-variables)
- [License](#license)

## Features

- **Product Management**: Browse and search products
- **User Authentication**: Register, login, and JWT-based authentication
- **Shopping Cart**: Add, update, remove items
- **Promotion System**: Automatic application of various promotion types:
  - Buy one get one free (MacBook Pro comes with a free Raspberry Pi B)
  - Buy 3 pay for 2 (3 Google Home devices for the price of 2)
  - Bulk discounts (10% off when buying more than 3 Alexa Speakers)
- **Checkout Process**: Complete orders with promotions applied
- **Order Management**: Track order status

## Architecture

The project follows a clean architecture approach with Domain-Driven Design:

```
e-commerce-be/
├── api/               # OpenAPI specifications
├── cmd/               # Application entry points
├── internal/          # Internal application code
│   ├── app/           # Business domains
│   │   ├── cart/      # Cart domain
│   │   ├── checkout/  # Checkout domain
│   │   ├── product/   # Product domain
│   │   ├── promotion/ # Promotion domain
│   │   └── user/      # User domain
│   ├── infrastructure/ # Infrastructure concerns
│   └── middleware/    # HTTP middleware components
├── migrations/        # Database migrations
├── web/               # Frontend web application
│   ├── css/           # Stylesheets
│   ├── js/            # JavaScript modules
│   └── img/           # Images and icons
└── docs/              # Documentation
```

### Backend Architecture

- **Domain Layer**: Core business logic and entities
- **Repository Layer**: Data access abstraction
- **Use Case Layer**: Application-specific business rules
- **Port Layer**: Adapters for external services (HTTP, DB)

### Frontend Architecture

The frontend is built with vanilla JavaScript, structured in a modular way for maintainability:

- **common.js**: Core shared functionality (API client, auth, navigation)
- **auth.js**: Authentication functionality (login/register)
- **products.js**: Product listing and management
- **cart.js**: Shopping cart functionality
- **checkout.js**: Checkout process and order completion
- **main.js**: Application entry point and initialization

## Web UI

The web folder contains a simple client-side web UI for the FORTEPAY E-Commerce backend. It allows customers to:

- Register for a new account
- Login to their account
- Browse products
- Add items to cart
- View and manage their shopping cart

### Web Structure

- `index.html` - Landing page
- `login.html` - User login page
- `register.html` - New user registration page
- `products.html` - Product listing page
- `cart.html` - Shopping cart page
- `checkout.html` - Checkout page

### Web Assets

- `css/` - Stylesheets
- `js/` - JavaScript files
- `img/` - Images and icons

### Web Features

The UI implements the following e-commerce functionality:

1. User authentication (login/register)
2. Product browsing
3. Cart management
4. Special promotions handling:
   - Each sale of a MacBook Pro comes with a free Raspberry Pi
   - Buy 3 Google Homes for the price of 2
   - 10% discount when buying more than 3 Alexa Speakers

### Web Authentication

The UI uses JWT token authentication. After login, the token is stored in localStorage and automatically included in API requests.

### Web Development

This is a static client-side application that communicates with the backend API. To modify:

1. Edit HTML files to change the structure
2. Modify CSS in `css/style.css` to change the appearance
3. Update JavaScript files in `js/` directory to change the behavior

No build step is required as this is plain HTML/CSS/JavaScript.

## Middleware

The middleware package provides components for the e-commerce API.

### Authentication Middleware

The authentication middleware (`auth.go`) provides role-based access control:

- **Public Access**: No authentication required
- **Bearer Authentication**: JWT token validation
- **Role-based Access**: Admin and Customer role checks

```go
// For a public endpoint (no auth required)
http.Handle("/public", middlewareFactory.WrapFunc(middleware.AuthTypePublic, publicHandler))

// For an authenticated endpoint (any user)
http.Handle("/user", middlewareFactory.WrapFunc(middleware.AuthTypeBearer, userHandler))

// For an admin-only endpoint
http.Handle("/admin", middlewareFactory.WrapFunc(middleware.AuthTypeRoleAdmin, adminHandler))

// For a customer-only endpoint
http.Handle("/customer", middlewareFactory.WrapFunc(middleware.AuthTypeRoleCustomer, customerHandler))
```

### Handler-Level Authentication

The handler-level authentication (`handler.go`) provides more fine-grained control over authentication requirements at the handler level rather than the endpoint level.

#### Protected Handler

Use `ProtectedHandler` when you want to set different auth requirements for different path patterns within the same handler:

```go
// Create a protected handler with default authentication type
protected := middleware.NewProtectedHandler(middlewareFactory).
    WithDefaultAuth(middleware.AuthTypeBearer).
    // Set specific auth requirements for particular paths
    WithPathAuth("/users/public/*", middleware.AuthTypePublic).
    WithPathAuth("/users/admin/*", middleware.AuthTypeRoleAdmin)

// Wrap your handler with the authentication middleware
http.Handle("/users/", protected.Wrap(userHandler))
```

#### Method-Specific Authentication

Use `MethodAuthHandler` when different HTTP methods require different authentication levels:

```go
// Create a method-auth handler with default and method-specific authentication
methodAuth := middleware.NewMethodAuthHandler(middlewareFactory).
    WithDefaultAuth(middleware.AuthTypePublic).
    // GET is public, POST and DELETE require admin role
    WithMethodAuth(http.MethodGet, middleware.AuthTypePublic).
    WithMethodAuth(http.MethodPost, middleware.AuthTypeRoleAdmin).
    WithMethodAuth(http.MethodDelete, middleware.AuthTypeRoleAdmin).
    // Specific path+method combinations
    WithPathMethodAuth("/resources/special", http.MethodPut, middleware.AuthTypeRoleAdmin)

// Wrap your handler with the method-specific authentication middleware
http.Handle("/resources/", methodAuth.Wrap(resourceHandler))
```

### Request ID Middleware

The request ID middleware (`requestid.go`) ensures each request has a unique ID:

- Preserves existing X-Request-ID header if present
- Generates a new UUID if no request ID is provided
- Adds the request ID to the response headers
- Stores the request ID in the request context

The request ID middleware is automatically applied as part of the middleware chain.

To retrieve the request ID in your handlers:

```go
requestID := middleware.GetRequestID(r.Context())
```

### Logging Middleware

The logging middleware (`logger.go`) provides structured logging using Go's `slog` package:

- Logs detailed information about requests and responses
- Uses appropriate log levels based on response status
- Masks sensitive information in request/response bodies
- Includes request ID, method, path, headers, timing information, etc.
- Structured JSON output for easy parsing and analysis
- OpenTelemetry integration for distributed tracing

The logging middleware is automatically applied as part of the middleware chain.

Access the logger in your code:

```go
middleware.Logger.Info("Custom log message",
    slog.String("key", "value"),
    slog.Int("count", 42))
```

### Middleware Factory

The middleware factory (`middleware.go`) provides a convenient way to apply middleware chains:

```go
// Create a middleware factory
factory := middleware.NewFactory(userUseCase, config)

// Apply middleware to a handler with specific auth type
handler := factory.Apply(myHandler, middleware.AuthTypeBearer)

// Or wrap a http.HandlerFunc
handler := factory.WrapFunc(middleware.AuthTypePublic, myHandlerFunc)
```

### Tracing Middleware

The tracing middleware (`otel.go`) provides distributed tracing with OpenTelemetry:

- Automatically creates spans for each HTTP request
- Captures request details (method, path, status code)
- Integrates with the logging system to include trace context in logs
- Supports exporting traces to OpenTelemetry collectors

Initialize OpenTelemetry at application startup:

```go
shutdown, err := middleware.InitOTEL()
if err != nil {
    log.Fatalf("Failed to initialize OpenTelemetry: %v", err)
}
defer shutdown(context.Background())
```

To log with trace context:

```go
middleware.LogWithContext(ctx, slog.LevelInfo, "Message with trace context")
```

#### Middleware Configuration

Middleware configuration is loaded from environment variables:

```
# JWT Configuration
JWT_SECRET_KEY=your-secret-key-change-in-production
JWT_EXPIRATION_HOURS=24

# OpenTelemetry Configuration
OTEL_ENABLED=true
OTEL_SERVICE_NAME=e-commerce-api
OTEL_EXPORTER_OTLP_ENDPOINT=http://localhost:4317
```

## Database Schema

The application uses PostgreSQL with the following schema:

### Products Table

```sql
CREATE TABLE products (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    sku VARCHAR(50) UNIQUE NOT NULL,
    name VARCHAR(255) NOT NULL,
    price NUMERIC(10, 2) NOT NULL,
    inventory INT DEFAULT 0 NOT NULL,
    created_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMPTZ NULL
);
```

### Users Table

```sql
CREATE TABLE users (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    email VARCHAR(255) UNIQUE NOT NULL,
    password VARCHAR(255) NOT NULL,
    name VARCHAR(255) NOT NULL,
    role VARCHAR(20) DEFAULT 'customer' NOT NULL,
    created_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMPTZ NULL
);
```

### Cart Items Table

```sql
CREATE TABLE cart_items (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    product_id UUID NOT NULL REFERENCES products(id) ON DELETE RESTRICT,
    quantity INT DEFAULT 1 NOT NULL,
    created_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMPTZ NULL,
    user_id UUID REFERENCES users(id) ON DELETE CASCADE,
    CONSTRAINT cart_items_user_id_product_id_key UNIQUE (user_id, product_id)
);
```

### Promotions Table

```sql
CREATE TABLE promotions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    type VARCHAR(50) NOT NULL,
    description TEXT NOT NULL,
    rule JSONB NOT NULL,
    active BOOLEAN DEFAULT true NOT NULL,
    created_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMPTZ NULL
);
```

### Checkouts Table

```sql
CREATE TABLE checkouts (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    subtotal NUMERIC(10, 2) NOT NULL,
    total_discount NUMERIC(10, 2) NOT NULL,
    total NUMERIC(10, 2) NOT NULL,
    created_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP,
    user_id UUID REFERENCES users(id),
    payment_status VARCHAR(50) DEFAULT 'PENDING' NOT NULL,
    payment_method VARCHAR(50) NULL,
    payment_reference VARCHAR(255) NULL,
    notes TEXT NULL,
    status VARCHAR(50) DEFAULT 'CREATED' NOT NULL,
    completed_at TIMESTAMPTZ NULL
);
```

### Checkout Items Table

```sql
CREATE TABLE checkout_items (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    checkout_id UUID NOT NULL REFERENCES checkouts(id) ON DELETE CASCADE,
    product_id UUID NOT NULL REFERENCES products(id),
    product_sku VARCHAR(50) NOT NULL,
    product_name VARCHAR(255) NOT NULL,
    quantity INT NOT NULL,
    unit_price NUMERIC(10, 2) NOT NULL,
    subtotal NUMERIC(10, 2) NOT NULL,
    discount NUMERIC(10, 2) NOT NULL,
    total NUMERIC(10, 2) NOT NULL
);
```

### Promotion Applied Table

```sql
CREATE TABLE promotion_applied (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    checkout_id UUID NOT NULL REFERENCES checkouts(id) ON DELETE CASCADE,
    promotion_id UUID NOT NULL REFERENCES promotions(id),
    description TEXT NOT NULL,
    discount NUMERIC(10, 2) NOT NULL
);
```

### Refresh Tokens Table

```sql
CREATE TABLE refresh_tokens (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    token VARCHAR(255) UNIQUE NOT NULL,
    expires_at TIMESTAMPTZ NOT NULL,
    created_at TIMESTAMPTZ DEFAULT CURRENT_TIMESTAMP
);
```

## Promotion System

The application implements three types of promotions:

1. **Buy One Get One Free**: When purchasing a specific product (e.g., MacBook Pro), another product (e.g., Raspberry Pi B) is free
2. **Buy 3 Pay 2**: When purchasing three of the same product (e.g., Google Home), one is free
3. **Bulk Discount**: When purchasing more than a threshold quantity (e.g., 3 Alexa Speakers), a percentage discount is applied

Promotions are stored as JSON rules in the database and applied dynamically during checkout.

## Frontend Implementation

The web UI is structured to provide a seamless shopping experience:

- **Modular JavaScript**: Code is separated by functionality for better organization and maintainability
- **Responsive Design**: Built with Tailwind CSS for a responsive UI
- **Client-Side Routing**: Handles navigation between pages
- **Token-Based Authentication**: Uses JWT for secure authenticated requests
- **Shopping Cart Management**: Real-time updates and quantity management
- **Promotion Preview**: Shows applicable promotions in the cart

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

6. Access the application at http://localhost:8080/web/
7. Access the Swagger UI at http://localhost:8080/swagger/

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

## Running with Docker

### Method 1: Using Docker Compose

```bash
# Run with default configuration
docker-compose up -d

# Or with custom environment variables
DB_HOST=192.168.1.10 SERVER_PORT=3000 docker-compose up -d
```

The `docker-compose.yml` includes the following services:

- **app**: The main application service running the Go backend
- **postgres**: PostgreSQL database service with data persistence
- **migrations**: A service that runs database migrations automatically

Benefits of using Docker Compose:

- Complete development environment with a single command
- Automatic database setup and migrations
- Hot-reloading for the web frontend (mounted as a volume)
- Network isolation between services
- Health checks for services

### Method 2: Using Helper Scripts

```bash
# Make scripts executable
chmod +x scripts/docker-build.sh scripts/docker-run.sh

# Build image
./scripts/docker-build.sh

# Run container
./scripts/docker-run.sh
```

## Environment Variables

| Variable     | Description                          | Default              |
| ------------ | ------------------------------------ | -------------------- |
| DB_HOST      | Database host                        | host.docker.internal |
| DB_PORT      | Database port                        | 5555                 |
| DB_USER      | Database username                    | fanzru               |
| DB_PASSWORD  | Database password                    | ganteng              |
| DB_NAME      | Database name                        | ecommerce            |
| SERVER_PORT  | Server port                          | 8080                 |
| APP_ENV      | Environment (development/production) | development          |
| SWAGGER_HOST | Host for swagger URL                 | host.docker.internal |

## License

MIT

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

## Running with Docker

### Method 1: Using Docker Compose

```bash
# Run with default configuration
docker-compose up -d

# Or with custom environment variables
DB_HOST=192.168.1.10 SERVER_PORT=3000 docker-compose up -d
```

### Method 2: Using Helper Scripts

```bash
# Make scripts executable
chmod +x scripts/docker-build.sh scripts/docker-run.sh

# Build image
./scripts/docker-build.sh

# Run container
./scripts/docker-run.sh
```

### Method 3: Manual

```bash
# Build image
docker build -t e-commerce-be \
  --build-arg DB_HOST=host.docker.internal \
  --build-arg SWAGGER_HOST=host.docker.internal .

# Run container
docker run -p 8080:8080 \
  -e DB_HOST=host.docker.internal \
  -e SWAGGER_HOST=host.docker.internal \
  e-commerce-be
```

## Accessing Swagger UI

Swagger UI is available at: http://host.docker.internal:8080/swagger/

## Deployment

This project uses GitHub Actions for CI/CD:

1. **Pull Requests (CI):** Verifies code format, builds Go binary, and builds Docker image without running the database
2. **Push to Main (CD):** Automatically builds and pushes to Docker Hub

### Setup GitHub Secrets

To enable GitHub Actions, add the following secrets in your repository settings:

1. `DOCKER_HUB_USERNAME` - Docker Hub username
2. `DOCKER_HUB_ACCESS_TOKEN` - Access token for Docker Hub (not password)

### Deployment Notes

Since this application requires a database, you need to ensure a PostgreSQL database is available when running the Docker image. Configure database connection using environment variables as described in the "Environment Variables" section.

### How to Get Docker Hub Access Token

1. Login to Docker Hub
2. Click username → Account Settings → Security
3. Click "New Access Token"
4. Name the token (e.g., "GitHub Actions")
5. Select required access (minimum: "Read & Write")
6. Click "Generate" and copy the token that appears

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

# Makefile for e-commerce-be

.PHONY: run run-debug gen-http gen-swagger clean test lint deps migrate-up migrate-down migrate-create help

# Run the application with INFO log level
run:
	LOG_LEVEL=info go run cmd/core/main.go

# Run the application with DEBUG log level
run-debug:
	LOG_LEVEL=debug go run cmd/core/main.go

# Generate HTTP handlers and routes
gen-http:
	@echo "Generating HTTP handlers..."
	./scripts/generate.sh

# Generate Swagger documentation
gen-swagger:
	@echo "Generating Swagger documentation..."
	./scripts/swaggerdoc.sh
	@echo "Swagger documentation generated in docs/swagger/docs.json"

# Generate all (HTTP + Swagger)
gen-all: gen-http gen-swagger
	@echo "All code generation completed"

# Clean generated files
clean:
	@echo "Cleaning generated files..."
	rm -rf docs/swagger/*.json
	find . -name "*.gen.go" -type f -delete

# Database migrations
migrate-up:
	@echo "Running migrations up..."
	migrate -database $(DB_URL) -path migrations/postgresqldb up

migrate-down:
	@echo "Running migrations down..."
	migrate -database $(DB_URL) -path migrations/postgresqldb down

migrate-create:
	@echo "Creating new migration files..."
	@read -p "Enter migration name: " name; \
	migrate create -ext sql -dir migrations/postgresqldb -seq $$name

# Run tests
test:
	go test ./... -v

# Run linter
lint:
	golangci-lint run ./...

# Install dependencies
deps:
	go mod tidy
	go mod vendor

# Run Docker container
docker-run:
	./scripts/docker-run.sh

# Build Docker image
docker-build:
	./scripts/docker-build.sh

# Help target
help:
	@echo "Available targets:"
	@echo "  run          - Run the application with INFO level logging"
	@echo "  run-debug    - Run the application with DEBUG level logging"
	@echo "  gen-http     - Generate HTTP handlers using OpenAPI specs"
	@echo "  gen-swagger  - Generate Swagger documentation"
	@echo "  gen-all      - Generate HTTP handlers and Swagger documentation"
	@echo "  clean        - Clean generated files"
	@echo "  migrate-up   - Run database migrations forward"
	@echo "  migrate-down - Rollback database migrations"
	@echo "  migrate-create - Create a new migration file"
	@echo "  test         - Run tests"
	@echo "  lint         - Run linter"
	@echo "  deps         - Install dependencies"
	@echo "  docker-run   - Run Docker container"
	@echo "  docker-build - Build Docker image"
	@echo "  help         - Show this help message"

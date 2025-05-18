# Build stage
FROM golang:1.24-alpine AS builder

WORKDIR /app

# Install build tools and dependencies
RUN apk add --no-cache make bash git

# Copy go mod and sum files
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download

# Install oapi-codegen v2
RUN go install github.com/oapi-codegen/oapi-codegen/v2/cmd/oapi-codegen@latest

# Copy source code
COPY . .

# Make scripts executable
RUN chmod +x ./scripts/*.sh

# Generate HTTP code
RUN echo "Generating HTTP code from OpenAPI specs" && \
    make gen-http || ./scripts/generate.sh

# Build the application
RUN CGO_ENABLED=0 GOOS=linux go build -o main ./cmd/core

# Final stage
FROM gcr.io/distroless/static-debian11

WORKDIR /app

# Define build arguments with defaults
ARG DB_HOST=localhost
ARG DB_PORT=5555
ARG DB_USER=fanzru
ARG DB_PASSWORD=
ARG DB_NAME=ecommerce
ARG SERVER_PORT=8080
ARG APP_ENV=development
ARG SWAGGER_HOST=localhost

# Set environment variables
ENV DB_HOST=${DB_HOST}
ENV DB_PORT=${DB_PORT}
ENV DB_USER=${DB_USER}
ENV DB_PASSWORD=${DB_PASSWORD}
ENV DB_NAME=${DB_NAME}
ENV SERVER_PORT=${SERVER_PORT}
ENV APP_ENV=${APP_ENV}
ENV SWAGGER_HOST=${SWAGGER_HOST}

# Copy the binary from builder
COPY --from=builder /app/main .

# Copy assets and docs
COPY --from=builder /app/assets ./assets
COPY --from=builder /app/docs ./docs
COPY --from=builder /app/web ./web

# Copy environment files
COPY sample-env .env

# Expose port
EXPOSE ${SERVER_PORT}

# Run the application
CMD ["./main"]


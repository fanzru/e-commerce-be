#!/bin/bash

# Set default values
SERVER_PORT=${SERVER_PORT:-8080}
APP_ENV=${APP_ENV:-development}
DB_HOST=${DB_HOST:-host.docker.internal}
SWAGGER_HOST=${SWAGGER_HOST:-host.docker.internal}

echo "Running Docker container with following configuration:"
echo "SERVER_PORT: $SERVER_PORT"
echo "APP_ENV: $APP_ENV"
echo "DB_HOST: $DB_HOST"
echo "SWAGGER_HOST: $SWAGGER_HOST"

# Run the Docker container
docker run -p $SERVER_PORT:$SERVER_PORT \
  -e DB_HOST=$DB_HOST \
  -e DB_PORT=${DB_PORT:-5555} \
  -e DB_USER=${DB_USER:-fanzru} \
  -e DB_PASSWORD=${DB_PASSWORD:-} \
  -e DB_NAME=${DB_NAME:-ecommerce} \
  -e DB_MAX_OPEN_CONNS=${DB_MAX_OPEN_CONNS:-25} \
  -e DB_MAX_IDLE_CONNS=${DB_MAX_IDLE_CONNS:-5} \
  -e DB_CONN_MAX_LIFETIME_MINUTES=${DB_CONN_MAX_LIFETIME_MINUTES:-30} \
  -e SERVER_PORT=$SERVER_PORT \
  -e APP_ENV=$APP_ENV \
  -e LOG_LEVEL=${LOG_LEVEL:-info} \
  -e LOG_FORMAT=${LOG_FORMAT:-json} \
  -e JWT_SECRET_KEY="${JWT_SECRET_KEY:-asnfsnfasngjnahgbwub2h03hbajfbajsfb1239anf9KDNASBN*HFasndfakfnasn8na8babs1-hbxasdnas09@kdmaskdas}" \
  -e JWT_EXPIRATION_HOURS=${JWT_EXPIRATION_HOURS:-24} \
  -e OTEL_ENABLED=${OTEL_ENABLED:-false} \
  -e OTEL_SERVICE_NAME=${OTEL_SERVICE_NAME:-e-commerce-api} \
  -e OTEL_EXPORTER_OTLP_ENDPOINT=${OTEL_EXPORTER_OTLP_ENDPOINT:-http://localhost:4317} \
  -e SWAGGER_HOST=$SWAGGER_HOST \
  e-commerce-be:latest

echo "Docker container started!" 
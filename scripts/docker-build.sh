#!/bin/bash

# Set default values
DB_HOST=${DB_HOST:-host.docker.internal}
DB_PORT=${DB_PORT:-5555}
DB_USER=${DB_USER:-fanzru}
DB_PASSWORD=${DB_PASSWORD:-}
DB_NAME=${DB_NAME:-ecommerce}
SERVER_PORT=${SERVER_PORT:-8080}
APP_ENV=${APP_ENV:-development}
SWAGGER_HOST=${SWAGGER_HOST:-host.docker.internal}

echo "Building Docker image with following configuration:"
echo "DB_HOST: $DB_HOST"
echo "DB_PORT: $DB_PORT"
echo "DB_USER: $DB_USER"
echo "DB_NAME: $DB_NAME"
echo "SERVER_PORT: $SERVER_PORT"
echo "APP_ENV: $APP_ENV"
echo "SWAGGER_HOST: $SWAGGER_HOST"

# Build the Docker image with build arguments
docker build \
  --build-arg DB_HOST=$DB_HOST \
  --build-arg DB_PORT=$DB_PORT \
  --build-arg DB_USER=$DB_USER \
  --build-arg DB_PASSWORD=$DB_PASSWORD \
  --build-arg DB_NAME=$DB_NAME \
  --build-arg SERVER_PORT=$SERVER_PORT \
  --build-arg APP_ENV=$APP_ENV \
  --build-arg SWAGGER_HOST=$SWAGGER_HOST \
  -t e-commerce-be:latest .

echo "Docker image built successfully!" 
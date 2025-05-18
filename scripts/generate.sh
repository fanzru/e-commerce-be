#!/bin/bash
set -e

echo "Generating OpenAPI code..."

# Create directories if they don't exist
for file in ./api/http/*.yaml; do
    f=$(basename $file .yaml)
    mkdir -p ./internal/app/$f/port/genhttp
done

# Install oapi-codegen if not exists
if ! command -v oapi-codegen &> /dev/null; then
    echo "Installing oapi-codegen..."
    go install github.com/oapi-codegen/oapi-codegen/v2/cmd/oapi-codegen@latest
fi

# Generate code for each API
for file in ./api/http/*.yaml; do
    f=$(basename $file .yaml)
    echo "Generating code for $f..."

    # Generate types
    oapi-codegen -generate types \
        -o ./internal/app/$f/port/genhttp/types.gen.go \
        -package genhttp \
        api/http/$f.yaml

    # Generate server with net/http
    oapi-codegen -generate std-http \
        -o ./internal/app/$f/port/genhttp/server.gen.go \
        -package genhttp \
        api/http/$f.yaml
done

echo "Code generation complete!" 
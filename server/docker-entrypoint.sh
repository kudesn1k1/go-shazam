#!/bin/bash

set -e

# Load .env file if it exists (variables may also be set via docker-compose)
if [ -f .env ]; then
    echo "Loading environment variables from .env file..."
    set -a
    source .env
    set +a
fi

# Quick verification that services are ready (docker-compose healthcheck should handle this,
# but this provides additional safety, especially when running outside docker-compose)
check_service() {
    local host=$1
    local port=$2
    local service_name=$3
    local max_attempts=${4:-5}
    local attempt=0
    
    echo "Verifying ${service_name} is ready..."
    while [ $attempt -lt $max_attempts ]; do
        if timeout 1 bash -c "cat < /dev/null > /dev/tcp/${host}/${port}" 2>/dev/null; then
            echo "${service_name} is ready!"
            return 0
        fi
        attempt=$((attempt + 1))
        if [ $attempt -lt $max_attempts ]; then
            echo "${service_name} not ready yet, retrying... ($attempt/$max_attempts)"
            sleep 1
        fi
    done
    
    echo "Warning: ${service_name} may not be ready, but continuing anyway..."
    return 0
}

check_service "${DB_HOST}" "${DB_PORT}" "Database"
check_service "${REDIS_HOST}" "${REDIS_PORT}" "Redis"

echo "Migrating database..."
export GOOSE_DRIVER=postgres
export GOOSE_DBSTRING="host=${DB_HOST} port=${DB_PORT} user=${DB_USER} password=${DB_PASSWORD} dbname=${DB_DATABASE} sslmode=disable"
goose -dir ./migrations up
echo "Database migrated successfully!"

echo "Starting server..."
exec /server
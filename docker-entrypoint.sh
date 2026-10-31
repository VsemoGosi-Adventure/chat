#!/bin/sh
set -e

echo "Starting Chat Service..."

# Wait for PostgreSQL to be ready
echo "Waiting for PostgreSQL to be ready..."
max_retries=30
retry_count=0

until goose -dir /app/migrations postgres "$DATABASE_DSN" version 2>/dev/null || [ $retry_count -eq $max_retries ]; do
  retry_count=$((retry_count + 1))
  echo "PostgreSQL is unavailable (attempt $retry_count/$max_retries) - sleeping"
  sleep 2
done

if [ $retry_count -eq $max_retries ]; then
  echo "ERROR: PostgreSQL did not become ready in time"
  exit 1
fi

echo "PostgreSQL is ready!"

# Run migrations
echo "Running database migrations..."
if goose -dir /app/migrations postgres "$DATABASE_DSN" up; then
  echo "Migrations completed successfully"
else
  echo "ERROR: Migrations failed"
  exit 1
fi

# Start the application
echo "Starting application..."
exec /app/chat

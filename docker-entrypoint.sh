#!/bin/sh
# docker-entrypoint.sh

# Exit immediately if a command exits with a non-zero status.
set -e

echo "Docker Entrypoint: Starting..."

# Check if DATABASE_URL is set (required for migrations)
if [ -z "$DATABASE_URL" ]; then
  echo "[Error] DATABASE_URL environment variable is not set. Cannot run migrations."
  exit 1
fi

# Run database migrations
echo "Running database migrations from /app/migrations..."
# Ensure the URL format is compatible with migrate (e.g., postgresql://...)
# The migrate tool often expects 'postgres://' or 'postgresql://' scheme.
# We might need to transform DATABASE_URL if it uses a different scheme.
# Example transformation (adjust based on actual DATABASE_URL format):
# MIGRATION_DB_URL=$(echo "$DATABASE_URL" | sed 's/^postgres:/postgresql:/')
# For now, assume DATABASE_URL is already in the correct format.
migrate -database "$DATABASE_URL" -path /app/migrations up

echo "Database migrations completed."

# Execute the CMD passed to the entrypoint (which should be supervisord)
echo "Starting application (executing CMD: $@)..."
exec "$@"
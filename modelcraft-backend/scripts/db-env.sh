#!/usr/bin/env bash
# Common database environment loader for Taskfile db:* tasks.
# Usage: source scripts/db-env.sh <env_file>

set -e

_ENV_FILE="${1:-.env}"

echo "📄 Using ENV file: ${_ENV_FILE}"
if [ ! -f "${_ENV_FILE}" ]; then
  echo "❌ ENV file not found: ${_ENV_FILE}"
  exit 1
fi

set -a
. "${_ENV_FILE}"
set +a

export DB_USER=${DB_USER:-${DB_USERNAME:-root}}
export DB_PASSWORD=${DB_PASSWORD:-password}
export DB_HOST=${DB_HOST:-127.0.0.1}
export DB_PORT=${DB_PORT:-3306}
export DB_NAME=${DB_NAME:-${DB_DATABASE:-modelcraft}}
export DB_URL="mysql://${DB_USER}:${DB_PASSWORD}@${DB_HOST}:${DB_PORT}/${DB_NAME}"
export DB_DEV_URL="mysql://${DB_USER}:${DB_PASSWORD}@${DB_HOST}:${DB_PORT}/${DB_NAME}_dev"

echo "🔗 Database: ${DB_USER}@${DB_HOST}:${DB_PORT}/${DB_NAME}"

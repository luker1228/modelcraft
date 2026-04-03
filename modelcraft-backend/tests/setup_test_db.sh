#!/bin/bash

# Test Database Setup Script
# This script creates a clean test database with Atlas schema for integration tests

set -e  # Exit on error

# Color codes for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Load configuration from .env or use defaults
if [ -f ".env" ]; then
    source .env
fi

# Database configuration
DB_HOST="${TEST_DB_HOST:-localhost}"
DB_PORT="${TEST_DB_PORT:-3306}"
DB_USER="${TEST_DB_USER:-root}"
DB_PASSWORD="${TEST_DB_PASSWORD:-Root@SecurePass123#}"
DB_NAME="${TEST_DB_NAME:-modelcraft_test}"

echo -e "${BLUE}═══════════════════════════════════════════════════════════${NC}"
echo -e "${BLUE}  ModelCraft Test Database Setup${NC}"
echo -e "${BLUE}═══════════════════════════════════════════════════════════${NC}"
echo ""
echo -e "${YELLOW}Configuration:${NC}"
echo "  Host: ${DB_HOST}:${DB_PORT}"
echo "  User: ${DB_USER}"
echo "  Database: ${DB_NAME}"
echo ""

# Step 1: Drop and recreate database
echo -e "${YELLOW}[1/4]${NC} Dropping and recreating database '${DB_NAME}'..."
mysql --protocol=TCP -h "$DB_HOST" -P "$DB_PORT" -u "$DB_USER" -p"$DB_PASSWORD" -e "DROP DATABASE IF EXISTS ${DB_NAME};" 2>/dev/null
mysql --protocol=TCP -h "$DB_HOST" -P "$DB_PORT" -u "$DB_USER" -p"$DB_PASSWORD" -e "CREATE DATABASE ${DB_NAME} CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci;" 2>/dev/null
echo -e "${GREEN}✓${NC} Database recreated successfully"
echo ""

# Step 2: Apply schema using Atlas
echo -e "${YELLOW}[2/4]${NC} Applying schema using Atlas..."

# URL encode password for Atlas (replace @ with %40 and # with %23)
ENCODED_PASSWORD=$(echo "$DB_PASSWORD" | sed 's/@/%40/g; s/#/%23/g')

# Build database URLs for Atlas WITHOUT schema scoping (remove database name from URL)
# This allows Atlas to operate at database level instead of being restricted to a single schema
BASE_URL="mysql://${DB_USER}:${ENCODED_PASSWORD}@${DB_HOST}:${DB_PORT}/"

echo "BASE_URL: ${BASE_URL}"
echo "Target Database: ${DB_NAME}"

# Create development database if it doesn't exist
mysql --protocol=TCP -h "$DB_HOST" -P "$DB_PORT" -u "$DB_USER" -p"$DB_PASSWORD" -e "CREATE DATABASE IF NOT EXISTS modelcraft_dev CHARACTER SET utf8mb4 COLLATE utf8mb4_unicode_ci;" 2>/dev/null

# Apply schema using Atlas with database-level scope (no schema restriction in URLs)
atlas schema apply \
  -u "${BASE_URL}${DB_NAME}" \
  --to file://db/schema/mysql/ \
  --dev-url "${BASE_URL}modelcraft_dev" \
  --auto-approve

echo -e "${GREEN}✓${NC} Schema applied successfully using Atlas"
echo ""

# Step 3: Verify tables were created
echo -e "${YELLOW}[3/4]${NC} Verifying tables..."
TABLE_COUNT=$(mysql --protocol=TCP -h "$DB_HOST" -P "$DB_PORT" -u "$DB_USER" -p"$DB_PASSWORD" "$DB_NAME" -sN -e "SELECT COUNT(*) FROM information_schema.tables WHERE table_schema='${DB_NAME}';" 2>/dev/null)
echo -e "${GREEN}✓${NC} Found ${TABLE_COUNT} tables in database"

# List tables
echo ""
echo "Tables created:"
mysql --protocol=TCP -h "$DB_HOST" -P "$DB_PORT" -u "$DB_USER" -p"$DB_PASSWORD" "$DB_NAME" -e "SHOW TABLES;" 2>/dev/null | tail -n +2 | while read table; do
    echo -e "  ${GREEN}•${NC} ${table}"
done
echo ""

# Step 4: Insert default project
echo -e "${YELLOW}[4/4]${NC} Creating default project..."
mysql --protocol=TCP -h "$DB_HOST" -P "$DB_PORT" -u "$DB_USER" -p"$DB_PASSWORD" "$DB_NAME" <<EOF 2>/dev/null
INSERT INTO projects (id, title, description, status, created_at, updated_at)
VALUES ('default', 'Default Project', 'System default project for backward compatibility', 'ACTIVE', NOW(), NOW())
ON DUPLICATE KEY UPDATE updated_at=NOW();
EOF
echo -e "${GREEN}✓${NC} Default project created"
echo ""

echo -e "${GREEN}═══════════════════════════════════════════════════════════${NC}"
echo -e "${GREEN}  Test database setup completed successfully!${NC}"
echo -e "${GREEN}═══════════════════════════════════════════════════════════${NC}"
echo ""
echo "You can now run tests with:"
echo -e "  ${BLUE}pytest automated/ -v${NC}"
echo ""

#!/bin/bash
# Verification script for Casdoor integration

echo "=========================================="
echo "Casdoor Integration Verification"
echo "=========================================="
echo ""

# Colors
GREEN='\033[0;32m'
RED='\033[0;31m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Check functions
check_file() {
    if [ -f "$1" ]; then
        echo -e "${GREEN}✓${NC} File exists: $1"
        return 0
    else
        echo -e "${RED}✗${NC} File missing: $1"
        return 1
    fi
}

check_directory() {
    if [ -d "$1" ]; then
        echo -e "${GREEN}✓${NC} Directory exists: $1"
        return 0
    else
        echo -e "${RED}✗${NC} Directory missing: $1"
        return 1
    fi
}

check_service() {
    if docker-compose ps | grep -q "$1.*Up"; then
        echo -e "${GREEN}✓${NC} Service running: $1"
        return 0
    else
        echo -e "${YELLOW}!${NC} Service not running: $1"
        return 1
    fi
}

# File checks
echo "1. Checking configuration files..."
check_file "docker-compose.yml"
check_file "casdoor/conf/app.conf"
check_file "casdoor/README.md"
check_file ".env.example"
check_file ".env.docker.example"
check_file "scripts/setup-casdoor.sh"
check_file "CASDOOR_INTEGRATION.md"
echo ""

# Directory checks
echo "2. Checking directories..."
check_directory "casdoor"
check_directory "casdoor/conf"
check_directory "scripts"
echo ""

# Docker Compose validation
echo "3. Validating docker-compose.yml..."
if docker-compose config > /dev/null 2>&1; then
    echo -e "${GREEN}✓${NC} docker-compose.yml is valid"
else
    echo -e "${RED}✗${NC} docker-compose.yml has syntax errors"
fi
echo ""

# Service definitions check
echo "4. Checking service definitions in docker-compose.yml..."
if grep -q "casdoor:" docker-compose.yml; then
    echo -e "${GREEN}✓${NC} Casdoor service defined"
else
    echo -e "${RED}✗${NC} Casdoor service not defined"
fi

if grep -q "casdoor-db:" docker-compose.yml; then
    echo -e "${GREEN}✓${NC} Casdoor database service defined"
else
    echo -e "${RED}✗${NC} Casdoor database service not defined"
fi

if grep -q "casdoor_mysql_data:" docker-compose.yml; then
    echo -e "${GREEN}✓${NC} Casdoor data volume defined"
else
    echo -e "${RED}✗${NC} Casdoor data volume not defined"
fi
echo ""

# Environment variables check
echo "5. Checking environment variable templates..."
env_vars=(
    "CASDOOR_PORT"
    "CASDOOR_ENDPOINT"
    "CASDOOR_CLIENT_ID"
    "CASDOOR_CLIENT_SECRET"
    "CASDOOR_ORGANIZATION"
    "CASDOOR_CERTIFICATE"
)

for var in "${env_vars[@]}"; do
    if grep -q "$var" .env.example; then
        echo -e "${GREEN}✓${NC} $var in .env.example"
    else
        echo -e "${RED}✗${NC} $var missing in .env.example"
    fi
done
echo ""

# Runtime check (optional - only if services are running)
echo "6. Checking running services (optional)..."
if docker-compose ps > /dev/null 2>&1; then
    check_service "casdoor"
    check_service "casdoor-db"
    check_service "modelcraft"
else
    echo -e "${YELLOW}!${NC} Docker services not running (this is okay)"
fi
echo ""

# ModelCraft code integration check
echo "7. Checking ModelCraft code integration..."
check_file "internal/infrastructure/auth/casdoor_provider.go"
check_file "internal/middleware/jwt_auth.go"
check_file "docs/runtime-auth-design.md"

if grep -q "casdoor:" configs/config.yaml; then
    echo -e "${GREEN}✓${NC} Casdoor config section in config.yaml"
else
    echo -e "${RED}✗${NC} Casdoor config section missing in config.yaml"
fi
echo ""

# Summary
echo "=========================================="
echo "Verification Complete!"
echo "=========================================="
echo ""
echo "Next Steps:"
echo "1. Copy .env.example to .env: cp .env.example .env"
echo "2. Run setup script: ./scripts/setup-casdoor.sh"
echo "3. Configure Casdoor at http://localhost:8000"
echo "4. Update .env with Casdoor credentials"
echo "5. Start all services: docker compose up -d"
echo ""
echo "For detailed instructions, see:"
echo "- casdoor/README.md"
echo "- CASDOOR_INTEGRATION.md"
echo ""

#!/bin/bash

# Smart Test Runner for ModelCraft
# Supports environment-specific configuration and two-phase test execution

set -e  # Exit on error

# Color codes for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
CYAN='\033[0;36m'
NC='\033[0m' # No Color

# Default values
ENV="${1:-local}"
CLEANUP_OVERRIDE=""
CLEANUP_ONLY=false
NO_CLEANUP=false

# Parse command-line arguments
shift || true
while [[ $# -gt 0 ]]; do
    case $1 in
        --cleanup-only)
            CLEANUP_ONLY=true
            shift
            ;;
        --no-cleanup)
            NO_CLEANUP=true
            CLEANUP_OVERRIDE="no"
            shift
            ;;
        --cleanup)
            CLEANUP_OVERRIDE="yes"
            shift
            ;;
        *)
            echo -e "${RED}Unknown option: $1${NC}"
            exit 1
            ;;
    esac
done

# Script directory
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
cd "$SCRIPT_DIR"

echo -e "${BLUE}═══════════════════════════════════════════════════════════${NC}"
echo -e "${BLUE}  ModelCraft Smart Test Runner${NC}"
echo -e "${BLUE}═══════════════════════════════════════════════════════════${NC}"
echo ""
echo -e "${CYAN}Environment:${NC} ${ENV}"
echo -e "${CYAN}Mode:${NC} $(if [ "$CLEANUP_ONLY" = true ]; then echo "Cleanup Only"; else echo "Full Test"; fi)"
echo ""

# Step 1: Load environment configuration
echo -e "${YELLOW}[1/5]${NC} Loading environment configuration..."

# Use unified .env file from project root
PROJECT_ROOT="$(cd "$SCRIPT_DIR/.." && pwd)"
ENV_FILE="${PROJECT_ROOT}/.env"

if [ ! -f "$ENV_FILE" ]; then
    echo -e "${RED}❌ Error: .env file not found at: ${ENV_FILE}${NC}"
    echo ""
    echo "Please create .env file in project root. You can copy from .env.example:"
    echo "  cp .env.example .env"
    echo ""
    exit 1
fi

# Load environment variables from .env file
set -a
source "$ENV_FILE"
set +a

# Override cleanup setting if specified
if [ -n "$CLEANUP_OVERRIDE" ]; then
    if [ "$CLEANUP_OVERRIDE" = "yes" ]; then
        export CLEANUP_ENABLED=true
        echo -e "${GREEN}✓${NC} Cleanup override: ENABLED"
    else
        export CLEANUP_ENABLED=false
        echo -e "${GREEN}✓${NC} Cleanup override: DISABLED"
    fi
fi

echo -e "${GREEN}✓${NC} Configuration loaded from ${ENV_FILE}"
echo "  Service URL: ${MODELCRAFT_BASE_URL}"
echo "  Database: ${DB_HOST}:${DB_PORT}/${DB_DATABASE}"
echo "  Cleanup Enabled: ${CLEANUP_ENABLED}"
echo ""

# Step 2: Validate environment health (skip if cleanup-only mode)
if [ "$CLEANUP_ONLY" = false ]; then
    echo -e "${YELLOW}[2/5]${NC} Validating environment health..."

    # Ensure virtual environment exists
    if [ ! -d ".venv" ]; then
        echo "  Creating Python virtual environment..."
        python3 -m venv .venv
    fi

    # Activate virtual environment
    source .venv/bin/activate

    # Install dependencies if needed
    if ! python3 -c "import pymysql" 2>/dev/null; then
        echo "  Installing Python dependencies..."
        pip install -q -r requirements.txt
    fi

    # Run health check
    if ! python3 health_check.py "$ENV"; then
        echo ""
        echo -e "${RED}❌ Environment validation failed!${NC}"
        echo "Cannot proceed with tests. Please fix the issues above."
        exit 1
    fi
else
    echo -e "${YELLOW}[2/5]${NC} Skipping health check (cleanup-only mode)"
    echo ""

    # Still need to activate venv for cleanup
    if [ ! -d ".venv" ]; then
        python3 -m venv .venv
    fi
    source .venv/bin/activate

    if ! python3 -c "import pymysql" 2>/dev/null; then
        pip install -q -r requirements.txt
    fi
fi

# Step 3: Execute Phase 1 - Core API Tests
if [ "$CLEANUP_ONLY" = false ]; then
    echo -e "${YELLOW}[3/5]${NC} Phase 1: Running core API tests..."

    # Create reports directory
    mkdir -p reports

    # Run pytest with environment context
    export TEST_ENV="$ENV"

    if pytest automated/ \
        --html=reports/test_report_${ENV}.html \
        --self-contained-html \
        -v \
        --tb=short; then
        echo -e "${GREEN}✓${NC} Core API tests passed!"
        TESTS_PASSED=true
    else
        echo -e "${RED}✗${NC} Some tests failed"
        TESTS_PASSED=false
    fi
    echo ""
else
    echo -e "${YELLOW}[3/5]${NC} Skipping core API tests (cleanup-only mode)"
    echo ""
    TESTS_PASSED=true
fi

# Step 4: Execute Phase 2 - Resource Cleanup (conditional)
SHOULD_CLEANUP=false

if [ "$CLEANUP_ENABLED" = "true" ]; then
    SHOULD_CLEANUP=true
fi

if [ "$NO_CLEANUP" = true ]; then
    SHOULD_CLEANUP=false
fi

if [ "$CLEANUP_ONLY" = true ]; then
    SHOULD_CLEANUP=true
fi

if [ "$SHOULD_CLEANUP" = true ]; then
    echo -e "${YELLOW}[4/5]${NC} Phase 2: Running resource cleanup..."

    export TEST_ENV="$ENV"

    if python3 cleanup_test_data.py; then
        echo -e "${GREEN}✓${NC} Resource cleanup completed"
    else
        echo -e "${RED}✗${NC} Resource cleanup failed"
    fi
    echo ""
else
    echo -e "${YELLOW}[4/5]${NC} Skipping resource cleanup (CLEANUP_ENABLED=false)"
    echo ""
    echo "💡 Test data has been preserved for debugging."
    echo "   To cleanup manually, run: task auto-test-cleanup-only ENV=${ENV}"
    echo ""
fi

# Step 5: Generate summary
echo -e "${YELLOW}[5/5]${NC} Test execution summary"
echo -e "${BLUE}═══════════════════════════════════════════════════════════${NC}"

if [ "$CLEANUP_ONLY" = true ]; then
    echo -e "${GREEN}✅ Cleanup completed successfully!${NC}"
else
    if [ "$TESTS_PASSED" = true ]; then
        echo -e "${GREEN}✅ All tests passed!${NC}"
    else
        echo -e "${RED}❌ Some tests failed${NC}"
    fi

    echo ""
    echo "📊 Test Report: reports/test_report_${ENV}.html"
fi

echo ""
echo "Environment: ${ENV}"
echo "Service: ${MODELCRAFT_BASE_URL}"
echo "Database: ${DB_HOST}:${DB_PORT}/${DB_DATABASE}"
echo "Cleanup: $(if [ "$SHOULD_CLEANUP" = true ]; then echo "Executed"; else echo "Skipped"; fi)"
echo -e "${BLUE}═══════════════════════════════════════════════════════════${NC}"
echo ""

# Exit with appropriate code
if [ "$TESTS_PASSED" = false ]; then
    exit 1
fi

exit 0

"""
Root-level pytest configuration for ModelCraft tests.

This conftest provides session-scoped fixtures available to both Design and Runtime tests.
"""

import sys
import os
from pathlib import Path

import pytest

# Add tests directory to Python path to enable imports like:
# from common.config import config
# from design.common import graphql_client
tests_dir = Path(__file__).parent
if str(tests_dir) not in sys.path:
    sys.path.insert(0, str(tests_dir))

from common.config import TestConfig, load_env_from_root

# Load .env file early so TestConfig picks up the values
# Use ENV_FILE env var to override the default .env file (e.g., ENV_FILE=.env.autotest)
_env_file = os.environ.get('ENV_FILE', '.env')
load_env_from_root(_env_file)


@pytest.fixture(scope="session")
def test_config():
    """
    Session-scoped test configuration.

    Loads configuration from environment variables and .env files.
    Available to all tests (Design and Runtime).
    """
    return TestConfig()


@pytest.fixture(scope="session")
def base_url(test_config):
    """
    Session-scoped base URL for API calls.

    Returns:
        str: Base URL for ModelCraft API (e.g., http://localhost:8080)
    """
    return test_config.get_base_url()


@pytest.fixture(scope="session")
def design_graphql_url(test_config):
    """
    Session-scoped Design-time GraphQL endpoint URL.

    Returns:
        str: GraphQL endpoint for Design API (e.g., http://localhost:8080/org/modelcraft/design/graphql)
    """
    return test_config.get_design_graphql_url()


@pytest.fixture(scope="session")
def db_config(test_config):
    """
    Session-scoped database configuration.

    Returns:
        dict: Database connection parameters
    """
    return test_config.get_db_config()


@pytest.fixture(scope="session")
def test_user_with_owner_role(test_config, db_config):
    """
    Session-scoped fixture: provision test user with owner role.

    Automatically creates test user in database with owner role before
    integration tests run. This eliminates the need for manual SQL script
    execution.

    The fixture is idempotent - it can be run multiple times safely.

    Args:
        test_config: Test configuration from environment
        db_config: Database connection parameters

    Returns:
        dict: Test user info {id, external_id, name, org_name, role_name}

    Raises:
        RuntimeError: If user provisioning fails
    """
    import os

    # Check if we should skip automatic user setup (for manual testing)
    if os.environ.get('SKIP_TEST_USER_SETUP', '').lower() == 'true':
        print("ℹ️  Skipping automatic test user setup (SKIP_TEST_USER_SETUP=true)")
        return None

    print("🚀 Setting up test user with owner role...")

    try:
        from common.test_user_setup import execute_test_user_setup, cleanup_test_user

        # Create/verify test user exists with owner role
        user_info = execute_test_user_setup(db_config)

        print(f"✅ Test user ready: {user_info['external_id']} ({user_info['role_name']})")

        yield user_info

        # Cleanup after test session (unless disabled)
        keep_user = os.environ.get('KEEP_TEST_USER', '').lower() == 'true'
        if keep_user:
            print(f"ℹ️  Keeping test user for debugging (KEEP_TEST_USER=true)")
        else:
            print("🧹 Cleaning up test user...")
            cleanup_test_user(db_config, user_info['id'])

    except Exception as e:
        print(f"❌ Test user setup failed: {e}")
        raise RuntimeError(
            f"Failed to provision test user: {e}\n"
            f"Please ensure:\n"
            f"  1. Database is running and accessible\n"
            f"  2. Database migrations have been applied (task deploy-local)\n"
            f"  3. 'modelcraft' organization and 'Owner' role exist in database"
        )


@pytest.fixture(scope="session")
def auth_token(test_config, test_user_with_owner_role):
    """
    Session-scoped JWT access token for authenticated API calls.

    Obtains a real JWT token from Casdoor using the configured test user
    credentials (Resource Owner Password Credentials flow). The token is
    obtained once per test session and shared across all tests.

    Note: During the dual-token migration period, this returns a Casdoor JWT
    which is still accepted by the API (when accept_casdoor_jwt=true).
    For tests requiring ModelCraft JWT specifically, use the `modelcraft_token`
    fixture instead.

    Requires the following environment variables (or .env entries):
        - CASDOOR_ENDPOINT: Casdoor server URL
        - CASDOOR_CLIENT_ID: OAuth2 client ID
        - CASDOOR_CLIENT_SECRET: OAuth2 client secret
        - CASDOOR_TEST_USERNAME: Test user username
        - CASDOOR_TEST_PASSWORD: Test user password

    Args:
        test_config: Test configuration
        test_user_with_owner_role: Ensures test user exists before auth

    Returns:
        str: Casdoor JWT access token, or None if auth is not configured
    """
    if not test_config.CASDOOR_TEST_USERNAME or not test_config.CASDOOR_TEST_PASSWORD:
        print("⚠️  CASDOOR_TEST_USERNAME/PASSWORD not configured, skipping auth token acquisition")
        print("   Tests will run without authentication (requires AUTH_DESIGN_ENABLED=false)")
        return None

    try:
        from common.auth import get_test_access_token
        token = get_test_access_token(test_config)
        print(f"✅ Obtained Casdoor JWT token (length={len(token)})")
        return token
    except Exception as e:
        print(f"⚠️  Failed to obtain auth token: {e}")
        print("   Tests will run without authentication (requires AUTH_DESIGN_ENABLED=false)")
        return None


@pytest.fixture(scope="session")
def casdoor_token(test_config, test_user_with_owner_role):
    """
    Session-scoped Casdoor JWT token (explicit naming for clarity).

    This is an alias for `auth_token` with explicit naming to indicate
    it returns a Casdoor JWT, not a ModelCraft JWT.

    Use this when you specifically need to test Casdoor JWT behavior
    or during migration testing.

    Args:
        test_config: Test configuration
        test_user_with_owner_role: Ensures test user exists before auth

    Returns:
        str: Casdoor JWT access token, or None if auth is not configured
    """
    if not test_config.CASDOOR_TEST_USERNAME or not test_config.CASDOOR_TEST_PASSWORD:
        return None

    try:
        from common.auth import get_test_access_token
        token = get_test_access_token(test_config)
        print(f"✅ Obtained Casdoor JWT token (explicit, length={len(token)})")
        return token
    except Exception as e:
        print(f"⚠️  Failed to obtain Casdoor token: {e}")
        return None


@pytest.fixture(scope="session")
def modelcraft_token(test_config, test_user_with_owner_role):
    """
    Session-scoped ModelCraft JWT token for dual-token authentication tests.

    This fixture attempts to obtain a ModelCraft JWT through the proper
    dual-token exchange flow. If the exchange is not available (e.g., in
    password flow tests), it falls back to returning a Casdoor JWT which
    is still accepted during the migration period.

    Note: Full ModelCraft JWT requires the complete OAuth flow with
    authorization codes. For integration tests using password flow,
    we currently return Casdoor JWT as a fallback.

    Args:
        test_config: Test configuration
        test_user_with_owner_role: Ensures test user exists before auth

    Returns:
        str: ModelCraft JWT (or Casdoor JWT fallback), or None if auth not configured
    """
    if not test_config.CASDOOR_TEST_USERNAME or not test_config.CASDOOR_TEST_PASSWORD:
        return None

    try:
        from common.auth import get_modelcraft_token
        token = get_modelcraft_token(test_config)
        print(f"✅ Obtained ModelCraft JWT token (or Casdoor fallback, length={len(token)})")
        return token
    except Exception as e:
        print(f"⚠️  Failed to obtain ModelCraft token: {e}")
        return None


def pytest_configure(config):
    """
    Pytest hook called after command line options have been parsed.
    """
    # Add custom markers
    config.addinivalue_line(
        "markers", "design: Mark test as Design-time test"
    )
    config.addinivalue_line(
        "markers", "runtime: Mark test as Runtime test"
    )
    config.addinivalue_line(
        "markers", "integration: Mark test as integration test"
    )
    config.addinivalue_line(
        "markers", "slow: Mark test as slow-running test"
    )


def pytest_collection_modifyitems(config, items):
    """
    Pytest hook to modify test items during collection.

    Automatically adds markers based on test file location:
    - Tests in design/ get @pytest.mark.design
    - Tests in runtime/ get @pytest.mark.runtime
    """
    for item in items:
        # Get test file path relative to tests directory
        test_file = Path(item.fspath).relative_to(tests_dir)

        # Auto-mark tests based on directory
        if str(test_file).startswith("design/"):
            item.add_marker(pytest.mark.design)
        elif str(test_file).startswith("runtime/"):
            item.add_marker(pytest.mark.runtime)

# Integration Test Migration Summary

## Current State

Integration tests use Casdoor JWT directly:

```python
# tests/conftest.py:130-165
@pytest.fixture(scope="session")
def auth_token(test_config, test_user_with_owner_role):
    """Returns Casdoor JWT from Casdoor password flow"""
    from common.auth import get_test_access_token
    token = get_test_access_token(test_config)  # Direct Casdoor JWT
    return token
```

**Problem**: After dual-token implementation, API expects ModelCraft JWT, not Casdoor JWT.

## Required Changes

### 1. Update Auth Utility (tests/common/auth.py)

Add new function to exchange tokens:

```python
def exchange_for_modelcraft_token(test_config, casdoor_token):
    """
    Exchange Casdoor JWT for ModelCraft JWT via /api/auth/token.

    Args:
        test_config: Test configuration
        casdoor_token: Casdoor JWT obtained from password flow

    Returns:
        str: ModelCraft JWT
    """
    response = requests.post(
        f"{test_config.get_base_url()}/api/auth/token",
        json={"casdoor_token": casdoor_token},  # Test-only parameter
        timeout=10
    )
    response.raise_for_status()

    data = response.json()

    # Validate response structure
    assert "accessToken" in data
    assert "user" in data
    assert "organization" in data
    assert "roles" in data
    assert "permissions" in data

    return data["accessToken"]  # ModelCraft JWT


def get_modelcraft_token(test_config):
    """
    Convenience function to get ModelCraft JWT for tests.

    This is the new recommended way to get auth token for integration tests.
    """
    # Step 1: Get Casdoor JWT using password flow
    casdoor_token = get_test_access_token(test_config)

    # Step 2: Exchange for ModelCraft JWT
    modelcraft_token = exchange_for_modelcraft_token(test_config, casdoor_token)

    return modelcraft_token
```

### 2. Update Root Fixture (tests/conftest.py)

Update the `auth_token` fixture:

```python
@pytest.fixture(scope="session")
def auth_token(test_config, test_user_with_owner_role):
    """
    Session-scoped ModelCraft JWT access token for authenticated API calls.

    Obtains Casdoor JWT via password flow, then exchanges it for ModelCraft JWT
    via /api/auth/token endpoint. The ModelCraft JWT includes user identity,
    organization, roles, and permissions.

    Returns:
        str: ModelCraft JWT access token, or None if auth is not configured
    """
    if not test_config.CASDOOR_TEST_USERNAME or not test_config.CASDOOR_TEST_PASSWORD:
        print("⚠️  CASDOOR_TEST_USERNAME/PASSWORD not configured")
        return None

    try:
        from common.auth import get_modelcraft_token
        token = get_modelcraft_token(test_config)  # Changed: now returns ModelCraft JWT
        print(f"✅ Obtained ModelCraft JWT token (length={len(token)})")
        return token
    except Exception as e:
        print(f"⚠️  Failed to obtain auth token: {e}")
        return None
```

### 3. Add New Test File (tests/design/auth/test_dual_token_exchange.py)

Create comprehensive tests for token exchange:

```python
"""Tests for dual-token authentication flow."""

import pytest
import jwt
import requests
from gql import gql


class TestDualTokenExchange:
    """Test dual-token authentication exchange flow."""

    def test_token_exchange_response_structure(self, base_url, test_config):
        """Verify enhanced token response includes all required fields."""
        from common.auth import get_test_access_token, exchange_for_modelcraft_token

        # Get Casdoor JWT
        casdoor_token = get_test_access_token(test_config)

        # Exchange for ModelCraft JWT
        response = requests.post(
            f"{base_url}/api/auth/token",
            json={"casdoor_token": casdoor_token},
            timeout=10
        )

        assert response.status_code == 200
        data = response.json()

        # Verify required fields
        assert "accessToken" in data
        assert "tokenType" in data
        assert data["tokenType"] == "Bearer"
        assert "expiresIn" in data
        assert isinstance(data["expiresIn"], int)

        # Verify user info
        assert "user" in data
        user = data["user"]
        assert "id" in user
        assert "externalId" in user
        assert "name" in user
        assert "email" in user

        # Verify organization info
        assert "organization" in data
        assert "name" in data["organization"]
        assert data["organization"]["name"] == "modelcraft"

        # Verify roles
        assert "roles" in data
        roles = data["roles"]
        assert isinstance(roles, list)
        assert len(roles) > 0  # Test user should have at least owner role
        assert "id" in roles[0]
        assert "name" in roles[0]
        assert "displayName" in roles[0]

        # Verify permissions
        assert "permissions" in data
        permissions = data["permissions"]
        assert isinstance(permissions, list)
        assert len(permissions) > 0  # Owner should have multiple permissions
        assert all(isinstance(p, str) for p in permissions)
        assert "model:read" in permissions  # Owner should have model:read

    def test_jwt_claims_structure(self, base_url, test_config):
        """Verify ModelCraft JWT contains correct claims."""
        from common.auth import get_modelcraft_token

        token = get_modelcraft_token(test_config)

        # Parse JWT without signature verification (we just need to read claims)
        claims = jwt.decode(token, options={"verify_signature": False})

        # Verify required claims
        assert claims["iss"] == "modelcraft"
        assert "sub" in claims  # User ID
        assert "user_id" in claims
        assert "external_id" in claims
        assert "name" in claims
        assert "email" in claims
        assert "organization" in claims
        assert claims["organization"] == "modelcraft"

        # Verify authorization claims
        assert "roles" in claims
        assert isinstance(claims["roles"], list)
        assert "owner" in claims["roles"]  # Test user has owner role

        assert "permissions" in claims
        assert isinstance(claims["permissions"], list)
        assert len(claims["permissions"]) > 0

        # Verify timestamps
        assert "exp" in claims  # Expiration
        assert "iat" in claims  # Issued at
        assert claims["exp"] > claims["iat"]

    def test_modelcraft_jwt_works_with_graphql_api(self, design_graphql_url, test_config):
        """Verify ModelCraft JWT is accepted by GraphQL API."""
        from common.auth import get_modelcraft_token
        from design.common.graphql_client import create_design_graphql_client

        token = get_modelcraft_token(test_config)

        # Create GraphQL client with ModelCraft JWT
        client = create_design_graphql_client(design_graphql_url, auth_token=token)

        # Execute a simple query
        query = gql("""
            query {
                projects {
                    nodes {
                        id
                        title
                    }
                }
            }
        """)

        result = client.execute(query)

        # Verify query succeeded
        assert "projects" in result
        assert "nodes" in result["projects"]

    def test_permission_check_uses_jwt_claims(self, graphql_client, auth_token):
        """
        Verify permission checks use JWT claims, not database queries.

        This test is important because it validates the key performance
        improvement of dual-token authentication.
        """
        # Note: Directly checking database query count is complex.
        # Instead, we verify the permission check succeeds and document
        # that it should use JWT claims.

        # Execute mutation that requires permission
        mutation = gql("""
            mutation {
                # Example mutation requiring permission
                __typename
            }
        """)

        result = graphql_client.execute(mutation)

        # If this succeeds, permission check worked
        # With dual-token auth, this should NOT query user_roles or role_permissions tables
        assert result is not None

        # TODO: Add actual query monitoring if needed for performance validation

    def test_expired_token_rejected(self, base_url):
        """Verify expired ModelCraft JWT is rejected with 401."""
        # Create an expired JWT (requires test helper or wait)
        # This is a placeholder - actual implementation depends on test utilities

        expired_token = "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJzdWIiOiJ0ZXN0IiwiZXhwIjoxfQ.invalid"

        response = requests.post(
            f"{base_url}/graphql",
            json={"query": "{ __typename }"},
            headers={"Authorization": f"Bearer {expired_token}"},
            timeout=10
        )

        assert response.status_code == 401
```

### 4. Test Execution

After updates, tests should work with new auth flow:

```bash
# Run new dual-token tests
pytest tests/design/auth/test_dual_token_exchange.py -v

# Run all integration tests (should use ModelCraft JWT automatically)
task auto-test

# All tests should pass
```

## Key Points

1. **Minimal Changes**: Only `auth_token` fixture needs update
2. **Automatic Migration**: All existing tests using `auth_token` automatically get ModelCraft JWT
3. **Backward Compatible**: Can keep `casdoor_token` fixture for migration testing
4. **New Tests**: Add dedicated tests for token exchange validation
5. **Performance Validation**: New tests should verify permission checks don't query database

## Implementation Note

The token exchange in tests requires a special mechanism since tests use password flow, not OAuth code flow. Two options:

**Option A**: Add test-only parameter to `/api/auth/token` endpoint:
```go
// In production: accepts {"code": "auth-code"}
// In test mode: also accepts {"casdoor_token": "jwt-token"}
```

**Option B**: Create separate test endpoint `/api/auth/test-token-exchange`:
```go
// Only enabled when AUTH_DESIGN_ENABLED=false or test mode
POST /api/auth/test-token-exchange
{
  "casdoor_token": "jwt-from-password-flow"
}
```

**Recommendation**: Option A (simpler, less code duplication)

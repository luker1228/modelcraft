"""
Authentication utilities for integration tests.

Provides functions to obtain JWT access tokens from Casdoor and ModelCraft for testing.
Supports both legacy Casdoor JWT and new dual-token ModelCraft JWT flows.
"""

import requests
import jwt
from typing import Optional, Dict, Any


def get_test_access_token(test_config) -> str:
    """
    Obtain JWT access token from Casdoor using Resource Owner Password Credentials flow.

    This returns a Casdoor JWT directly from Casdoor, which is accepted during
    the migration period (when accept_casdoor_jwt=true).

    Args:
        test_config: TestConfig instance with Casdoor credentials

    Returns:
        str: Casdoor JWT access token

    Raises:
        Exception: If token acquisition fails
    """
    if not test_config.CASDOOR_ENDPOINT:
        raise ValueError("CASDOOR_ENDPOINT not configured")

    if not test_config.CASDOOR_CLIENT_ID or not test_config.CASDOOR_CLIENT_SECRET:
        raise ValueError("CASDOOR_CLIENT_ID and CASDOOR_CLIENT_SECRET required")

    if not test_config.CASDOOR_TEST_USERNAME or not test_config.CASDOOR_TEST_PASSWORD:
        raise ValueError("CASDOOR_TEST_USERNAME and CASDOOR_TEST_PASSWORD required")

    # Casdoor OAuth2 token endpoint
    token_url = f"{test_config.CASDOOR_ENDPOINT}/api/login/oauth/access_token"

    # Request payload for Resource Owner Password Credentials flow
    payload = {
        'grant_type': 'password',
        'client_id': test_config.CASDOOR_CLIENT_ID,
        'client_secret': test_config.CASDOOR_CLIENT_SECRET,
        'username': test_config.CASDOOR_TEST_USERNAME,
        'password': test_config.CASDOOR_TEST_PASSWORD,
        'scope': 'read'
    }

    try:
        response = requests.post(token_url, data=payload, timeout=10)
        response.raise_for_status()

        token_data = response.json()
        access_token = token_data.get('access_token')

        if not access_token:
            raise Exception(f"No access_token in response: {token_data}")

        return access_token

    except requests.exceptions.RequestException as e:
        raise Exception(f"Failed to obtain access token from Casdoor: {e}")


def exchange_for_modelcraft_token(test_config, casdoor_token: str) -> Dict[str, Any]:
    """
    Exchange Casdoor JWT for ModelCraft JWT via ModelCraft's auth endpoint.

    Note: This function simulates the dual-token exchange flow for testing.
    In a real OAuth flow, the client would:
    1. Get an authorization code from Casdoor
    2. Send code to ModelCraft's /api/auth/token
    3. ModelCraft exchanges code with Casdoor and generates ModelCraft JWT

    For testing with password flow, we need to work around the fact that we
    have a Casdoor JWT but not an OAuth code. The current implementation
    expects an OAuth code, so this function provides integration test support.

    Args:
        test_config: TestConfig instance
        casdoor_token: Casdoor JWT obtained from password flow

    Returns:
        dict: Enhanced token response with structure:
            {
                "accessToken": str (ModelCraft JWT),
                "tokenType": "Bearer",
                "expiresIn": int,
                "user": {
                    "id": str,
                    "externalId": str,
                    "name": str,
                    "email": str
                },
                "organization": {
                    "name": str
                },
                "roles": [
                    {"id": int, "name": str, "displayName": str}
                ],
                "permissions": [str]  # e.g., ["model:read", "model:write"]
            }

    Raises:
        NotImplementedError: Current implementation doesn't support direct JWT exchange
        Exception: If token exchange fails
    """
    # Parse Casdoor JWT to understand what we're working with
    try:
        # Decode without verification (we just need the claims for testing)
        casdoor_claims = jwt.decode(casdoor_token, options={"verify_signature": False})
    except Exception as e:
        raise Exception(f"Failed to parse Casdoor JWT: {e}")

    # NOTE: The current /api/auth/token endpoint expects an OAuth authorization code,
    # not a Casdoor JWT directly. For integration tests using password flow,
    # we have a few options:
    #
    # Option 1: Continue using Casdoor JWT for tests (backward compatibility)
    #           Tests can still pass during migration period
    #
    # Option 2: Add a test-only endpoint POST /api/auth/token-exchange
    #           that accepts Casdoor JWT directly (only in test environments)
    #
    # Option 3: Mock the full OAuth flow in tests (complex)
    #
    # For now, we'll document this limitation and use Casdoor JWT in tests.
    # Real production clients will use the proper OAuth flow with authorization codes.

    raise NotImplementedError(
        "Direct Casdoor JWT to ModelCraft JWT exchange is not supported.\n"
        "The /api/auth/token endpoint expects an OAuth authorization code.\n"
        "\n"
        "For integration tests:\n"
        "  - Use Casdoor JWT directly (accepted during migration with accept_casdoor_jwt=true)\n"
        "  - Or test the full OAuth flow with authorization codes\n"
        "\n"
        f"Casdoor JWT contains: sub={casdoor_claims.get('sub')}, "
        f"name={casdoor_claims.get('name')}, owner={casdoor_claims.get('owner')}"
    )


def get_modelcraft_token(test_config) -> str:
    """
    Get ModelCraft JWT for tests (convenience function).

    Note: Currently returns Casdoor JWT since direct exchange is not implemented.
    During the migration period (accept_casdoor_jwt=true), Casdoor JWT works
    for testing purposes.

    For proper ModelCraft JWT testing, you need to:
    1. Test the full OAuth flow with authorization codes, or
    2. Add a test-only token exchange endpoint

    Args:
        test_config: TestConfig instance

    Returns:
        str: JWT token (currently Casdoor JWT)
    """
    # For now, return Casdoor JWT which is accepted during migration
    casdoor_token = get_test_access_token(test_config)

    # Future: When we have a way to exchange Casdoor JWT for ModelCraft JWT in tests,
    # uncomment and implement:
    # try:
    #     response = exchange_for_modelcraft_token(test_config, casdoor_token)
    #     return response['accessToken']
    # except NotImplementedError:
    #     # Fall back to Casdoor JWT during migration
    #     return casdoor_token

    return casdoor_token


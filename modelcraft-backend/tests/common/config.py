"""
Test configuration module.

Loads database and API configuration from environment variables.
"""

import os
from pathlib import Path
from dotenv import load_dotenv


def load_env_from_root(env_file='.env'):
    """
    Load environment variables from .env file at project root.

    Args:
        env_file (str): Name of .env file to load (default: '.env')
    """
    # Navigate from tests/common/config.py to project root
    project_root = Path(__file__).parent.parent.parent
    env_path = project_root / env_file

    if env_path.exists():
        load_dotenv(env_path, override=True)
        print(f"✅ Loaded environment from {env_path}")
    else:
        print(f"⚠️  Warning: .env file not found at {env_path}")


class TestConfig:
    """Test configuration from environment variables."""

    def __init__(self):
        """
        Initialize test configuration.

        Environment variables should be loaded before instantiation
        using load_env_from_root().
        """
        # Load Casdoor configuration
        self.CASDOOR_ENDPOINT = os.getenv('CASDOOR_ENDPOINT', '')
        self.CASDOOR_CLIENT_ID = os.getenv('CASDOOR_CLIENT_ID', '')
        self.CASDOOR_CLIENT_SECRET = os.getenv('CASDOOR_CLIENT_SECRET', '')
        self.CASDOOR_TEST_USERNAME = os.getenv('CASDOOR_TEST_USERNAME', '')
        self.CASDOOR_TEST_PASSWORD = os.getenv('CASDOOR_TEST_PASSWORD', '')

    def get_db_config(self):
        """
        Get database configuration.

        Returns:
            dict: Database configuration with keys:
                - host: Database host
                - port: Database port
                - user: Database user
                - password: Database password
                - database: Database name
        """
        return {
            'host': os.getenv('DB_HOST', '127.0.0.1'),
            'port': int(os.getenv('DB_PORT', 3307)),
            'user': os.getenv('DB_USER', 'root'),
            'password': os.getenv('DB_PASSWORD', 'modelcraft123'),
            'database': os.getenv('DB_NAME', 'modelcraft')
        }

    def get_base_url(self):
        """
        Get API base URL.

        Returns:
            str: API base URL (e.g., http://localhost:8080)
        """
        host = os.getenv('API_HOST', 'localhost')
        port = os.getenv('API_PORT', '8080')
        return f"http://{host}:{port}"

    def get_api_base_url(self):
        """
        Get API base URL (legacy alias for get_base_url).

        Returns:
            str: API base URL (e.g., http://localhost:8080)
        """
        return self.get_base_url()

    def get_design_graphql_url(self, org_name=None):
        """
        Get Design-time GraphQL endpoint URL.

        Args:
            org_name (str, optional): Organization name. Defaults to 'modelcraft' if not provided.

        Returns:
            str: Design GraphQL URL (e.g., http://localhost:8080/org/modelcraft/design/graphql)
        """
        if org_name is None:
            org_name = os.getenv('DEFAULT_ORG_NAME', 'modelcraft')
        return f"{self.get_base_url()}/org/{org_name}/design/graphql"

    def get_graphql_url(self):
        """
        Get GraphQL endpoint URL (legacy method).

        Returns:
            str: GraphQL URL (e.g., http://localhost:8080/graphql)
        """
        return f"{self.get_base_url()}/graphql"

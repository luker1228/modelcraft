"""
Database connection configuration for tests.

This module provides a centralized configuration for database connections
used in integration tests. Configuration can be overridden via environment variables.

Environment Variables:
    TEST_DB_HOST: Database host (default: localhost)
    TEST_DB_PORT: Database port (default: 3307)
    TEST_DB_USER: Database username (default: root)
    TEST_DB_PASSWORD: Database password (default: modelcraft123)
"""

import os
from dataclasses import dataclass


@dataclass
class DatabaseConfig:
    """Database connection configuration."""
    
    host: str
    port: int
    username: str
    password: str
    
    @classmethod
    def from_env(cls) -> "DatabaseConfig":
        """
        Create configuration from environment variables with fallback defaults.
        
        Returns:
            DatabaseConfig: Configuration instance
        """
        return cls(
            host=os.getenv("TEST_DB_HOST", "localhost"),
            port=int(os.getenv("TEST_DB_PORT", "3307")),
            username=os.getenv("TEST_DB_USER", "root"),
            password=os.getenv("TEST_DB_PASSWORD", "modelcraft123"),
        )
    
    def to_dict(self) -> dict:
        """
        Convert configuration to dictionary format.
        
        Returns:
            dict: Configuration as dictionary with keys: host, port, username, password
        """
        return {
            "host": self.host,
            "port": self.port,
            "username": self.username,
            "password": self.password,
        }


# Singleton instance for default test database configuration
_default_config: DatabaseConfig = None


def get_test_db_config() -> DatabaseConfig:
    """
    Get the default test database configuration.
    
    This function returns a singleton instance, created once from environment
    variables. Use this for most test scenarios.
    
    Returns:
        DatabaseConfig: Default test database configuration
    
    Example:
        >>> config = get_test_db_config()
        >>> print(config.host, config.port)
        localhost 3307
    """
    global _default_config
    
    if _default_config is None:
        _default_config = DatabaseConfig.from_env()
    
    return _default_config


def reset_test_db_config() -> None:
    """
    Reset the default test database configuration.
    
    This is useful for testing or when you need to reload configuration
    from environment variables.
    """
    global _default_config
    _default_config = None

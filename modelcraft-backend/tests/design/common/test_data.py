"""
Test data builders for Design-time tests.

Provides builder functions to generate test data with sensible defaults.
Reduces duplication and makes tests more maintainable.
"""

import uuid
from typing import Optional


def generate_test_id(prefix: str = "test") -> str:
    """
    Generate a unique test ID.

    Args:
        prefix: ID prefix (default: "test")

    Returns:
        str: Unique ID like "test_a1b2c3d4"
    """
    return f"{prefix}_{uuid.uuid4().hex[:8]}"


def build_project_input(
    name: Optional[str] = None,
    title: Optional[str] = None,
    description: str = "",
    project_id: Optional[str] = None,  # Deprecated: use name instead
    cluster_name: Optional[str] = None,
    cluster_title: Optional[str] = None,
    cluster_description: str = "",
    host: Optional[str] = None,
    port: Optional[int] = None,
    username: Optional[str] = None,
    password: Optional[str] = None,
    skip_connection_test: bool = True,
    **kwargs,
) -> dict:
    """
    Build a project creation input with sensible defaults.

    Includes clusterInput since CreateProjectInput now requires a cluster.

    Args:
        name: Project name (auto-generated if not provided)
        title: Project title (default: "Test Project {name}")
        description: Project description
        project_id: (Deprecated) Alias for name parameter, maintained for backward compatibility
        cluster_name: Cluster name (defaults to "{name}-cluster")
        cluster_title: Cluster title (defaults to "Cluster for {name}")
        cluster_description: Cluster description
        host: Database host (default: from test config)
        port: Database port (default: from test config)
        username: Database username (default: from test config)
        password: Database password (default: from test config)
        skip_connection_test: Skip connection test (default: True for tests)

    Returns:
        dict: Project input data matching CreateProjectInput schema

    Example:
        >>> input_data = build_project_input(name="my_project", title="My Project")
        >>> # Returns: {"name": "my_project", "title": "My Project", "clusterInput": {...}, ...}
    """
    from design.common.db_config import get_test_db_config

    # Support backward compatibility: project_id as alias for name
    if project_id is not None and name is None:
        name = project_id

    if name is None:
        name = generate_test_id("test_project")

    if title is None:
        title = f"Test Project {name[-8:]}"

    if cluster_name is None:
        cluster_name = f"{name}_cluster"

    if cluster_title is None:
        cluster_title = f"Cluster for {name}"

    db_config = get_test_db_config()
    final_host = host if host is not None else db_config.host
    final_port = port if port is not None else db_config.port
    final_username = username if username is not None else db_config.username
    final_password = password if password is not None else db_config.password

    return {
        "name": name,
        "title": title,
        "description": description,
        "clusterInput": {
            "name": cluster_name,
            "title": cluster_title,
            "description": cluster_description,
            "connectionInfo": {
                "host": final_host,
                "port": final_port,
                "username": final_username,
                "password": final_password,
            },
        },
        "skipConnectionTest": skip_connection_test,
    }


def build_cluster_input(
    project_name: str = "default",
    name: Optional[str] = None,
    title: Optional[str] = None,
    description: str = "",
    host: Optional[str] = None,
    port: Optional[int] = None,
    username: Optional[str] = None,
    password: Optional[str] = None,
    project_id: Optional[str] = None,  # Deprecated: use project_name instead
) -> dict:
    """
    Build a database cluster input with sensible defaults.

    Uses test database configuration from get_test_db_config() when connection
    parameters are not explicitly provided.

    Args:
        project_name: Project name (default: "default")
        name: Cluster name (auto-generated if not provided)
        title: Cluster title (defaults to name if not provided)
        description: Cluster description
        host: Database host (default: from test config)
        port: Database port (default: from test config)
        username: Database username (default: from test config)
        password: Database password (default: from test config)
        project_id: (Deprecated) Alias for project_name, maintained for backward compatibility

    Returns:
        dict: Cluster input data matching CreateDatabaseClusterInput schema

    Example:
        >>> input_data = build_cluster_input(project_name="my_project", name="test_cluster")
        >>> # Uses default test database configuration
        >>> input_data = build_cluster_input(project_name="my_project", host="custom.host", port=3306)
        >>> # Override specific connection parameters
    """
    # Import here to avoid circular dependency
    from design.common.db_config import get_test_db_config
    
    # Support backward compatibility: project_id as alias for project_name
    if project_id is not None:
        project_name = project_id

    if name is None:
        name = generate_test_id("test-cluster")

    if title is None:
        title = f"Test Cluster {name}"

    # Get default test database configuration
    db_config = get_test_db_config()
    
    # Use provided values or fall back to test config defaults
    final_host = host if host is not None else db_config.host
    final_port = port if port is not None else db_config.port
    final_username = username if username is not None else db_config.username
    final_password = password if password is not None else db_config.password

    return {
        "projectSlug": project_name,
        "name": name,
        "title": title,
        "description": description,
        "connectionInfo": {
            "host": final_host,
            "port": final_port,
            "username": final_username,
            "password": final_password,
        }
    }


def build_model_input(
    project_name: str = "default",
    name: Optional[str] = None,
    title: Optional[str] = None,
    description: str = "",
    cluster_name: Optional[str] = None,
    database_name: str = "test_db",
    project_id: Optional[str] = None,  # Deprecated: use project_name instead
    **kwargs,
) -> dict:
    """
    Build a model creation input with sensible defaults.

    Args:
        project_name: Project name (default: "default")
        name: Model name (auto-generated if not provided)
        title: Model title (defaults to "Test Model {name}")
        description: Model description
        cluster_name: Database cluster name
        database_name: Database name (default: "test_db")
        project_id: (Deprecated) Alias for project_name, maintained for backward compatibility

    Returns:
        dict: Model input data matching CreateModelInput schema

    Example:
        >>> input_data = build_model_input(project_name="my_project", cluster_name="main", name="User")
    """
    # Support backward compatibility: project_id as alias for project_name
    if project_id is not None:
        project_name = project_id

    if name is None:
        name = f"TestModel{generate_test_id()[-8:].upper()}"

    if title is None:
        title = f"Test Model {name}"

    return {
        "projectSlug": project_name,
        "name": name,
        "title": title,
        "description": description,
        "clusterName": cluster_name or "test_cluster",
        "databaseName": database_name,
    }


def build_field_input(
    name: Optional[str] = None,
    field_type: str = "string",
    required: bool = False,
    unique: bool = False,
    default_value: Optional[str] = None,
) -> dict:
    """
    Build a field input with sensible defaults.

    Args:
        name: Field name (auto-generated if not provided)
        field_type: Field type (string, int, float, boolean, datetime, etc.)
        required: Whether field is required
        unique: Whether field is unique
        default_value: Default value

    Returns:
        dict: Field input data

    Example:
        >>> field = build_field_input(name="email", field_type="string", unique=True)
    """
    if name is None:
        name = f"test_field_{uuid.uuid4().hex[:6]}"

    result = {
        "name": name,
        "type": field_type,
        "required": required,
        "unique": unique,
    }

    if default_value is not None:
        result["defaultValue"] = default_value

    return result


def build_enum_option(code: str, label: str, order: int, description: str = "") -> dict:
    """
    Build an enum option input.

    Args:
        code: Option code (unique identifier)
        label: Display label
        order: Sort order
        description: Optional description

    Returns:
        dict: Enum option input data

    Example:
        >>> option = build_enum_option("active", "Active", 1, "Active status")
    """
    return {
        "code": code,
        "label": label,
        "order": order,
        "description": description,
    }


def build_enum_input(
    project_name: str = "default",
    name: Optional[str] = None,
    display_name: Optional[str] = None,
    description: str = "",
    options: Optional[list] = None,
    is_multi_select: bool = False,  # Deprecated: CreateEnumInput no longer accepts this field
    project_id: Optional[str] = None,  # Deprecated: use project_name instead
) -> dict:
    """
    Build an enum creation input with sensible defaults.

    Args:
        project_name: Project name (default: "default")
        name: Enum name (auto-generated if not provided)
        display_name: Enum display name (defaults to name if not provided)
        description: Enum description
        options: List of enum options (default status options if not provided)
        is_multi_select: (Deprecated) Kept for backward compatibility; ignored
        project_id: (Deprecated) Alias for project_name, maintained for backward compatibility

    Returns:
        dict: Enum input data matching CreateEnumInput schema

    Example:
        >>> enum_input = build_enum_input(project_name="my_project", name="Status", options=[...])
    """
    # Support backward compatibility: project_id as alias for project_name
    if project_id is not None:
        project_name = project_id

    if name is None:
        name = f"TestEnum{generate_test_id()[-8:].upper()}"

    if display_name is None:
        display_name = f"Test Enum {name}"

    if options is None:
        options = [
            build_enum_option("active", "Active", 1, "Active status"),
            build_enum_option("inactive", "Inactive", 2, "Inactive status"),
            build_enum_option("pending", "Pending", 3, "Pending status"),
        ]

    return {
        "projectSlug": project_name,
        "name": name,
        "displayName": display_name,
        "description": description,
        "options": options,
    }

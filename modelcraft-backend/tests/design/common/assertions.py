"""
Custom assertions for Design-time tests.

Provides reusable assertion helpers for common validation patterns.
"""

from typing import Any, Optional, List


def assert_graphql_success(result: dict, expected_field: Optional[str] = None):
    """
    Assert that a GraphQL query succeeded without errors.

    Args:
        result: GraphQL query result
        expected_field: Optional expected field in result

    Raises:
        AssertionError: If query has errors or expected field is missing

    Example:
        >>> assert_graphql_success(result, "createProject")
    """
    assert result is not None, "GraphQL result is None"
    assert "errors" not in result or not result["errors"], \
        f"GraphQL query returned errors: {result.get('errors')}"

    if expected_field:
        assert expected_field in result, \
            f"Expected field '{expected_field}' not found in result: {result.keys()}"


def assert_project_fields(project: dict, expected: dict):
    """
    Assert that a project object has expected fields and values.

    Args:
        project: Project object from API
        expected: Expected field values

    Example:
        >>> assert_project_fields(project, {"id": "test-123", "title": "Test"})
    """
    assert project is not None, "Project object is None"

    for key, value in expected.items():
        assert key in project, f"Field '{key}' not found in project"
        assert project[key] == value, \
            f"Field '{key}': expected {value}, got {project[key]}"


def assert_cluster_fields(cluster: dict, expected: dict):
    """
    Assert that a cluster object has expected fields and values.

    Args:
        cluster: Cluster object from API
        expected: Expected field values

    Example:
        >>> assert_cluster_fields(cluster, {"name": "test-cluster", "host": "localhost"})
    """
    assert cluster is not None, "Cluster object is None"

    for key, value in expected.items():
        assert key in cluster, f"Field '{key}' not found in cluster"
        if value is not None:  # Skip None checks (optional fields)
            assert cluster[key] == value, \
                f"Field '{key}': expected {value}, got {cluster[key]}"


def assert_model_fields(model: dict, expected: dict):
    """
    Assert that a model object has expected fields and values.

    Args:
        model: Model object from API
        expected: Expected field values

    Example:
        >>> assert_model_fields(model, {"name": "User", "tableName": "users"})
    """
    assert model is not None, "Model object is None"

    for key, value in expected.items():
        assert key in model, f"Field '{key}' not found in model"
        assert model[key] == value, \
            f"Field '{key}': expected {value}, got {model[key]}"


def assert_contains_fields(obj: dict, fields: List[str]):
    """
    Assert that an object contains all specified fields.

    Args:
        obj: Object to check
        fields: List of required field names

    Example:
        >>> assert_contains_fields(project, ["name", "title", "status"])
    """
    assert obj is not None, "Object is None"

    for field in fields:
        assert field in obj, f"Required field '{field}' not found in object"


def assert_not_empty(value: Any, message: str = "Value is empty"):
    """
    Assert that a value is not empty (not None, not empty string, not empty list).

    Args:
        value: Value to check
        message: Custom error message

    Example:
        >>> assert_not_empty(project_name, "Project name is empty")
    """
    assert value is not None, f"{message} (value is None)"
    assert value != "", f"{message} (value is empty string)"
    assert value != [], f"{message} (value is empty list)"
    assert value != {}, f"{message} (value is empty dict)"


def assert_graphql_error(result: dict, mutation_field: str, expected_error_type: str):
    """
    Assert that a GraphQL mutation returned a specific error type.

    Args:
        result: GraphQL mutation result
        mutation_field: The mutation field name (e.g., "createProject")
        expected_error_type: Expected error type name (e.g., "ProjectAlreadyExists")

    Raises:
        AssertionError: If expected error is not found

    Example:
        >>> assert_graphql_error(result, "createProject", "ProjectAlreadyExists")
    """
    assert result is not None, "GraphQL result is None"
    assert mutation_field in result, f"Mutation field '{mutation_field}' not found in result"
    
    mutation_result = result[mutation_field]
    assert "errors" in mutation_result, f"No errors field found in {mutation_field}"
    
    errors = mutation_result["errors"]
    assert errors, f"Errors list is empty in {mutation_field}"
    assert len(errors) > 0, f"Expected at least one error in {mutation_field}"
    
    # Check if any error has the expected type
    error_types = [error.get("__typename") for error in errors]
    assert expected_error_type in error_types, \
        f"Expected error type '{expected_error_type}' not found. Got: {error_types}"


def assert_query_returns_none(result: dict, query_field: str):
    """
    Assert that a GraphQL query returns None (e.g., when resource not found).

    Args:
        result: GraphQL query result
        query_field: The query field name (e.g., "project")

    Example:
        >>> assert_query_returns_none(result, "project")
    """
    assert result is not None, "GraphQL result is None"
    assert "errors" not in result or not result["errors"], \
        f"GraphQL query returned errors: {result.get('errors')}"
    assert query_field in result, f"Query field '{query_field}' not found in result"
    assert result[query_field] is None, \
        f"Expected {query_field} to be None, but got: {result[query_field]}"

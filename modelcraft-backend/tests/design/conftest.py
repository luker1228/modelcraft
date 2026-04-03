"""
Design-time test fixtures and configuration.

This conftest provides fixtures specific to Design-time API testing.
"""

import pytest
from gql import Client

from design.common.graphql_client import create_design_graphql_client
from design.common.test_data import generate_test_id


@pytest.fixture(scope="session")
def auth_token(test_config):
    """
    Session-scoped JWT access token for Design tests.

    Design tests require authentication. This fixture overrides the root
    auth_token fixture to enforce mandatory credentials.

    Args:
        test_config: Test configuration from root conftest

    Returns:
        str: JWT access token

    Raises:
        RuntimeError: If CASDOOR_TEST_USERNAME or CASDOOR_TEST_PASSWORD not configured
    """
    if not test_config.CASDOOR_TEST_USERNAME or not test_config.CASDOOR_TEST_PASSWORD:
        raise RuntimeError(
            "Authentication required: CASDOOR_TEST_USERNAME and CASDOOR_TEST_PASSWORD "
            "must be configured in .env file"
        )

    from common.auth import get_test_access_token
    token = get_test_access_token(test_config)
    print(f"✅ Obtained test access token for Design tests (length={len(token)})")
    return token


@pytest.fixture(scope="module")
def graphql_client(design_graphql_url, auth_token):
    """
    Module-scoped authenticated GraphQL client for Design-time API.

    Args:
        design_graphql_url: GraphQL endpoint URL from root conftest
        auth_token: JWT access token from root conftest (may be None)

    Returns:
        Client: Configured GQL client instance with authentication
    """
    return create_design_graphql_client(design_graphql_url, token=auth_token)


@pytest.fixture
def created_projects(graphql_client):
    """
    Fixture to track and cleanup created projects.

    Projects are stored as dictionaries with 'name' and optionally 'orgName'.

    Usage:
        def test_something(graphql_client, created_projects):
            # Create project
            result = graphql_client.execute(CREATE_PROJECT, ...)
            project = result["createProject"]["project"]

            # Store project info
            created_projects.append({
                "name": project["name"],
                "orgName": project["orgName"]  # optional
            })

            # Test cleanup happens automatically after test
    """
    projects = []
    yield projects

    # Cleanup: Delete all created projects after test
    if projects:
        from gql import gql
        DELETE_PROJECT = gql("""
            mutation DeleteProject($name: String!) {
                deleteProject(name: $name) {
                    success
                }
            }
        """)

        for project_info in projects:
            project_name = None
            try:
                # Handle both dict and legacy formats (tuple/string)
                if isinstance(project_info, dict):
                    project_name = project_info["name"]
                elif isinstance(project_info, tuple):
                    project_name = project_info[1]  # Legacy: (orgName, projectName)
                else:
                    project_name = project_info  # Legacy: projectName string

                result = graphql_client.execute(DELETE_PROJECT, variable_values={
                    "name": project_name
                })

                # Verify deletion success
                if result and "deleteProject" in result:
                    payload = result["deleteProject"]
                    if payload.get("success"):
                        print(f"✓ Cleaned up project: {project_name}")
                    else:
                        print(f"✗ Failed to cleanup project {project_name}: success=false in response")
                        print(f"  Response: {payload}")
                else:
                    print(f"✗ Failed to cleanup project {project_name}: unexpected response format")
                    print(f"  Response: {result}")

            except Exception as e:
                project_name_str = project_name if project_name else (
                    project_info.get("name") if isinstance(project_info, dict) else str(project_info)
                )
                print(f"✗ Failed to cleanup project {project_name_str}: {type(e).__name__}: {e}")


@pytest.fixture
def created_clusters(graphql_client):
    """
    Legacy fixture for backward compatibility.

    Clusters are now sub-resources of projects and are deleted automatically
    when the project is deleted. This fixture is kept for test code that
    still references it, but performs no cleanup action.

    Prefer using created_projects for cleanup instead.
    """
    clusters = []
    yield clusters
    # No cleanup needed: clusters are cascade-deleted with their project.


@pytest.fixture
def created_models(graphql_client):
    """
    Fixture to track and cleanup created models.

    Models are stored as tuples of (projectSlug, id).

    Usage:
        def test_something(graphql_client, created_models):
            # Create model
            result = graphql_client.execute(CREATE_MODEL, ...)
            model = result["createModel"]["model"]
            created_models.append((model["projectSlug"], model["id"]))
    """
    models = []
    yield models

    # Cleanup: Delete all created models after test
    if models:
        from gql import gql
        DELETE_MODEL = gql("""
            mutation DeleteModel($projectSlug: String!, $id: ID!) {
                deleteModel(projectSlug: $projectSlug, id: $id) {
                    success
                }
            }
        """)

        for project_slug, model_id in models:
            try:
                graphql_client.execute(DELETE_MODEL, variable_values={
                    "projectSlug": project_slug,
                    "id": model_id
                })
                print(f"✓ Cleaned up model: {project_slug}/{model_id}")
            except Exception as e:
                print(f"⚠ Failed to cleanup model {project_slug}/{model_id}: {e}")


@pytest.fixture
def created_enums(graphql_client):
    """
    Fixture to track and cleanup created enums.

    Enums are stored as tuples of (projectSlug, enumName).

    Usage:
        def test_something(graphql_client, created_enums):
            # Create enum
            result = graphql_client.execute(CREATE_ENUM, ...)
            enum = result["createEnum"]["enum"]
            created_enums.append((enum["projectSlug"], enum["name"]))
    """
    enums = []
    yield enums

    # Cleanup: Delete all created enums after test
    if enums:
        from gql import gql
        DELETE_ENUM = gql("""
            mutation DeleteEnum($projectSlug: String!, $name: String!) {
                deleteEnum(projectSlug: $projectSlug, name: $name) {
                    success
                }
            }
        """)

        for project_slug, enum_name in enums:
            try:
                graphql_client.execute(DELETE_ENUM, variable_values={
                    "projectSlug": project_slug,
                    "name": enum_name
                })
                print(f"✓ Cleaned up enum: {project_slug}/{enum_name}")
            except Exception as e:
                print(f"⚠ Failed to cleanup enum {project_slug}/{enum_name}: {e}")


@pytest.fixture
def default_project(graphql_client, created_projects):
    """
    Fixture that ensures the default project exists.

    Returns:
        dict: Default project data with name

    Usage:
        def test_something(graphql_client, default_project):
            # Use default project
            assert default_project["name"] == "default"
    """
    from gql import gql

    GET_PROJECT = gql("""
        query GetProject($name: String!) {
            project(name: $name) {
                project {
                    name
                    title
                    status
                }
                error {
                    __typename
                }
            }
        }
    """)

    # Check if default project exists
    result = graphql_client.execute(GET_PROJECT, variable_values={
        "name": "default"
    })

    project_payload = result.get("project", {})
    project = project_payload.get("project")

    if project:
        return project

    # Default project doesn't exist, create it
    CREATE_PROJECT = gql("""
        mutation CreateProject($input: CreateProjectInput!) {
            createProject(input: $input) {
                project {
                    name
                    title
                    status
                }
            }
        }
    """)

    from design.common.test_data import build_project_input
    input_data = build_project_input(
        name="default",
        title="Default Project",
        description="Default project for testing"
    )

    create_result = graphql_client.execute(CREATE_PROJECT, variable_values={"input": input_data})
    project = create_result["createProject"]["project"]

    # Track for cleanup (though default project usually persists)
    created_projects.append(project["name"])

    return project
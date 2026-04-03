"""
Runtime test configuration and fixtures.

This conftest provides fixtures specific to Runtime tests.
Runtime tests can also import and use Design utilities from tests.design.common.
"""

import pytest

# Runtime tests can import Design utilities
from design.common.graphql_client import create_design_graphql_client


@pytest.fixture(scope="module")
def runtime_graphql_client(design_graphql_url):
    """
    Module-scoped GraphQL client for Runtime API.

    Note: Currently uses the same endpoint as Design API.
    Update this if Runtime has a separate endpoint.

    Example:
        >>> def test_query_data(runtime_graphql_client):
        ...     result = runtime_graphql_client.execute(query)
    """
    client = create_design_graphql_client(design_graphql_url)
    print(f"\n🔗 Runtime GraphQL endpoint: {design_graphql_url}")
    return client


@pytest.fixture(scope="module")
def sample_model_setup(graphql_client):
    """
    Fixture to set up a sample model for Runtime tests.

    Runtime tests often need models created in the Design phase.
    This fixture demonstrates how to use Design utilities to set up test data.

    Returns:
        dict: Model information (id, name, etc.)

    Example:
        >>> def test_query_model_data(runtime_graphql_client, sample_model_setup):
        ...     model_id = sample_model_setup["id"]
        ...     # Query runtime data for this model
    """
    from design.common.test_data import build_project_input, build_model_input
    from gql import gql

    # Create a project with embedded cluster for the test model
    CREATE_PROJECT = gql("""
        mutation CreateProject($input: CreateProjectInput!) {
            createProject(input: $input) {
                project {
                    name
                    cluster {
                        name
                    }
                }
                error {
                    __typename
                }
            }
        }
    """)

    project_name = "runtime_test_project"
    project_input = build_project_input(
        name=project_name,
        title="Runtime Test Project",
        cluster_name="runtime-test-cluster",
        cluster_title="Runtime Test Cluster",
        skip_connection_test=True,
    )
    project_result = graphql_client.execute(CREATE_PROJECT, variable_values={"input": project_input})
    payload = project_result.get("createProject", {})
    if payload.get("error") and payload["error"].get("__typename") != "ProjectAlreadyExists":
        raise Exception(f"Failed to create runtime test project: {payload['error']}")

    cluster_name = "runtime-test-cluster"

    # Create a test model
    CREATE_MODEL = gql("""
        mutation CreateModel($input: CreateModelInput!) {
            createModel(input: $input) {
                model {
                    id
                    name
                    tableName
                }
                error {
                    __typename
                }
            }
        }
    """)

    model_input = build_model_input(
        project_name=project_name,
        cluster_name=cluster_name,
        name="RuntimeTestModel",
        database_name="test_db",
    )
    model_result = graphql_client.execute(CREATE_MODEL, variable_values={"input": model_input})
    model = model_result["createModel"]["model"]

    yield model

    # Cleanup is optional here - Design tests should handle cleanup
    # Or implement cleanup if needed
    print(f"\n🧹 Runtime test model {model['id']} used for testing")

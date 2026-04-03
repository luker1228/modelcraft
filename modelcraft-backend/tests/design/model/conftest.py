"""
Shared fixtures for model tests
"""
import pytest
from gql import gql

from design.common.test_data import generate_test_id, build_project_input
from design.common.db_config import get_test_db_config


# Constants
DEFAULT_PROJECT_NAME = "default"


CREATE_PROJECT = gql("""
    mutation CreateProject($input: CreateProjectInput!) {
        createProject(input: $input) {
            project {
                name
                cluster {
                    id
                    name
                    projectSlug
                }
            }
            error {
                __typename
            }
        }
    }
""")


@pytest.fixture(scope="module")
def shared_test_cluster(graphql_client):
    """
    Ensure the default project exists with a cluster for all model tests.

    Model tests require a cluster (via the project's cluster sub-resource).
    This fixture ensures the default project exists and returns its cluster name.
    If the default project already exists, its cluster name is returned directly.
    """
    GET_PROJECT = gql("""
        query GetProject($name: String!) {
            project(name: $name) {
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

    result = graphql_client.execute(GET_PROJECT, variable_values={"name": DEFAULT_PROJECT_NAME})
    project_payload = result.get("project", {})
    project = project_payload.get("project")

    if project and project.get("cluster"):
        cluster_name = project["cluster"]["name"]
        print(f"\n✓ Using existing default project cluster: {cluster_name}")
        yield cluster_name
        return

    # Default project does not exist or has no cluster — create it
    cluster_name = f"{DEFAULT_PROJECT_NAME}-cluster"
    input_data = build_project_input(
        name=DEFAULT_PROJECT_NAME,
        title="Default Project",
        description="Default project for model testing",
        cluster_name=cluster_name,
        cluster_title="Default Cluster",
        skip_connection_test=True,
    )

    print(f"\n🔧 Creating default project with cluster: {cluster_name}")
    create_result = graphql_client.execute(CREATE_PROJECT, variable_values={"input": input_data})

    payload = create_result.get("createProject", {})
    if payload.get("error") is not None:
        error_type = payload["error"].get("__typename", "Unknown")
        # If already exists (race condition), try fetching again
        if error_type == "ProjectAlreadyExists":
            refetch = graphql_client.execute(GET_PROJECT, variable_values={"name": DEFAULT_PROJECT_NAME})
            proj = refetch.get("project", {}).get("project")
            if proj and proj.get("cluster"):
                cluster_name = proj["cluster"]["name"]
                print(f"✓ Default project already exists, using cluster: {cluster_name}")
                yield cluster_name
                return
        raise Exception(f"Failed to create default project: {error_type}")

    created_project = payload.get("project")
    if created_project and created_project.get("cluster"):
        cluster_name = created_project["cluster"]["name"]

    print(f"✓ Default project created with cluster: {cluster_name}")
    yield cluster_name
    # Default project is intentionally not deleted after tests

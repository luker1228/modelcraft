"""
Project-Cluster One-to-One Relationship Integration Tests

Tests for the one-to-one relationship between Project and Cluster.
In the new design, Cluster is a mandatory sub-resource created atomically
with the Project. The project.cluster field exposes the associated cluster.
"""

import pytest
from gql import gql

from design.common.test_data import build_project_input, generate_test_id
from design.common.assertions import assert_graphql_success


# GraphQL Queries and Mutations
CREATE_PROJECT = gql("""
    mutation CreateProject($input: CreateProjectInput!) {
        createProject(input: $input) {
            project {
                orgName
                name
                title
                cluster {
                    id
                    name
                    projectSlug
                    title
                }
            }
            error {
                __typename
                ... on ProjectAlreadyExists {
                    message
                }
                ... on InvalidProjectInput {
                    message
                }
                ... on DatabaseConnectionFailed {
                    message
                }
            }
        }
    }
""")

GET_PROJECT_WITH_CLUSTER = gql("""
    query GetProject($name: String!) {
        project(name: $name) {
            project {
                orgName
                name
                title
                cluster {
                    id
                    name
                    projectSlug
                    title
                }
            }
            error {
                __typename
                ... on ProjectNotFound {
                    message
                }
            }
        }
    }
""")

DELETE_PROJECT = gql("""
    mutation DeleteProject($name: String!) {
        deleteProject(name: $name) {
            success
            error {
                __typename
            }
        }
    }
""")

GET_CLUSTER = gql("""
    query GetCluster($projectSlug: String!) {
        databaseCluster(projectSlug: $projectSlug) {
            cluster {
                id
                name
                projectSlug
            }
            error {
                __typename
            }
        }
    }
""")


class TestProjectClusterOneToOne:
    """Test suite for Project-Cluster one-to-one relationship."""

    def test_create_project_has_cluster(self, graphql_client, created_projects):
        """Test creating a project always includes a cluster."""
        project_name = generate_test_id("proj-has-cluster")
        input_data = build_project_input(
            name=project_name,
            title="Project With Cluster",
            skip_connection_test=True,
        )

        result = graphql_client.execute(CREATE_PROJECT, variable_values={"input": input_data})

        assert_graphql_success(result, "createProject")
        project = result["createProject"]["project"]
        assert project["name"] == project_name
        assert project["cluster"] is not None, "Project must always have a cluster"
        assert project["cluster"]["projectSlug"] == project_name

        created_projects.append({"name": project["name"], "orgName": project["orgName"]})

    def test_project_cluster_created_atomically(self, graphql_client, created_projects):
        """Test that project and cluster are created in a single atomic operation."""
        project_name = generate_test_id("proj-atomic")
        cluster_name = f"{project_name}-db"
        input_data = build_project_input(
            name=project_name,
            title="Atomic Create Test",
            cluster_name=cluster_name,
            cluster_title="Atomic Cluster",
            skip_connection_test=True,
        )

        result = graphql_client.execute(CREATE_PROJECT, variable_values={"input": input_data})

        assert_graphql_success(result, "createProject")
        payload = result["createProject"]
        assert payload["error"] is None
        project = payload["project"]
        created_projects.append({"name": project["name"], "orgName": project["orgName"]})

        cluster = project["cluster"]
        assert cluster is not None
        assert cluster["name"] == cluster_name
        assert cluster["projectSlug"] == project_name

        # Verify cluster is accessible via databaseCluster query
        get_result = graphql_client.execute(GET_CLUSTER, variable_values={"projectSlug": project_name})
        assert_graphql_success(get_result, "databaseCluster")
        fetched = get_result["databaseCluster"]["cluster"]
        assert fetched is not None
        assert fetched["name"] == cluster_name

    def test_connection_failure_prevents_project_creation(self, graphql_client):
        """Test that bad connection info (without skipConnectionTest) prevents project creation."""
        project_name = generate_test_id("proj-conn-fail")
        input_data = build_project_input(
            name=project_name,
            title="Connection Failure Test",
            host="invalid.host.example",
            port=9999,
            username="baduser",
            password="badpass",
            skip_connection_test=False,
        )

        result = graphql_client.execute(CREATE_PROJECT, variable_values={"input": input_data})

        assert_graphql_success(result, "createProject")
        payload = result["createProject"]
        assert payload["project"] is None, "Project should not be created on connection failure"
        assert payload["error"] is not None
        assert payload["error"]["__typename"] == "DatabaseConnectionFailed"

        # Verify project was NOT created (atomic rollback)
        get_result = graphql_client.execute(GET_PROJECT_WITH_CLUSTER, variable_values={"name": project_name})
        project_payload = get_result.get("project", {})
        assert project_payload.get("project") is None or project_payload.get("error") is not None

    def test_get_project_with_cluster_field(self, graphql_client, created_projects):
        """Test retrieving project includes nested cluster field."""
        project_name = generate_test_id("proj-cluster-field")
        cluster_name = f"{project_name}-c"
        input_data = build_project_input(
            name=project_name,
            title="Project Cluster Field Test",
            cluster_name=cluster_name,
            cluster_title="Test Cluster",
            skip_connection_test=True,
        )

        project_result = graphql_client.execute(CREATE_PROJECT, variable_values={"input": input_data})
        assert_graphql_success(project_result, "createProject")
        project = project_result["createProject"]["project"]
        created_projects.append({"name": project["name"], "orgName": project["orgName"]})

        get_result = graphql_client.execute(GET_PROJECT_WITH_CLUSTER, variable_values={"name": project_name})

        assert_graphql_success(get_result, "project")
        payload = get_result["project"]
        retrieved_project = payload["project"]

        assert retrieved_project is not None
        assert retrieved_project["name"] == project_name
        # cluster is always present in new schema
        cluster = retrieved_project["cluster"]
        assert cluster is not None, "Project.cluster must be non-null"
        assert cluster["name"] == cluster_name
        assert cluster["projectSlug"] == project_name

    def test_delete_project_cascades_to_cluster(self, graphql_client):
        """Test that deleting a project also removes its cluster."""
        project_name = generate_test_id("proj-cascade-delete")
        input_data = build_project_input(
            name=project_name,
            title="Cascade Delete Test",
            skip_connection_test=True,
        )

        project_result = graphql_client.execute(CREATE_PROJECT, variable_values={"input": input_data})
        assert_graphql_success(project_result, "createProject")

        delete_result = graphql_client.execute(DELETE_PROJECT, variable_values={"name": project_name})
        assert_graphql_success(delete_result, "deleteProject")
        assert delete_result["deleteProject"]["success"] is True

        # Cluster should no longer exist
        get_result = graphql_client.execute(GET_CLUSTER, variable_values={"projectSlug": project_name})
        payload = get_result.get("databaseCluster", {})
        assert payload.get("cluster") is None or payload.get("error") is not None

    def test_skip_connection_test_creates_project_with_any_connection(self, graphql_client, created_projects):
        """Test that skipConnectionTest allows project creation even with invalid connection info."""
        project_name = generate_test_id("proj-skip-test")
        input_data = build_project_input(
            name=project_name,
            title="Skip Connection Test",
            host="unreachable.host.invalid",
            port=9999,
            username="noop",
            password="noop",
            skip_connection_test=True,
        )

        result = graphql_client.execute(CREATE_PROJECT, variable_values={"input": input_data})

        assert_graphql_success(result, "createProject")
        payload = result["createProject"]
        assert payload["error"] is None
        assert payload["project"] is not None
        assert payload["project"]["name"] == project_name
        assert payload["project"]["cluster"] is not None

        created_projects.append({
            "name": payload["project"]["name"],
            "orgName": payload["project"]["orgName"],
        })

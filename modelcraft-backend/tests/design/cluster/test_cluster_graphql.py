"""
Database Cluster GraphQL Tests

Tests for database cluster operations via GraphQL API.
Cluster is now a sub-resource of Project: created with the project, updated via
updateProjectCluster, and deleted when the project is deleted.
"""

import pytest
import uuid
from gql import gql

from design.common.test_data import build_project_input, generate_test_id
from design.common.assertions import assert_graphql_success
from design.common.db_config import get_test_db_config


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
                    description
                    connectionInfo {
                        host
                        port
                        username
                    }
                    status
                    createdAt
                }
            }
            error {
                __typename
                ... on ProjectAlreadyExists {
                    message
                    suggestion
                }
                ... on InvalidProjectInput {
                    message
                    suggestion
                }
                ... on DatabaseConnectionFailed {
                    message
                    suggestion
                }
            }
        }
    }
""")

GET_CLUSTER = gql("""
    query GetCluster($projectSlug: String!) {
        databaseCluster(projectSlug: $projectSlug) {
            cluster {
                id
                projectSlug
                name
                title
                connectionInfo {
                    host
                    port
                }
            }
            error {
                ... on ClusterNotFound {
                    message
                }
                ... on ProjectNotFound {
                    message
                }
            }
        }
    }
""")

UPDATE_PROJECT_CLUSTER = gql("""
    mutation UpdateProjectCluster($projectSlug: String!, $input: UpdateClusterConnectionInput!) {
        updateProjectCluster(projectSlug: $projectSlug, input: $input) {
            cluster {
                id
                name
                title
                connectionInfo {
                    host
                    port
                }
                updatedAt
            }
            error {
                ... on ClusterNotFound {
                    message
                }
                ... on InvalidClusterInput {
                    message
                    suggestion
                }
                ... on DatabaseConnectionFailed {
                    message
                    suggestion
                }
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

GET_PROJECT = gql("""
    query GetProject($name: String!) {
        project(name: $name) {
            project {
                name
                cluster {
                    id
                    name
                    projectSlug
                    title
                }
            }
            error {
                __typename
            }
        }
    }
""")


def unique_name(prefix: str = "test") -> str:
    """Generate a unique name to avoid conflicts."""
    return f"{prefix}-{uuid.uuid4().hex[:8]}"


class TestClusterCRUD:
    """Test suite for database cluster operations via project sub-resource API."""

    @pytest.mark.parametrize("cluster_suffix", [
        "alpha",
        "beta",
    ])
    def test_create_project_with_cluster_success(self, graphql_client, created_projects, cluster_suffix):
        """Test creating a project creates the cluster atomically."""
        project_name = generate_test_id(f"proj-{cluster_suffix}")
        input_data = build_project_input(
            name=project_name,
            title=f"Project {cluster_suffix}",
            cluster_name=f"{project_name}-cluster",
            cluster_title=f"Cluster {cluster_suffix}",
            skip_connection_test=True,
        )

        result = graphql_client.execute(CREATE_PROJECT, variable_values={"input": input_data})

        assert_graphql_success(result, "createProject")
        payload = result["createProject"]
        assert payload["error"] is None, f"Unexpected error: {payload.get('error')}"
        project = payload["project"]
        assert project is not None
        assert project["name"] == project_name
        cluster = project["cluster"]
        assert cluster is not None
        assert cluster["projectSlug"] == project_name
        assert cluster["name"] == f"{project_name}-cluster"

        created_projects.append({"name": project["name"], "orgName": project["orgName"]})

    def test_create_project_with_invalid_connection(self, graphql_client, created_projects):
        """Test that invalid connection info returns DatabaseConnectionFailed (without skipConnectionTest)."""
        project_name = generate_test_id("proj-bad-conn")
        input_data = build_project_input(
            name=project_name,
            title="Project with bad connection",
            host="invalid.host.example",
            port=9999,
            username="invalid",
            password="invalid",
            skip_connection_test=False,
        )

        result = graphql_client.execute(CREATE_PROJECT, variable_values={"input": input_data})

        assert_graphql_success(result, "createProject")
        payload = result["createProject"]
        assert payload["project"] is None
        assert payload["error"] is not None
        error = payload["error"]
        assert error["__typename"] == "DatabaseConnectionFailed"
        assert "message" in error

    def test_get_cluster_success(self, graphql_client, created_projects):
        """Test retrieving a cluster by project name."""
        project_name = generate_test_id("proj-get-cluster")
        input_data = build_project_input(
            name=project_name,
            title="Project for Get Cluster Test",
            skip_connection_test=True,
        )

        project_result = graphql_client.execute(CREATE_PROJECT, variable_values={"input": input_data})
        assert_graphql_success(project_result, "createProject")
        project = project_result["createProject"]["project"]
        created_projects.append({"name": project["name"], "orgName": project["orgName"]})
        cluster_name = project["cluster"]["name"]

        result = graphql_client.execute(GET_CLUSTER, variable_values={"projectSlug": project_name})

        assert_graphql_success(result, "databaseCluster")
        payload = result["databaseCluster"]
        assert payload["error"] is None, f"Unexpected error: {payload.get('error')}"
        retrieved_cluster = payload["cluster"]
        assert retrieved_cluster["name"] == cluster_name
        assert retrieved_cluster["projectSlug"] == project_name

    def test_get_cluster_project_not_found(self, graphql_client):
        """Test retrieving cluster for a non-existent project."""
        nonexistent_project = generate_test_id("proj-nonexistent")

        try:
            result = graphql_client.execute(GET_CLUSTER, variable_values={"projectSlug": nonexistent_project})
            payload = result.get("databaseCluster", {})
            assert payload.get("cluster") is None or payload.get("error") is not None
        except Exception as e:
            assert "not found" in str(e).lower()

    @pytest.mark.slow
    def test_update_project_cluster_success(self, graphql_client, created_projects):
        """Test updating the project's cluster connection info."""
        project_name = generate_test_id("proj-update-cluster")
        input_data = build_project_input(
            name=project_name,
            title="Project for Update Cluster Test",
            skip_connection_test=True,
        )

        project_result = graphql_client.execute(CREATE_PROJECT, variable_values={"input": input_data})
        assert_graphql_success(project_result, "createProject")
        project = project_result["createProject"]["project"]
        created_projects.append({"name": project["name"], "orgName": project["orgName"]})

        # Update the cluster title only (no connection change needed)
        update_input = {
            "title": "Updated Cluster Title",
            "skipConnectionTest": True,
        }
        result = graphql_client.execute(UPDATE_PROJECT_CLUSTER, variable_values={
            "projectSlug": project_name,
            "input": update_input,
        })

        assert_graphql_success(result, "updateProjectCluster")
        payload = result["updateProjectCluster"]
        assert payload["error"] is None, f"Unexpected error: {payload.get('error')}"
        updated_cluster = payload["cluster"]
        assert updated_cluster["title"] == "Updated Cluster Title"

    @pytest.mark.slow
    def test_delete_project_removes_cluster(self, graphql_client, created_projects):
        """Test that deleting a project also removes its cluster."""
        project_name = generate_test_id("proj-delete-cascade")
        input_data = build_project_input(
            name=project_name,
            title="Project for Delete Test",
            skip_connection_test=True,
        )

        project_result = graphql_client.execute(CREATE_PROJECT, variable_values={"input": input_data})
        assert_graphql_success(project_result, "createProject")
        project = project_result["createProject"]["project"]
        # Not adding to created_projects since we're deleting manually

        result = graphql_client.execute(DELETE_PROJECT, variable_values={"name": project_name})

        assert_graphql_success(result, "deleteProject")
        delete_result = result["deleteProject"]
        assert delete_result["success"] is True
        assert delete_result["error"] is None

        # Cluster should no longer be retrievable
        get_result = graphql_client.execute(GET_CLUSTER, variable_values={"projectSlug": project_name})
        payload = get_result.get("databaseCluster", {})
        assert payload.get("cluster") is None or payload.get("error") is not None

    def test_get_project_includes_cluster(self, graphql_client, created_projects):
        """Test that querying a project includes the nested cluster."""
        project_name = generate_test_id("proj-with-cluster")
        cluster_name = f"{project_name}-cluster"
        input_data = build_project_input(
            name=project_name,
            title="Project With Cluster",
            cluster_name=cluster_name,
            cluster_title="Nested Cluster",
            skip_connection_test=True,
        )

        project_result = graphql_client.execute(CREATE_PROJECT, variable_values={"input": input_data})
        assert_graphql_success(project_result, "createProject")
        project = project_result["createProject"]["project"]
        created_projects.append({"name": project["name"], "orgName": project["orgName"]})

        result = graphql_client.execute(GET_PROJECT, variable_values={"name": project_name})

        assert_graphql_success(result, "project")
        retrieved_project = result["project"]["project"]
        assert retrieved_project is not None
        assert retrieved_project["name"] == project_name
        cluster = retrieved_project["cluster"]
        assert cluster is not None
        assert cluster["name"] == cluster_name
        assert cluster["projectSlug"] == project_name

    def test_create_project_with_skip_connection_test(self, graphql_client, created_projects):
        """Test that skipConnectionTest: true bypasses validation with invalid credentials."""
        project_name = generate_test_id("proj-skip-conn")
        # Use invalid connection info but skip the test
        input_data = build_project_input(
            name=project_name,
            title="Project Skip Connection Test",
            host="unreachable.host.invalid",
            port=9999,
            username="noop",
            password="noop",
            skip_connection_test=True,
        )

        result = graphql_client.execute(CREATE_PROJECT, variable_values={"input": input_data})

        assert_graphql_success(result, "createProject")
        payload = result["createProject"]
        assert payload["error"] is None, f"Unexpected error: {payload.get('error')}"
        assert payload["project"] is not None
        assert payload["project"]["name"] == project_name

        created_projects.append({
            "name": payload["project"]["name"],
            "orgName": payload["project"]["orgName"],
        })

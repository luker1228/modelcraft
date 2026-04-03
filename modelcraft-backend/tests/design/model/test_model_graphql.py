"""
Model Design GraphQL Tests

Tests for model CRUD operations and field management via GraphQL API.
"""

import pytest
import uuid
from gql import gql

from design.common.assertions import assert_graphql_success


# Constants
DEFAULT_ORG_NAME = "built-in"
DEFAULT_PROJECT_NAME = "default"
DEFAULT_DB_NAME = "test_db"


def unique_name(prefix: str = "test") -> str:
    """Generate a unique name to avoid conflicts."""
    return f"{prefix}-{uuid.uuid4().hex[:8]}"

CREATE_MODEL = gql("""
    mutation CreateModel($input: CreateModelInput!) {
        createModel(input: $input) {
            model {
                id
                projectSlug
                name
                title
                description
                clusterName
                databaseName
            }
            error {
                __typename
                ... on ModelAlreadyExists {
                    message
                }
                ... on InvalidModelInput {
                    message
                }
                ... on ProjectNotFound {
                    message
                }
            }
        }
    }
""")

GET_MODEL = gql("""
    query GetModel($projectSlug: String!, $id: ID!) {
        model(projectSlug: $projectSlug, id: $id) {
            model {
                id
                name
                title
                projectSlug
                clusterName
                databaseName
            }
            error {
                __typename
                ... on ModelNotFound {
                    message
                }
                ... on InvalidModelInput {
                    message
                }
                ... on ProjectNotFound {
                    message
                }
            }
        }
    }
""")

LIST_MODELS = gql("""
    query ListModels($input: ModelQueryInput) {
        models(input: $input) {
            edges {
                node {
                    id
                    name
                    projectSlug
                    clusterName
                    databaseName
                }
                cursor
            }
            pageInfo {
                hasNextPage
                hasPreviousPage
            }
            totalCount
        }
    }
""")


def build_model_input(name, cluster_name, project_name=DEFAULT_PROJECT_NAME, **kwargs):
    """Build model input data with defaults"""
    return {
        "projectSlug": project_name,
        "name": name,
        "title": kwargs.get("title", f"Test Model {name}"),
        "description": kwargs.get("description", "Test model description"),
        "clusterName": cluster_name,
        "databaseName": kwargs.get("databaseName", DEFAULT_DB_NAME)
    }


class TestModelCRUD:
    """Test suite for model CRUD operations"""

    @pytest.fixture(scope="function")
    def test_cluster(self, shared_test_cluster):
        """
        Provide the shared test cluster to individual tests.
        Uses session-scoped shared_test_cluster from test_model_errors.py
        to avoid cluster creation timing issues.
        """
        return shared_test_cluster

    def test_create_model_success(self, graphql_client, test_cluster, created_models, default_project):
        """Test creating a new model"""
        # Arrange
        model_name = unique_name("User")
        input_data = build_model_input(
            name=model_name,
            cluster_name=test_cluster,
            project_name=default_project["name"],
            title="User Model",
            description="User model for testing"
        )

        # Act
        result = graphql_client.execute(CREATE_MODEL, variable_values={"input": input_data})

        # Assert
        assert_graphql_success(result, "createModel")
        payload = result["createModel"]
        assert payload["error"] is None, f"Unexpected error: {payload.get('error')}"
        model = payload["model"]
        assert model is not None
        assert model["name"] == model_name
        assert model["title"] == "User Model"
        assert model["projectSlug"] == default_project["name"]

        # Track for cleanup
        created_models.append((model["projectSlug"], model["id"]))

    def test_get_model_success(self, graphql_client, test_cluster, created_models, default_project):
        """Test retrieving a model by id"""
        # Arrange: Create a model first
        model_name = unique_name("Product")
        input_data = build_model_input(
            name=model_name,
            cluster_name=test_cluster,
            project_name=default_project["name"])
        create_result = graphql_client.execute(CREATE_MODEL, variable_values={"input": input_data})
        assert create_result["createModel"]["error"] is None
        created_model = create_result["createModel"]["model"]
        created_models.append((created_model["projectSlug"], created_model["id"]))

        # Act
        result = graphql_client.execute(GET_MODEL, variable_values={
            "projectSlug": default_project["name"],
            "id": created_model["id"]
        })

        # Assert
        assert_graphql_success(result, "model")
        payload = result["model"]
        assert payload["error"] is None, f"Unexpected error: {payload.get('error')}"
        model = payload["model"]
        assert model["name"] == model_name
        assert model["projectSlug"] == default_project["name"]

    def test_list_models_for_project(self, graphql_client, test_cluster, created_models, default_project):
        """Test listing models for a specific project"""
        # Arrange: Create some test models
        created_names = []
        for i in range(2):
            model_name = unique_name(f"list-model-{i}")
            created_names.append(model_name)
            input_data = build_model_input(
                name=model_name,
                cluster_name=test_cluster,
                project_name=default_project["name"])
            create_result = graphql_client.execute(CREATE_MODEL, variable_values={"input": input_data})
            assert create_result["createModel"]["error"] is None
            model = create_result["createModel"]["model"]
            created_models.append((model["projectSlug"], model["id"]))

        # Act
        result = graphql_client.execute(LIST_MODELS, variable_values={
            "input": {
                "projectSlug": default_project["name"],
                "clusterName": test_cluster,
                "databaseName": DEFAULT_DB_NAME
            }
        })

        # Assert
        assert_graphql_success(result, "models")
        connection = result["models"]
        assert "edges" in connection
        assert "totalCount" in connection
        assert connection["totalCount"] >= 2  # Should include our test models

        # Verify our created models are in the list
        model_names = [edge["node"]["name"] for edge in connection["edges"]]
        for created_name in created_names:
            assert created_name in model_names

    def test_create_model_duplicate_name_error(self, graphql_client, test_cluster, created_models, default_project):
        """Test that creating a model with duplicate name returns error"""
        # Arrange: Create first model
        model_name = unique_name("DuplicateTest")
        input_data = build_model_input(
            name=model_name,
            cluster_name=test_cluster,
            project_name=default_project["name"])
        first_result = graphql_client.execute(CREATE_MODEL, variable_values={"input": input_data})
        assert first_result["createModel"]["error"] is None
        model = first_result["createModel"]["model"]
        created_models.append((model["projectSlug"], model["id"]))

        # Act: Try to create model with same name
        duplicate_result = graphql_client.execute(CREATE_MODEL, variable_values={"input": input_data})

        # Assert
        assert_graphql_success(duplicate_result, "createModel")
        payload = duplicate_result["createModel"]
        assert payload["error"] is not None
        assert payload["error"]["__typename"] == "ModelAlreadyExists"
        assert payload["model"] is None

    def test_get_model_not_found(self, graphql_client, default_project):
        """Test that getting non-existent model returns error"""
        # Act
        result = graphql_client.execute(GET_MODEL, variable_values={
            "projectSlug": default_project["name"],
            "id": "999999"
        })

        # Assert
        assert_graphql_success(result, "model")
        payload = result["model"]
        assert payload["error"] is not None
        assert payload["error"]["__typename"] == "ModelNotFound"
        assert payload["model"] is None

    def test_list_models_empty_result(self, graphql_client, default_project):
        """Test listing models returns empty result when no models exist"""
        # Use a unique cluster/database combination that shouldn't exist
        nonexistent_cluster = unique_name("nonexistent-cluster")

        # Act
        result = graphql_client.execute(LIST_MODELS, variable_values={
            "input": {
                "projectSlug": default_project["name"],
                "clusterName": nonexistent_cluster,
                "databaseName": "nonexistent_db"
            }
        })

        # Assert
        assert_graphql_success(result, "models")
        connection = result["models"]
        assert connection["totalCount"] == 0
        assert len(connection["edges"]) == 0
        assert connection["pageInfo"]["hasNextPage"] is False

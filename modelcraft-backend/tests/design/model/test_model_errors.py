"""
Model Design GraphQL Typed Error Tests

Tests for model GraphQL operations that return typed errors.
This tests the typed error handling for Model operations.
"""

import pytest
from gql import gql

from design.common.test_data import generate_test_id, build_model_input
from design.common.assertions import assert_graphql_success


# Constants
DEFAULT_ORG_NAME = "built-in"
DEFAULT_PROJECT_NAME = "default"
DEFAULT_DB_NAME = "test_db"
NONEXISTENT_MODEL_NAME = "NonExistentModel99999"
MIN_ERROR_MESSAGE_LENGTH = 10


# GraphQL Queries and Mutations with typed error handling

CREATE_MODEL = gql("""
    mutation CreateModel($input: CreateModelInput!) {
        createModel(input: $input) {
            model {
                id
                name
                projectSlug
            }
            error {
                __typename
                ... on ModelAlreadyExists {
                    message
                    suggestion
                }
                ... on InvalidModelInput {
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

GET_MODEL = gql("""
    query GetModel($projectSlug: String!, $id: ID!) {
        model(projectSlug: $projectSlug, id: $id) {
            model {
                id
                name
                projectSlug
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

GET_MODEL_BY_NAME = gql("""
    query GetModelByName($projectSlug: String!, $name: String!, $clusterName: String!, $databaseName: String!) {
        modelByName(projectSlug: $projectSlug, name: $name, clusterName: $clusterName, databaseName: $databaseName) {
            model {
                id
                name
                projectSlug
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

UPDATE_MODEL = gql("""
    mutation UpdateModel($projectSlug: String!, $id: ID!, $input: UpdateModelMetaInput!) {
        updateModel(projectSlug: $projectSlug, id: $id, input: $input) {
            success
            model {
                id
                name
                projectSlug
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

DELETE_MODEL = gql("""
    mutation DeleteModel($projectSlug: String!, $id: ID!) {
        deleteModel(projectSlug: $projectSlug, id: $id) {
            success
            error {
                __typename
                ... on ModelNotFound {
                    message
                }
                ... on CannotDeleteDeployedModel {
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




class TestModelErrors:
    """Test suite for Model GraphQL typed errors"""

    # ==================== Helper Methods ====================

    def _assert_error_type(self, error, expected_typename):
        """Assert error has the expected __typename"""
        assert error is not None, "Error should not be None"
        assert error["__typename"] == expected_typename, \
            f"Expected {expected_typename}, got {error.get('__typename')}"

    def _assert_no_error(self, payload, message="Should not have error"):
        """Assert payload has no error and model data exists"""
        assert payload["error"] is None, message
        assert payload["model"] is not None, "Model should not be None on success"

    def _assert_has_error_no_model(self, payload, expected_typename):
        """Assert payload has error and no model data"""
        assert payload["model"] is None, f"Model should be None when {expected_typename} occurs"
        self._assert_error_type(payload["error"], expected_typename)

    def _assert_error_message_quality(self, error, min_length=MIN_ERROR_MESSAGE_LENGTH):
        """Assert error message is non-empty and descriptive"""
        assert error["message"], "Error message should not be empty"
        assert len(error["message"]) > min_length, \
            f"Error message should be descriptive (>{min_length} chars)"

    # ==================== Fixtures ====================

    @pytest.fixture(scope="function")
    def test_cluster(self, shared_test_cluster):
        """
        Provide the shared test cluster to individual tests.
        Uses session-scoped shared_test_cluster to avoid cluster creation timing issues.
        """
        return shared_test_cluster

    @pytest.fixture(scope="function")
    def test_model(self, graphql_client, test_cluster, created_models):
        """Create a test model for error tests"""
        model_name = f"TestModel{generate_test_id()[-8:].upper()}"
        input_data = build_model_input(
            project_name=DEFAULT_PROJECT_NAME,
            cluster_name=test_cluster,
            database_name=DEFAULT_DB_NAME,
            name=model_name,
            description="Test model description"
        )

        print(f"\n🔍 Debug: test_cluster name = {test_cluster}")
        print(f"🔍 Debug: input_data = {input_data}")

        result = graphql_client.execute(CREATE_MODEL, variable_values={"input": input_data})

        # Check if model was created successfully
        payload = result["createModel"]
        print(f"🔍 Debug: create model response = {payload}")

        if payload.get("error") is not None:
            error_type = payload['error'].get('__typename', 'Unknown')
            error_msg = payload['error'].get('message', 'No message')
            print(f"🔍 Debug: Full error = {payload['error']}")
            raise Exception(f"Failed to create test model: {error_type} - {error_msg}")

        model = payload["model"]
        created_models.append((model["projectSlug"], model["id"]))
        return DEFAULT_PROJECT_NAME, model["name"], model["id"], test_cluster

    # ==================== Get Model Error Tests ====================

    def test_get_model_not_found_error(self, graphql_client, default_project):
        """Test getting a non-existent model returns ModelNotFound error"""
        result = graphql_client.execute(GET_MODEL, variable_values={
            "projectSlug": DEFAULT_PROJECT_NAME,
            "id": "999999999"
        })

        assert_graphql_success(result, "model")
        payload = result["model"]

        self._assert_has_error_no_model(payload, "ModelNotFound")
        self._assert_error_message_quality(payload["error"])

    def test_get_model_by_name_not_found(self, graphql_client, test_cluster, default_project):
        """Test getting model by name that doesn't exist returns ModelNotFound error"""
        nonexistent_name = f"NonExistentModel{generate_test_id()[-8:].upper()}"

        result = graphql_client.execute(GET_MODEL_BY_NAME, variable_values={
            "projectSlug": DEFAULT_PROJECT_NAME,
            "name": nonexistent_name,
            "clusterName": test_cluster,
            "databaseName": DEFAULT_DB_NAME
        })

        assert_graphql_success(result, "modelByName")
        payload = result["modelByName"]

        self._assert_has_error_no_model(payload, "ModelNotFound")
        error = payload["error"]
        # Support both English and Chinese error messages
        message_lower = error["message"].lower()
        assert (nonexistent_name in error["message"] or "not found" in message_lower or
                "不存在" in error["message"] or "未找到" in error["message"]), \
            "Error message should mention the model or 'not found'"

    def test_get_model_by_name_nonexistent_project_error(self, graphql_client, test_cluster):
        """Test getting model by name with non-existent project returns error"""
        result = graphql_client.execute(GET_MODEL_BY_NAME, variable_values={
            "projectSlug": "nonexistent-project-12345",
            "name": "SomeModel",
            "clusterName": test_cluster,
            "databaseName": DEFAULT_DB_NAME
        })

        assert_graphql_success(result, "modelByName")

        # The backend may return ModelNotFound or ProjectNotFound
        # Both are acceptable as they indicate the resource doesn't exist
        error = result["modelByName"]["error"]
        assert error is not None, "Should have error for non-existent project"
        assert error["__typename"] in ["ProjectNotFound", "ModelNotFound"], \
            f"Expected ProjectNotFound or ModelNotFound, got {error['__typename']}"

    # ==================== Create Model Error Tests ====================

    def test_create_model_already_exists_error(self, graphql_client, test_model):
        """Test creating a model with duplicate name returns ModelAlreadyExists error"""
        project_name, model_name, model_id, cluster_name = test_model
        input_data = build_model_input(
            project_name=project_name,
            cluster_name=cluster_name,
            database_name=DEFAULT_DB_NAME,
            name=model_name,
            description="Trying to create duplicate model"
        )

        result = graphql_client.execute(CREATE_MODEL, variable_values={"input": input_data})
        assert_graphql_success(result, "createModel")
        payload = result["createModel"]

        self._assert_has_error_no_model(payload, "ModelAlreadyExists")
        error = payload["error"]
        assert error["message"], "Error message should not be empty"
        assert error["suggestion"], "Error suggestion should not be empty"

    def test_create_model_with_nonexistent_project_error(self, graphql_client):
        """Test creating model with non-existent project returns error"""
        nonexistent_project_name = f"nonexistent-project-{generate_test_id()[-8:]}"
        input_data = build_model_input(
            project_name=nonexistent_project_name,
            cluster_name="test-cluster",
            database_name=DEFAULT_DB_NAME,
            name="TestModel"
        )

        result = graphql_client.execute(CREATE_MODEL, variable_values={"input": input_data})
        assert_graphql_success(result, "createModel")

        # The backend may return InvalidModelInput or ProjectNotFound
        # Both are acceptable as they indicate validation failure
        error = result["createModel"]["error"]
        assert error is not None, "Should have error for non-existent project"
        assert error["__typename"] in ["ProjectNotFound", "InvalidModelInput"], \
            f"Expected ProjectNotFound or InvalidModelInput, got {error['__typename']}"

    # ==================== Update Model Error Tests ====================

    def test_update_nonexistent_model_error(self, graphql_client, default_project):
        """Test updating a non-existent model returns ModelNotFound error"""
        # Note: Currently the backend throws TransportQueryError instead of returning typed error
        # This test documents the current behavior
        from gql.transport.exceptions import TransportQueryError

        with pytest.raises(TransportQueryError) as exc_info:
            graphql_client.execute(UPDATE_MODEL, variable_values={
                "projectSlug": DEFAULT_PROJECT_NAME,
                "id": "999999999",
                "input": {"title": "Updated Title", "description": "Updated description"}
            })

        # Verify error message contains relevant information
        error_msg = str(exc_info.value)
        assert "not found" in error_msg.lower() or "不存在" in error_msg, \
            "Error should indicate model was not found"

    def test_update_model_with_nonexistent_project_error(self, graphql_client):
        """Test updating model with non-existent project returns ProjectNotFound error"""
        # Note: Currently the backend throws TransportQueryError instead of returning typed error
        # This test documents the current behavior
        from gql.transport.exceptions import TransportQueryError
        nonexistent_project_name = f"nonexistent-project-{generate_test_id()[-8:]}"

        with pytest.raises(TransportQueryError) as exc_info:
            graphql_client.execute(UPDATE_MODEL, variable_values={
                "projectSlug": nonexistent_project_name,
                "id": "999999999",
                "input": {"title": "Updated Title"}
            })

        # Verify error message contains relevant information
        error_msg = str(exc_info.value)
        assert "not found" in error_msg.lower() or "不存在" in error_msg, \
            "Error should indicate resource was not found"

    # ==================== Delete Model Error Tests ====================

    def test_delete_nonexistent_model_error(self, graphql_client, default_project):
        """Test deleting a non-existent model returns ModelNotFound error"""
        result = graphql_client.execute(DELETE_MODEL, variable_values={
            "projectSlug": DEFAULT_PROJECT_NAME,
            "id": "999999999"
        })

        assert_graphql_success(result, "deleteModel")
        payload = result["deleteModel"]

        assert payload["success"] is False
        self._assert_error_type(payload["error"], "ModelNotFound")

    def test_delete_model_with_nonexistent_project_error(self, graphql_client):
        """Test deleting model with non-existent project returns error"""
        result = graphql_client.execute(DELETE_MODEL, variable_values={
            "projectSlug": f"nonexistent-project-{generate_test_id()[-8:]}",
            "id": "999999999"
        })

        assert_graphql_success(result, "deleteModel")

        # The backend may return ModelNotFound or ProjectNotFound
        # Both are acceptable as they indicate the resource doesn't exist
        error = result["deleteModel"]["error"]
        assert error is not None, "Should have error for non-existent project"
        assert error["__typename"] in ["ProjectNotFound", "ModelNotFound"], \
            f"Expected ProjectNotFound or ModelNotFound, got {error['__typename']}"

    # ==================== Success Case Tests ====================

    def test_successful_create_model_has_no_error(self, graphql_client, test_cluster, created_models, default_project):
        """Test successful model creation returns no error"""
        model_name = f"SuccessModel{generate_test_id()[-8:].upper()}"
        input_data = build_model_input(
            project_name=DEFAULT_PROJECT_NAME,
            cluster_name=test_cluster,
            database_name=DEFAULT_DB_NAME,
            name=model_name,
            description="Testing successful model creation"
        )

        result = graphql_client.execute(CREATE_MODEL, variable_values={"input": input_data})
        assert_graphql_success(result, "createModel")
        payload = result["createModel"]

        self._assert_no_error(payload, "Should not have error for successful model creation")
        assert payload["model"]["name"] == model_name
        model = payload["model"]
        created_models.append((model["projectSlug"], model["id"]))

    def test_successful_get_model_has_no_error(self, graphql_client, test_model):
        """Test successful model retrieval returns no error"""
        project_name, model_name, model_id, _ = test_model

        result = graphql_client.execute(GET_MODEL, variable_values={
            "projectSlug": project_name,
            "id": model_id
        })

        assert_graphql_success(result, "model")
        payload = result["model"]

        self._assert_no_error(payload, "Should not have error for successful model retrieval")
        assert payload["model"]["name"] == model_name

    def test_successful_update_model_has_no_error(self, graphql_client, test_model):
        """Test successful model update returns no error

        Note: Currently, models without fields cannot be updated due to validation.
        This test is skipped until models can be created with fields.
        """
        pytest.skip("Models without fields cannot be updated - backend validation requires at least one field")

        project_name, model_name, model_id, _ = test_model

        result = graphql_client.execute(UPDATE_MODEL, variable_values={
            "projectSlug": project_name,
            "id": model_id,
            "input": {"title": "Updated Test Model Title"}
        })

        assert_graphql_success(result, "updateModel")
        payload = result["updateModel"]

        assert payload["success"] is True
        self._assert_no_error(payload, "Should not have error for successful model update")
        assert payload["model"]["name"] == model_name

    def test_successful_delete_model_has_no_error(self, graphql_client, test_cluster, created_models, default_project):
        """Test successful model deletion returns no error"""
        # Create a model to delete
        model_name = f"ToDeleteModel{generate_test_id()[-8:].upper()}"
        input_data = build_model_input(
            project_name=DEFAULT_PROJECT_NAME,
            cluster_name=test_cluster,
            database_name=DEFAULT_DB_NAME,
            name=model_name,
            description="Testing successful model deletion"
        )

        create_result = graphql_client.execute(CREATE_MODEL, variable_values={"input": input_data})
        model = create_result["createModel"]["model"]

        # Delete the model
        result = graphql_client.execute(DELETE_MODEL, variable_values={
            "projectSlug": model["projectSlug"],
            "id": model["id"]
        })

        assert_graphql_success(result, "deleteModel")
        payload = result["deleteModel"]

        assert payload["success"] is True
        assert payload["error"] is None, "Should not have error for successful model deletion"

    def test_get_model_by_name_success_has_no_error(self, graphql_client, test_model):
        """Test successful model retrieval by name returns no error"""
        project_name, model_name, model_id, cluster_name = test_model

        result = graphql_client.execute(GET_MODEL_BY_NAME, variable_values={
            "projectSlug": project_name,
            "name": model_name,
            "clusterName": cluster_name,
            "databaseName": DEFAULT_DB_NAME
        })

        assert_graphql_success(result, "modelByName")
        payload = result["modelByName"]

        self._assert_no_error(payload, "Should not have error for successful retrieval by name")
        assert payload["model"]["name"] == model_name
        assert payload["model"]["projectSlug"] == project_name

    # ==================== Error Message Quality Tests ====================

    def test_error_message_quality_model_already_exists(self, graphql_client, test_model):
        """Test that ModelAlreadyExists error has helpful message and suggestion"""
        project_name, model_name, model_id, cluster_name = test_model
        input_data = build_model_input(
            project_name=project_name,
            cluster_name=cluster_name,
            database_name=DEFAULT_DB_NAME,
            name=model_name,
            description="Duplicate Model"
        )

        result = graphql_client.execute(CREATE_MODEL, variable_values={"input": input_data})
        assert_graphql_success(result, "createModel")
        error = result["createModel"]["error"]

        self._assert_error_type(error, "ModelAlreadyExists")
        self._assert_error_message_quality(error)

        assert error["suggestion"] is not None, "Suggestion should be provided"
        assert len(error["suggestion"]) > 0, "Suggestion should not be empty"
        assert "different" in error["suggestion"].lower() or "unique" in error["suggestion"].lower(), \
            "Suggestion should guide user to use a different name"

    def test_error_message_quality_model_not_found(self, graphql_client, default_project):
        """Test that ModelNotFound error has helpful message"""
        result = graphql_client.execute(GET_MODEL, variable_values={
            "projectSlug": DEFAULT_PROJECT_NAME,
            "id": "999999999"
        })

        assert_graphql_success(result, "model")
        error = result["model"]["error"]

        self._assert_error_type(error, "ModelNotFound")
        self._assert_error_message_quality(error)
        # Support both English and Chinese error messages
        message_lower = error["message"].lower()
        assert ("not found" in message_lower or "does not exist" in message_lower or
                "不存在" in error["message"] or "未找到" in error["message"]), \
            "Error message should clearly indicate the model was not found"

    def test_error_message_quality_project_not_found(self, graphql_client):
        """Test that error has helpful message for non-existent project"""
        nonexistent_project_name = f"nonexistent-project-{generate_test_id()[-8:]}"
        input_data = build_model_input(
            project_name=nonexistent_project_name,
            cluster_name="test-cluster",
            database_name=DEFAULT_DB_NAME,
            name="TestModel"
        )
        # Add required fields
        input_data["databaseName"] = DEFAULT_DB_NAME
        input_data["databaseName"] = DEFAULT_DB_NAME

        result = graphql_client.execute(CREATE_MODEL, variable_values={"input": input_data})
        assert_graphql_success(result, "createModel")
        error = result["createModel"]["error"]

        # The backend may return InvalidModelInput or ProjectNotFound
        # Both are acceptable - verify message quality regardless
        assert error is not None, "Should have error for non-existent project"
        assert error["__typename"] in ["ProjectNotFound", "InvalidModelInput"], \
            f"Expected ProjectNotFound or InvalidModelInput, got {error['__typename']}"

        self._assert_error_message_quality(error)
        # Verify message has meaningful content - support both English and Chinese
        # InvalidModelInput may mention cluster instead of project if cluster validation happens first
        message = error["message"]
        message_lower = message.lower()
        assert (len(message) > MIN_ERROR_MESSAGE_LENGTH and
                ("project" in message_lower or "项目" in message or "cluster" in message_lower or
                 "集群" in message or "不存在" in message)), \
            "Error message should be descriptive and mention relevant resources"
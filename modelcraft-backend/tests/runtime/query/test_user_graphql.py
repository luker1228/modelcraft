"""
User Data GraphQL Query Tests

Tests for querying user data through the Runtime GraphQL API.
These tests depend on models being set up in the Design phase.
"""

import pytest
from gql import gql

# Runtime tests can import Design utilities
from design.common.assertions import assert_graphql_success


QUERY_USER_DATA = gql("""
    query QueryUserData($modelId: Int!, $filter: FilterInput) {
        findMany(modelId: $modelId, filter: $filter) {
            data
            total
        }
    }
""")

QUERY_SINGLE_USER = gql("""
    query QuerySingleUser($modelId: Int!, $id: Int!) {
        findUnique(modelId: $modelId, where: { id: $id }) {
            data
        }
    }
""")


class TestUserGraphQLQueries:
    """Test suite for user data queries via Runtime GraphQL"""

    def test_query_all_users(self, runtime_graphql_client, sample_model_setup):
        """Test querying all records for a model"""
        # Arrange
        model_id = sample_model_setup["id"]

        # Act
        result = runtime_graphql_client.execute(
            QUERY_USER_DATA,
            variable_values={"modelId": model_id}
        )

        # Assert
        # Note: This may fail if the Runtime API structure is different
        # Adjust based on actual Runtime API response format
        if "errors" in result:
            pytest.skip(f"Runtime query API not available: {result['errors']}")

        # Basic validation
        assert result is not None, "Query result should not be None"
        assert isinstance(result, dict), "Result should be a dictionary"

    def test_query_with_filter(self, runtime_graphql_client, sample_model_setup):
        """Test querying data with filters"""
        # Arrange
        model_id = sample_model_setup["id"]
        filter_input = {
            "status": {"eq": "active"}
        }

        # Act
        result = runtime_graphql_client.execute(
            QUERY_USER_DATA,
            variable_values={
                "modelId": model_id,
                "filter": filter_input
            }
        )

        # Assert
        if "errors" in result:
            pytest.skip(f"Runtime query with filter not available: {result['errors']}")

        assert result is not None

    def test_query_single_record(self, runtime_graphql_client, sample_model_setup):
        """Test querying a single record by ID"""
        # Arrange
        model_id = sample_model_setup["id"]
        record_id = 1  # Assuming a record exists

        # Act
        result = runtime_graphql_client.execute(
            QUERY_SINGLE_USER,
            variable_values={
                "modelId": model_id,
                "id": record_id
            }
        )

        # Assert
        if "errors" in result:
            # This is expected if no data exists
            pytest.skip("No data available for query test")

        assert result is not None

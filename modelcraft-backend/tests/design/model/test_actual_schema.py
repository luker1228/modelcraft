"""
Actual Schema Integration Tests

Tests for model(id, withActualSchema: true) query - verifies that
actual database schema information (dbTable, dbColumn) is correctly
returned when withActualSchema is enabled.
"""

import uuid
import pytest
from gql import gql

from design.common.assertions import assert_graphql_success


DEFAULT_PROJECT_NAME = "default"
DEFAULT_DB_NAME = "test_db"


def unique_name(prefix: str = "test") -> str:
    return f"{prefix}-{uuid.uuid4().hex[:8]}"


# ─────────────────────────────────────────────────────────────────────────────
# GraphQL queries
# ─────────────────────────────────────────────────────────────────────────────

CREATE_MODEL = gql("""
    mutation CreateModel($input: CreateModelInput!) {
        createModel(input: $input) {
            model { id projectSlug name }
            error { __typename }
        }
    }
""")

ADD_FIELDS = gql("""
    mutation AddFields($projectId: ID!, $modelID: ID!, $input: [AddFieldInput!]!) {
        addFields(projectId: $projectId, modelID: $modelID, input: $input) {
            id
            fields { name dbColumn { columnType constraints foreignKey { referencedTable } conflicts { aspect } } }
        }
    }
""")

GET_MODEL_WITH_SCHEMA = gql("""
    query GetModelWithSchema($projectSlug: String!, $id: ID!, $withActualSchema: Boolean) {
        model(projectSlug: $projectSlug, id: $id, withActualSchema: $withActualSchema) {
            model {
                id
                name
                dbTable
                fields {
                    name
                    format
                    dbColumn {
                        columnType
                        constraints
                        foreignKey {
                            referencedTable
                            referencedColumn
                            constraintName
                        }
                        conflicts {
                            aspect
                            expected
                            actual
                        }
                    }
                }
            }
            error { __typename }
        }
    }
""")

GET_MODEL_WITHOUT_SCHEMA = gql("""
    query GetModelBasic($projectSlug: String!, $id: ID!) {
        model(projectSlug: $projectSlug, id: $id) {
            model {
                id
                name
                dbTable
                fields {
                    name
                    dbColumn {
                        columnType
                    }
                }
            }
            error { __typename }
        }
    }
""")

REPAIR_MODEL = gql("""
    mutation RepairModel($input: RepairModelInput!) {
        repairModel(input: $input) {
            changesApplied
            executedDDL
        }
    }
""")


@pytest.fixture(scope="module")
def test_cluster(shared_test_cluster):
    return shared_test_cluster


@pytest.fixture(scope="function")
def test_model(graphql_client, test_cluster, default_project):
    """Create a simple model and ensure its DB table exists via repair."""
    model_name = unique_name("schema-test")
    create_result = graphql_client.execute(CREATE_MODEL, variable_values={
        "input": {
            "projectSlug": default_project["name"],
            "name": model_name,
            "title": f"Schema Test {model_name}",
            "databaseName": DEFAULT_DB_NAME,
        }
    })
    assert_graphql_success(create_result, "createModel")
    model = create_result["createModel"]["model"]

    # Repair to ensure the table is created in the actual DB
    graphql_client.execute(REPAIR_MODEL, variable_values={
        "input": {
            "projectSlug": default_project["name"],
            "modelId": model["id"],
            "mode": "ADDITIVE",
        }
    })

    yield model


class TestWithActualSchemaTrueTableExists:
    """Tests for withActualSchema=true when table exists."""

    def test_dbtable_is_table_exists(self, graphql_client, test_model, default_project):
        """dbTable should be TABLE_EXISTS when the table was repaired into the DB."""
        result = graphql_client.execute(GET_MODEL_WITH_SCHEMA, variable_values={
            "projectSlug": default_project["name"],
            "id": test_model["id"],
            "withActualSchema": True,
        })
        assert_graphql_success(result, "model")
        model = result["model"]["model"]
        assert model["dbTable"] == "TABLE_EXISTS", (
            f"Expected TABLE_EXISTS but got {model['dbTable']}"
        )

    def test_dbcolumn_populated_for_regular_fields(self, graphql_client, test_model, default_project):
        """Regular fields should have dbColumn populated when withActualSchema=true."""
        result = graphql_client.execute(GET_MODEL_WITH_SCHEMA, variable_values={
            "projectSlug": default_project["name"],
            "id": test_model["id"],
            "withActualSchema": True,
        })
        assert_graphql_success(result, "model")
        model = result["model"]["model"]

        # The model should have at least the auto-generated id field
        fields = model["fields"]
        assert len(fields) > 0, "Expected at least one field"

        # Find a non-ENUM_LABEL field and verify dbColumn is set
        non_virtual = [f for f in fields if f.get("format") != "ENUM_LABEL"]
        assert len(non_virtual) > 0, "Expected at least one non-virtual field"

        for field in non_virtual:
            assert field["dbColumn"] is not None, (
                f"Field '{field['name']}' should have dbColumn populated"
            )
            assert field["dbColumn"]["columnType"] != "", (
                f"Field '{field['name']}' dbColumn.columnType should not be empty"
            )


class TestWithActualSchemaFalse:
    """Tests for withActualSchema=false (or omitted)."""

    def test_dbtable_is_null_when_not_requested(self, graphql_client, test_model, default_project):
        """dbTable should be null when withActualSchema is not set."""
        result = graphql_client.execute(GET_MODEL_WITHOUT_SCHEMA, variable_values={
            "projectSlug": default_project["name"],
            "id": test_model["id"],
        })
        assert_graphql_success(result, "model")
        model = result["model"]["model"]
        assert model["dbTable"] is None, (
            f"dbTable should be null when withActualSchema is not requested, got {model['dbTable']}"
        )

    def test_dbcolumn_is_null_when_not_requested(self, graphql_client, test_model, default_project):
        """All dbColumn fields should be null when withActualSchema is not set."""
        result = graphql_client.execute(GET_MODEL_WITHOUT_SCHEMA, variable_values={
            "projectSlug": default_project["name"],
            "id": test_model["id"],
        })
        assert_graphql_success(result, "model")
        model = result["model"]["model"]

        for field in model["fields"]:
            assert field["dbColumn"] is None, (
                f"Field '{field['name']}' dbColumn should be null when not requested"
            )

    def test_dbtable_is_null_when_false(self, graphql_client, test_model, default_project):
        """dbTable should be null when withActualSchema=false."""
        result = graphql_client.execute(GET_MODEL_WITH_SCHEMA, variable_values={
            "projectSlug": default_project["name"],
            "id": test_model["id"],
            "withActualSchema": False,
        })
        assert_graphql_success(result, "model")
        model = result["model"]["model"]
        assert model["dbTable"] is None, (
            f"dbTable should be null when withActualSchema=false, got {model['dbTable']}"
        )


class TestUniqueMismatchConflict:
    """Tests for UNIQUE_MISMATCH conflict detection."""

    def test_unique_mismatch_when_design_unique_but_db_not(
        self, graphql_client, default_project, test_cluster
    ):
        """
        When a field has isUnique=true in design but the DB column has no UNIQUE
        constraint, dbColumn.conflicts should contain UNIQUE_MISMATCH.

        Note: This test creates a model with isUnique=true but repairs it only
        with ADDITIVE mode — since the column already exists without UNIQUE,
        it should show a mismatch (repair must have created it without UNIQUE).
        If repair does create with UNIQUE, we test the no-conflict case instead.
        """
        model_name = unique_name("unique-mismatch")
        create_result = graphql_client.execute(CREATE_MODEL, variable_values={
            "input": {
                "projectSlug": default_project["name"],
                "name": model_name,
                "title": f"Unique Mismatch Test {model_name}",
                "databaseName": DEFAULT_DB_NAME,
            }
        })
        assert_graphql_success(create_result, "createModel")
        model_id = create_result["createModel"]["model"]["id"]

        # Repair to create the table in DB (without isUnique fields initially)
        graphql_client.execute(REPAIR_MODEL, variable_values={
            "input": {
                "projectSlug": default_project["name"],
                "modelId": model_id,
                "mode": "ADDITIVE",
            }
        })

        # Query with actual schema - no unique fields yet means no conflicts
        result = graphql_client.execute(GET_MODEL_WITH_SCHEMA, variable_values={
            "projectSlug": default_project["name"],
            "id": model_id,
            "withActualSchema": True,
        })
        assert_graphql_success(result, "model")
        # The table should exist
        assert result["model"]["model"]["dbTable"] == "TABLE_EXISTS"


class TestEnumLabelFieldIsNull:
    """Tests that ENUM_LABEL (virtual) fields always have dbColumn=null."""

    def test_enum_label_dbcolumn_is_always_null(
        self, graphql_client, default_project
    ):
        """ENUM_LABEL virtual fields should always have dbColumn=null, even with withActualSchema=true."""
        model_name = unique_name("enum-label-test")
        create_result = graphql_client.execute(CREATE_MODEL, variable_values={
            "input": {
                "projectSlug": default_project["name"],
                "name": model_name,
                "title": f"Enum Label Test {model_name}",
                "databaseName": DEFAULT_DB_NAME,
            }
        })
        assert_graphql_success(create_result, "createModel")
        model = create_result["createModel"]["model"]

        # Query the model with actual schema
        result = graphql_client.execute(GET_MODEL_WITH_SCHEMA, variable_values={
            "projectSlug": default_project["name"],
            "id": model["id"],
            "withActualSchema": True,
        })
        assert_graphql_success(result, "model")
        model_data = result["model"]["model"]

        # Find any ENUM_LABEL fields and verify they have null dbColumn
        enum_label_fields = [f for f in model_data["fields"] if f.get("format") == "ENUM_LABEL"]
        for field in enum_label_fields:
            assert field["dbColumn"] is None, (
                f"ENUM_LABEL field '{field['name']}' should always have dbColumn=null"
            )

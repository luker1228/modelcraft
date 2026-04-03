"""
Enum GraphQL Tests

Tests for enum CRUD operations via GraphQL API with typed error handling.
"""

import pytest
from gql import gql

from design.common.test_data import (
    generate_test_id,
    build_enum_input,
    build_enum_option,
)
from design.common.assertions import assert_graphql_success


# GraphQL Queries and Mutations

CREATE_ENUM = gql("""
    mutation CreateEnum($input: CreateEnumInput!) {
        createEnum(input: $input) {
            enum {
                id
                projectSlug
                name
                displayName
                description
                options {
                    code
                    label
                    order
                    description
                }
                isMultiSelect
                createdAt
                updatedAt
            }
            error {
                __typename
                ... on EnumAlreadyExists {
                    message
                    suggestion
                }
                ... on InvalidEnumInput {
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

GET_ENUM = gql("""
    query GetEnum($projectSlug: String!, $name: String!) {
        enum(projectSlug: $projectSlug, name: $name) {
            enum {
                id
                projectSlug
                name
                displayName
                description
                options {
                    code
                    label
                    order
                }
                isMultiSelect
            }
            error {
                __typename
                ... on EnumNotFound {
                    message
                }
                ... on ProjectNotFound {
                    message
                }
            }
        }
    }
""")

LIST_ENUMS = gql("""
    query ListEnums($projectSlug: String!) {
        enums(projectSlug: $projectSlug) {
            id
            projectSlug
            name
            displayName
            options {
                code
                label
                order
            }
        }
    }
""")

UPDATE_ENUM = gql("""
    mutation UpdateEnum($projectSlug: String!, $name: String!, $input: UpdateEnumInput!) {
        updateEnum(projectSlug: $projectSlug, name: $name, input: $input) {
            enum {
                id
                projectSlug
                name
                displayName
                description
                options {
                    code
                    label
                    order
                    description
                }
            }
            error {
                __typename
                ... on EnumNotFound {
                    message
                }
                ... on InvalidEnumInput {
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

DELETE_ENUM = gql("""
    mutation DeleteEnum($projectSlug: String!, $name: String!) {
        deleteEnum(projectSlug: $projectSlug, name: $name) {
            success
            error {
                __typename
                ... on EnumNotFound {
                    message
                }
                ... on CannotDeleteReferencedEnum {
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

GET_ENUM_REFERENCES = gql("""
    query GetEnumReferences($projectSlug: String!, $name: String!) {
        enumReferences(projectSlug: $projectSlug, name: $name)
    }
""")


# Test Class

class TestEnumTypedErrors:
    """Test enum GraphQL operations with typed error handling."""

    def test_create_enum_success(self, graphql_client, default_project, created_enums):
        """Test successful enum creation returns enum data with no error."""
        enum_input = build_enum_input(
            project_name="default",

            name="OrderStatus",
            displayName="Order Status",
            description="Status for orders",
        )

        result = graphql_client.execute(CREATE_ENUM, variable_values={"input": enum_input})

        # Should return enum data with no error
        assert result["createEnum"]["enum"] is not None
        assert result["createEnum"]["error"] is None

        enum_data = result["createEnum"]["enum"]
        created_enums.append((enum_data["projectSlug"], enum_data["name"]))  # Track for cleanup

        assert enum_data["projectSlug"] == "default"
        assert enum_data["name"] == "OrderStatus"
        assert enum_data["displayName"] == "Order Status"
        assert len(enum_data["options"]) == 3
        assert enum_data["isMultiSelect"] is False

    def test_create_enum_already_exists(self, graphql_client, default_project, created_enums):
        """Test creating duplicate enum returns EnumAlreadyExists error."""
        enum_input = build_enum_input(
            project_name="default",
            name="DuplicateStatus",
        )

        # Create enum first time - should succeed
        result1 = graphql_client.execute(CREATE_ENUM, variable_values={"input": enum_input})

        # If enum already exists from previous run, delete it first
        if result1["createEnum"]["error"] is not None and result1["createEnum"]["error"]["__typename"] == "EnumAlreadyExists":
            DELETE_ENUM = gql("""
                mutation DeleteEnum($projectSlug: String!, $name: String!) {
                    deleteEnum(projectSlug: $projectSlug, name: $name) {
                        success
                    }
                }
            """)
            graphql_client.execute(DELETE_ENUM, variable_values={"projectSlug": "default", "name": "DuplicateStatus"})
            # Try again
            result1 = graphql_client.execute(CREATE_ENUM, variable_values={"input": enum_input})

        assert result1["createEnum"]["enum"] is not None, f"First creation failed: {result1}"
        assert result1["createEnum"]["error"] is None

        # Track for cleanup
        enum_data = result1["createEnum"]["enum"]
        created_enums.append((enum_data["projectSlug"], enum_data["name"]))

        # Create same enum again - should return typed error
        result2 = graphql_client.execute(CREATE_ENUM, variable_values={"input": enum_input})

        # Should return error with null enum
        assert result2["createEnum"]["enum"] is None
        assert result2["createEnum"]["error"] is not None

        error = result2["createEnum"]["error"]
        assert error["__typename"] == "EnumAlreadyExists"
        assert "DuplicateStatus" in error["message"]
        assert error["suggestion"] is not None
        assert "different enum name" in error["suggestion"]

    def test_create_enum_invalid_project(self, graphql_client):
        """Test creating enum with invalid project returns ProjectNotFound error."""
        enum_input = build_enum_input(
            project_name="nonexistent-project",
            name="TestEnum",
        )

        result = graphql_client.execute(CREATE_ENUM, variable_values={"input": enum_input})

        # Should return ProjectNotFound error
        assert result["createEnum"]["enum"] is None
        assert result["createEnum"]["error"] is not None

        error = result["createEnum"]["error"]
        assert error["__typename"] == "ProjectNotFound"
        assert "nonexistent-project" in error["message"] or "nonexistent-org" in error["message"]

    def test_create_enum_invalid_options(self, graphql_client, default_project, created_enums):
        """Test creating enum with duplicate option codes returns InvalidEnumInput error."""
        enum_input = build_enum_input(
            project_name="default",
            
            name="InvalidEnum",
            options=[
                build_enum_option("active", "Active", 1),
                build_enum_option("active", "Also Active", 2),  # Duplicate code
            ],
        )

        result = graphql_client.execute(CREATE_ENUM, variable_values={"input": enum_input})

        # Should return InvalidEnumInput error
        assert result["createEnum"]["enum"] is None
        assert result["createEnum"]["error"] is not None

        error = result["createEnum"]["error"]
        assert error["__typename"] == "InvalidEnumInput"
        assert "duplicate" in error["message"].lower() or "invalid" in error["message"].lower()
        assert error["suggestion"] is not None

    def test_get_enum_success(self, graphql_client, default_project, created_enums):
        """Test getting existing enum returns enum data with no error."""
        # Create enum first
        enum_input = build_enum_input(
            project_name="default",
            
            name="GetTestStatus",
        )
        create_result = graphql_client.execute(CREATE_ENUM, variable_values={"input": enum_input})
        assert create_result["createEnum"]["enum"] is not None
        created_enums.append((create_result["createEnum"]["enum"]["projectSlug"], create_result["createEnum"]["enum"]["name"]))
        # Get the enum
        result = graphql_client.execute(
            GET_ENUM,
            variable_values={
                "projectSlug": "default",
                "name": "GetTestStatus",
            }
        )

        # Should return enum data with no error
        assert result["enum"]["enum"] is not None
        assert result["enum"]["error"] is None

        enum_data = result["enum"]["enum"]
        assert enum_data["name"] == "GetTestStatus"
        assert enum_data["projectSlug"] == "default"

    def test_get_enum_not_found(self, graphql_client, default_project, created_enums):
        """Test getting non-existent enum returns EnumNotFound error."""
        result = graphql_client.execute(
            GET_ENUM,
            variable_values={
                "projectSlug": "default",
                "name": "NonExistentEnum",
            }
        )

        # Should return EnumNotFound error
        assert result["enum"]["enum"] is None
        assert result["enum"]["error"] is not None

        error = result["enum"]["error"]
        assert error["__typename"] == "EnumNotFound"
        assert "NonExistentEnum" in error["message"]

    def test_get_enum_invalid_project(self, graphql_client):
        """Test getting enum with invalid project returns ProjectNotFound error."""
        result = graphql_client.execute(
            GET_ENUM,
            variable_values={
                "projectSlug": "invalid-project",
                "name": "SomeEnum",
            }
        )

        # Should return ProjectNotFound error
        assert result["enum"]["enum"] is None
        assert result["enum"]["error"] is not None

        error = result["enum"]["error"]
        assert error["__typename"] == "ProjectNotFound"
        assert "invalid-project" in error["message"] or "invalid-org" in error["message"]

    def test_update_enum_success(self, graphql_client, default_project, created_enums):
        """Test updating existing enum returns updated enum data with no error."""
        # Create enum first
        enum_input = build_enum_input(
            project_name="default",
            
            name="UpdateTestStatus",
            displayName="Original Title",
        )
        create_result = graphql_client.execute(CREATE_ENUM, variable_values={"input": enum_input})
        assert create_result["createEnum"]["enum"] is not None
        created_enums.append((create_result["createEnum"]["enum"]["projectSlug"], create_result["createEnum"]["enum"]["name"]))
        # Update the enum
        update_input = {
            "displayName": "Updated Title",
            "description": "Updated description",
        }
        result = graphql_client.execute(
            UPDATE_ENUM,
            variable_values={
                "projectSlug": "default",
                "name": "UpdateTestStatus",
                "input": update_input,
            }
        )

        # Should return updated enum with no error
        assert result["updateEnum"]["enum"] is not None
        assert result["updateEnum"]["error"] is None

        enum_data = result["updateEnum"]["enum"]
        assert enum_data["displayName"] == "Updated Title"
        assert enum_data["description"] == "Updated description"

    def test_update_enum_not_found(self, graphql_client, default_project, created_enums):
        """Test updating non-existent enum returns EnumNotFound error."""
        update_input = {
            "displayName": "New Title",
        }
        result = graphql_client.execute(
            UPDATE_ENUM,
            variable_values={
                "projectSlug": "default",
                "name": "NonExistentEnum",
                "input": update_input,
            }
        )

        # Should return EnumNotFound error
        assert result["updateEnum"]["enum"] is None
        assert result["updateEnum"]["error"] is not None

        error = result["updateEnum"]["error"]
        assert error["__typename"] == "EnumNotFound"
        assert "NonExistentEnum" in error["message"]

    def test_update_enum_invalid_options(self, graphql_client, default_project, created_enums):
        """Test updating enum with invalid options returns InvalidEnumInput error."""
        # Create enum first
        enum_input = build_enum_input(
            project_name="default",
            
            name="InvalidUpdateEnum",
        )
        create_result = graphql_client.execute(CREATE_ENUM, variable_values={"input": enum_input})
        assert create_result["createEnum"]["enum"] is not None
        created_enums.append((create_result["createEnum"]["enum"]["projectSlug"], create_result["createEnum"]["enum"]["name"]))
        # Update with duplicate codes
        update_input = {
            "options": [
                build_enum_option("same", "Option 1", 1),
                build_enum_option("same", "Option 2", 2),  # Duplicate code
            ],
        }
        result = graphql_client.execute(
            UPDATE_ENUM,
            variable_values={
                "projectSlug": "default",
                "name": "InvalidUpdateEnum",
                "input": update_input,
            }
        )

        # Should return InvalidEnumInput error
        assert result["updateEnum"]["enum"] is None
        assert result["updateEnum"]["error"] is not None

        error = result["updateEnum"]["error"]
        assert error["__typename"] == "InvalidEnumInput"
        assert error["suggestion"] is not None

    def test_delete_enum_success(self, graphql_client, default_project, created_enums):
        """Test deleting existing enum returns success with no error."""
        # Create enum first
        enum_input = build_enum_input(
            project_name="default",
            
            name="DeleteTestStatus",
        )
        create_result = graphql_client.execute(CREATE_ENUM, variable_values={"input": enum_input})
        assert create_result["createEnum"]["enum"] is not None
        created_enums.append((create_result["createEnum"]["enum"]["projectSlug"], create_result["createEnum"]["enum"]["name"]))
        # Delete the enum
        result = graphql_client.execute(
            DELETE_ENUM,
            variable_values={
                "projectSlug": "default",
                "name": "DeleteTestStatus",
            }
        )

        # Should return success with no error
        assert result["deleteEnum"]["success"] is True
        assert result["deleteEnum"]["error"] is None

        # Verify enum is deleted by trying to get it
        get_result = graphql_client.execute(
            GET_ENUM,
            variable_values={
                "projectSlug": "default",
                "name": "DeleteTestStatus",
            }
        )
        assert get_result["enum"]["error"] is not None
        assert get_result["enum"]["error"]["__typename"] == "EnumNotFound"

    def test_delete_enum_not_found(self, graphql_client, default_project, created_enums):
        """Test deleting non-existent enum returns EnumNotFound error."""
        result = graphql_client.execute(
            DELETE_ENUM,
            variable_values={
                "projectSlug": "default",
                "name": "NonExistentEnum",
            }
        )

        # Should return EnumNotFound error
        assert result["deleteEnum"]["success"] is False
        assert result["deleteEnum"]["error"] is not None

        error = result["deleteEnum"]["error"]
        assert error["__typename"] == "EnumNotFound"
        assert "NonExistentEnum" in error["message"]

    def test_list_enums(self, graphql_client, default_project, created_enums):
        """Test listing enums returns array (not wrapped in payload)."""
        # Create multiple enums
        for i in range(3):
            enum_input = build_enum_input(
                project_name="default",
                
                name=f"ListTestEnum{i}",
            )
            result = graphql_client.execute(CREATE_ENUM, variable_values={"input": enum_input})
            assert result["createEnum"]["enum"] is not None
            created_enums.append((result["createEnum"]["enum"]["projectSlug"], result["createEnum"]["enum"]["name"]))
        # List enums
        result = graphql_client.execute(
            LIST_ENUMS,
            variable_values={
                "projectSlug": "default",
            }
        )

        # Should return array directly (backward compatibility)
        enums = result["enums"]
        assert isinstance(enums, list)
        assert len(enums) >= 3

        # Verify some enum names are in the list
        enum_names = [e["name"] for e in enums]
        assert "ListTestEnum0" in enum_names
        assert "ListTestEnum1" in enum_names
        assert "ListTestEnum2" in enum_names

    def test_enum_references(self, graphql_client, default_project, created_enums):
        """Test getting enum references returns array (not wrapped in payload)."""
        # Create an enum
        enum_input = build_enum_input(
            project_name="default",
            
            name="ReferencesTestEnum",
        )
        create_result = graphql_client.execute(CREATE_ENUM, variable_values={"input": enum_input})
        assert create_result["createEnum"]["enum"] is not None
        created_enums.append((create_result["createEnum"]["enum"]["projectSlug"], create_result["createEnum"]["enum"]["name"]))
        # Get references (should be empty for now)
        result = graphql_client.execute(
            GET_ENUM_REFERENCES,
            variable_values={
                "projectSlug": "default",
                "name": "ReferencesTestEnum",
            }
        )

        # Should return array directly (backward compatibility)
        references = result["enumReferences"]
        assert isinstance(references, list)
        assert len(references) == 0  # No fields reference this enum yet

    def test_create_enum_with_multi_select(self, graphql_client, default_project, created_enums):
        """Test creating multi-select enum."""
        enum_input = build_enum_input(
            project_name="default",
            
            name="MultiSelectTags",
            displayName="Tags",
            is_multi_select=True,
            options=[
                build_enum_option("urgent", "Urgent", 1),
                build_enum_option("important", "Important", 2),
                build_enum_option("review", "Needs Review", 3),
            ],
        )

        result = graphql_client.execute(CREATE_ENUM, variable_values={"input": enum_input})

        assert result["createEnum"]["enum"] is not None
        assert result["createEnum"]["error"] is None

        enum_data = result["createEnum"]["enum"]
        created_enums.append((enum_data["projectSlug"], enum_data["name"]))  # Track for cleanup
        assert enum_data["isMultiSelect"] is True
        assert enum_data["name"] == "MultiSelectTags"

    def test_update_enum_options(self, graphql_client, default_project, created_enums):
        """Test updating enum options."""
        # Create enum
        enum_input = build_enum_input(
            project_name="default",
            
            name="OptionsUpdateEnum",
            options=[
                build_enum_option("draft", "Draft", 1),
                build_enum_option("published", "Published", 2),
            ],
        )
        create_result = graphql_client.execute(CREATE_ENUM, variable_values={"input": enum_input})
        assert create_result["createEnum"]["enum"] is not None
        created_enums.append((create_result["createEnum"]["enum"]["projectSlug"], create_result["createEnum"]["enum"]["name"]))
        # Update options (add new one)
        update_input = {
            "options": [
                build_enum_option("draft", "Draft", 1),
                build_enum_option("published", "Published", 2),
                build_enum_option("archived", "Archived", 3),  # New option
            ],
        }
        result = graphql_client.execute(
            UPDATE_ENUM,
            variable_values={
                "projectSlug": "default",
                "name": "OptionsUpdateEnum",
                "input": update_input,
            }
        )

        assert result["updateEnum"]["enum"] is not None
        assert result["updateEnum"]["error"] is None

        enum_data = result["updateEnum"]["enum"]
        assert len(enum_data["options"]) == 3
        option_codes = [opt["code"] for opt in enum_data["options"]]
        assert "archived" in option_codes

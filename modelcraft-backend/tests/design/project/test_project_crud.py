"""
Project Management CRUD Tests

Tests for project creation, read, update, delete operations via GraphQL API.
"""

import pytest
from gql import gql

from design.common.test_data import build_project_input, generate_test_id
from design.common.assertions import (
    assert_graphql_success,
    assert_project_fields,
    assert_graphql_error,
    assert_query_returns_none
)


# GraphQL Queries and Mutations
CREATE_PROJECT = gql("""
    mutation CreateProject($input: CreateProjectInput!) {
        createProject(input: $input) {
            project {
                orgName
                name
                title
                description
                status
                createdAt
                updatedAt
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
            }
        }
    }
""")

GET_PROJECT = gql("""
    query GetProject($name: String!) {
        project(name: $name) {
            project {
                orgName
                name
                title
                description
                status
                createdAt
                updatedAt
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
        }
    }
""")

LIST_PROJECTS = gql("""
    query ListProjects {
        projects {
            orgName
            name
            title
            status
        }
    }
""")


class TestProjectCRUD:
    """Test suite for project CRUD operations"""

    def test_create_project_success(self, graphql_client, created_projects):
        """Test creating a new project with valid data"""
        # Arrange
        project_name = generate_test_id("test_project_001")
        input_data = build_project_input(
            name=project_name,
            title="Test Project",
            description="A test project for integration testing"
        )

        # Act
        print(f"[INFO] Creating project with input: {input_data}")
        result = graphql_client.execute(CREATE_PROJECT, variable_values={"input": input_data})
        print(f"[DEBUG] CREATE_PROJECT result: {result}")

        # Assert
        assert_graphql_success(result, "createProject")
        project = result["createProject"]["project"]
        assert_project_fields(project, {
            "orgName": "modelcraft",
            "name": project_name,
            "title": "Test Project",
            "status": "ACTIVE"
        })

        # Track for cleanup (using tuple of orgName and name)
        created_projects.append({"name": project["name"], "orgName": project["orgName"]})

    def test_get_project_success(self, graphql_client, created_projects):
        """Test retrieving a project by orgName and name"""
        # Arrange: Create a project first
        project_name = generate_test_id("test_project_002")
        input_data = build_project_input(name=project_name, org_name="built-in")
        create_result = graphql_client.execute(CREATE_PROJECT, variable_values={"input": input_data})
        project = create_result["createProject"]["project"]
        created_projects.append({"name": project["name"], "orgName": project["orgName"]})

        # Act
        result = graphql_client.execute(GET_PROJECT, variable_values={
            "name": project["name"]
        })

        # Assert
        assert_graphql_success(result, "project")
        payload = result["project"]
        retrieved_project = payload["project"]
        assert retrieved_project is not None, "Project should exist"
        assert retrieved_project["orgName"] == project["orgName"]
        assert retrieved_project["name"] == project["name"]
        # Verify no errors (error field should be None for successful query)
        assert payload.get("error") is None, "Should not have any error for existing project"

    def test_list_projects_success(self, graphql_client, created_projects):
        """Test listing all projects"""
        # Arrange: Create 2 test projects
        project_name_1 = generate_test_id("test_project_list-001")
        project_name_2 = generate_test_id("test_project_list-002")

        input_data_1 = build_project_input(
            name=project_name_1,
            org_name="built-in",
            title="List Test Project 1",
            description="First project for list testing"
        )
        input_data_2 = build_project_input(
            name=project_name_2,
            org_name="built-in",
            title="List Test Project 2",
            description="Second project for list testing"
        )

        # Create both projects
        create_result_1 = graphql_client.execute(CREATE_PROJECT, variable_values={"input": input_data_1})
        create_result_2 = graphql_client.execute(CREATE_PROJECT, variable_values={"input": input_data_2})

        assert_graphql_success(create_result_1, "createProject")
        assert_graphql_success(create_result_2, "createProject")

        project_1 = create_result_1["createProject"]["project"]
        project_2 = create_result_2["createProject"]["project"]
        created_projects.append({"name": project_1["name"], "orgName": project_1["orgName"]})
        created_projects.append({"name": project_2["name"], "orgName": project_2["orgName"]})

        # Act: List all projects for the organization
        result = graphql_client.execute(LIST_PROJECTS)

        # Assert
        assert_graphql_success(result, "projects")
        projects = result["projects"]
        assert isinstance(projects, list)

        # Both created projects should exist in the list
        project_keys = [(p["orgName"], p["name"]) for p in projects]
        assert (project_1["orgName"], project_1["name"]) in project_keys, \
            f"Project {project_name_1} should be in the list"
        assert (project_2["orgName"], project_2["name"]) in project_keys, \
            f"Project {project_name_2} should be in the list"

    def test_create_duplicate_project_error(self, graphql_client, created_projects):
        """Test creating a duplicate project returns ProjectAlreadyExists error"""
        # Arrange: Create a project first
        project_name = generate_test_id("test_project_duplicate")
        input_data = build_project_input(
            name=project_name,
            org_name="built-in",
            title="Duplicate Test Project",
            description="Testing duplicate project creation"
        )

        # Create the project for the first time
        first_result = graphql_client.execute(CREATE_PROJECT, variable_values={"input": input_data})
        print(f"[DEBUG] First CREATE_PROJECT result: {first_result}")
        assert_graphql_success(first_result, "createProject")
        project = first_result["createProject"]["project"]
        created_projects.append({"name": project["name"], "orgName": project["orgName"]})

        # Act: Try to create the same project again
        print(f"[INFO] Attempting to create duplicate project: {project_name}")
        second_result = graphql_client.execute(CREATE_PROJECT, variable_values={"input": input_data})
        print(f"[DEBUG] Second CREATE_PROJECT result: {second_result}")

        # Assert: Should return ProjectAlreadyExists error
        assert_graphql_success(second_result, "createProject")
        payload = second_result["createProject"]

        # Project should be None when there's an error
        assert payload["project"] is None, "Project should be None when duplicate error occurs"

        # Should have ProjectAlreadyExists error
        assert payload["error"] is not None, "Should have an error for duplicate project"
        error = payload["error"]
        assert error["__typename"] == "ProjectAlreadyExists", \
            f"Expected ProjectAlreadyExists error, got {error.get('__typename')}"

        # Verify error message contains useful information
        assert "message" in error, "Error should contain a message"
        assert project_name in error["message"] or "already exists" in error["message"].lower(), \
            f"Error message should mention the project or 'already exists': {error['message']}"

    def test_get_nonexistent_project_returns_error(self, graphql_client):
        """Test getting a non-existent project returns ProjectNotFound error"""
        # Arrange
        nonexistent_name = "nonexistent-project-12345"

        # Act
        print(f"[INFO] Attempting to get non-existent project: {nonexistent_name}")
        result = graphql_client.execute(GET_PROJECT, variable_values={
            "name": nonexistent_name
        })
        print(f"[DEBUG] GET_PROJECT result: {result}")

        # Assert: Should return ProjectNotFound error
        assert_graphql_success(result, "project")
        payload = result["project"]

        # Project should be None
        assert payload["project"] is None

        # Should have ProjectNotFound error
        assert payload["error"] != None
        error = payload["error"]
        assert error["__typename"] == "ProjectNotFound"

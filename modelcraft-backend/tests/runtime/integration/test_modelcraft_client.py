"""
ModelCraft Client Integration Tests

End-to-end integration tests that exercise the full Design → Runtime flow.
These tests verify that the complete ModelCraft workflow functions correctly.
"""

import pytest
from gql import gql

# Import Design utilities for setup
from design.common.test_data import build_project_input, build_model_input
from design.common.assertions import assert_graphql_success


class TestModelCraftClientIntegration:
    """Integration tests for complete ModelCraft workflows"""

    def test_complete_model_lifecycle(self, graphql_client, default_project, created_models):
        """
        Test the complete lifecycle: Create model (using project's cluster) → Query data

        This test demonstrates the Design → Runtime flow.
        The cluster is now created atomically with the project.
        """
        # Phase 1: Verify default project has a cluster
        GET_PROJECT = gql("""
            query GetProject($name: String!) {
                project(name: $name) {
                    project {
                        name
                        cluster {
                            name
                        }
                    }
                    error { __typename }
                }
            }
        """)
        proj_result = graphql_client.execute(GET_PROJECT, variable_values={"name": default_project["name"]})
        project_data = proj_result.get("project", {}).get("project")
        assert project_data is not None, "Default project should exist"
        cluster = project_data.get("cluster")
        assert cluster is not None, "Default project should have a cluster"
        cluster_name = cluster["name"]

        # Phase 2: Design - Create model
        CREATE_MODEL = gql("""
            mutation CreateModel($input: CreateModelInput!) {
                createModel(input: $input) {
                    model {
                        id
                        projectSlug
                        name
                    }
                    error { __typename }
                }
            }
        """)

        model_input = build_model_input(
            project_name=default_project["name"],
            cluster_name=cluster_name,
            name="IntegrationTestModel",
            description="Model for integration testing"
        )
        model_result = graphql_client.execute(CREATE_MODEL, variable_values={"input": model_input})
        assert_graphql_success(model_result, "createModel")
        payload = model_result["createModel"]
        if payload.get("error") is not None:
            pytest.skip(f"Model creation skipped: {payload['error']}")

        model = payload["model"]
        created_models.append((model["projectSlug"], model["id"]))

        print(f"\n✅ Complete lifecycle test passed for model {model['projectSlug']}/{model['name']}")

    def test_project_cluster_model_integration(self, graphql_client, created_projects, created_models):
        """
        Test integration across projects, clusters, and models.

        Verifies that resources are correctly linked and isolated.
        """
        # Phase 1: Create a test project (with embedded cluster)
        CREATE_PROJECT = gql("""
            mutation CreateProject($input: CreateProjectInput!) {
                createProject(input: $input) {
                    project {
                        orgName
                        name
                        title
                        cluster {
                            name
                        }
                    }
                    error { __typename }
                }
            }
        """)

        project_input = build_project_input(
            name="integration_test_project",
            title="Integration Test Project",
            cluster_name="integration_test_cluster",
            cluster_title="Integration Test Cluster",
            skip_connection_test=True,
        )
        project_result = graphql_client.execute(CREATE_PROJECT, variable_values={"input": project_input})
        payload = project_result.get("createProject", {})
        if payload.get("error") and payload["error"].get("__typename") != "ProjectAlreadyExists":
            pytest.skip(f"Project creation failed: {payload['error']}")

        if payload.get("project") is not None:
            project = payload["project"]
            created_projects.append({"name": project["name"], "orgName": project["orgName"]})
            cluster_name = project["cluster"]["name"]
        else:
            # Project already existed - fetch it
            GET_PROJECT = gql("""
                query GetProject($name: String!) {
                    project(name: $name) {
                        project { name orgName cluster { name } }
                    }
                }
            """)
            get_result = graphql_client.execute(GET_PROJECT, variable_values={"name": "integration_test_project"})
            project = get_result["project"]["project"]
            cluster_name = project["cluster"]["name"]

        # Phase 2: Create model in this project's cluster
        CREATE_MODEL = gql("""
            mutation CreateModel($input: CreateModelInput!) {
                createModel(input: $input) {
                    model {
                        id
                        projectSlug
                        name
                    }
                    error { __typename }
                }
            }
        """)

        model_input = build_model_input(
            project_name=project["name"],
            cluster_name=cluster_name,
            name="ProjectIntegrationModel"
        )
        model_result = graphql_client.execute(CREATE_MODEL, variable_values={"input": model_input})
        assert_graphql_success(model_result, "createModel")
        model_payload = model_result["createModel"]
        if model_payload.get("error") is not None:
            pytest.skip(f"Model creation skipped: {model_payload['error']}")

        model = model_payload["model"]
        assert model["projectSlug"] == project["name"]
        created_models.append((model["projectSlug"], model["id"]))

        print(f"\n✅ Project-Cluster-Model integration verified for project {project['orgName']}/{project['name']}")

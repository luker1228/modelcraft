# Specification: Model Integration Testing

## ADDED Requirements

### Requirement: GraphQL test queries must use unique parameter names

GraphQL queries and mutations in integration tests MUST NOT have duplicate parameter names. Each parameter name SHALL appear exactly once in the parameter list and query arguments, as duplicate parameters violate GraphQL specification and cause query execution failures.

#### Scenario: Model query with single projectName parameter

**Given** a test needs to query a model by project and ID
**When** defining the GraphQL GET_MODEL query
**Then** the query should have exactly one `projectName` parameter
**And** the query should have exactly one `id` parameter
**And** field selections should not duplicate parameter names

Example (correct):
```graphql
query GetModel($projectName: String!, $id: ID!) {
    model(projectName: $projectName, id: $id) {
        model { id, name, projectName }
        error { __typename }
    }
}
```

Example (incorrect - duplicate projectName):
```graphql
query GetModel($projectName: String!, $projectName: String!, $id: ID!) {
    model(projectName: $projectName, projectName: $projectName, id: $id) {
        model { id, name, projectName, projectName }
        error { __typename }
    }
}
```

#### Scenario: Model creation with cluster reference validation

**Given** a test needs to create a model associated with a cluster
**When** the cluster is created in a fixture
**Then** the fixture must validate cluster creation succeeded
**And** must return the cluster name (not ID) for model reference
**And** model creation must reference the cluster by name

Example:
```python
@pytest.fixture
def test_cluster(graphql_client, created_clusters):
    result = graphql_client.execute(CREATE_CLUSTER, variables={"input": input_data})
    payload = result["createDatabaseCluster"]

    # Validate cluster creation succeeded
    if payload.get("error") is not None:
        raise Exception(f"Cluster creation failed: {payload['error']}")

    cluster = payload["cluster"]
    created_clusters.append((cluster["projectName"], cluster["name"]))  # Track by name
    return cluster["name"]  # Return name for model reference
```

### Requirement: Test fixtures must track resources with correct identifiers for cleanup

Fixture cleanup MUST use the same identifier type (ID or name) that the deletion mutation expects. Tracking tuples SHALL match the mutation parameter requirements to ensure successful resource cleanup.

#### Scenario: Cluster fixture tracking and cleanup

**Given** a cluster is created in a test fixture
**When** the cluster is tracked for cleanup
**Then** it must be stored as `(projectName, name)` tuple
**And** the cleanup mutation must use `projectName` and `name` parameters
**And** cleanup must handle deletion failures gracefully

Example:
```python
@pytest.fixture
def created_clusters(graphql_client):
    clusters = []
    yield clusters

    DELETE_CLUSTER = gql("""
        mutation DeleteCluster($projectName: String!, $name: String!) {
            deleteDatabaseCluster(projectName: $projectName, name: $name) {
                success
            }
        }
    """)

    for project_name, cluster_name in clusters:
        try:
            graphql_client.execute(DELETE_CLUSTER, variable_values={
                "projectName": project_name,
                "name": cluster_name
            })
        except Exception as e:
            print(f"⚠ Failed to cleanup cluster {project_name}/{cluster_name}: {e}")
```

#### Scenario: Model fixture tracking and cleanup

**Given** a model is created in a test fixture
**When** the model is tracked for cleanup
**Then** it must be stored as `(projectName, id)` tuple
**And** the cleanup mutation must use `projectName` and `id` parameters
**And** model creation must validate success before tracking

Example:
```python
@pytest.fixture
def test_model(graphql_client, test_cluster, created_models):
    result = graphql_client.execute(CREATE_MODEL, variables={"input": input_data})
    payload = result["createModel"]

    # Validate model creation succeeded
    if payload.get("error") is not None:
        raise Exception(f"Model creation failed: {payload['error']}")

    model = payload["model"]
    created_models.append((model["projectName"], model["id"]))  # Track by id
    return model
```

### Requirement: Fixtures must validate resource creation before using resources

Test fixtures that create resources (clusters, models, enums) MUST validate creation succeeded before returning or using the resource. Fixtures SHALL check for error fields and raise exceptions with descriptive messages when resource creation fails. This prevents cascading failures and provides clear error diagnostics.

#### Scenario: Cluster creation validation in fixture

**Given** a fixture creates a cluster for testing
**When** the cluster creation mutation returns
**Then** the fixture must check for error field
**And** must raise exception if error exists
**And** must not proceed with resource tracking if creation failed

Example:
```python
@pytest.fixture
def test_cluster(graphql_client, created_clusters):
    result = graphql_client.execute(CREATE_CLUSTER, variables={"input": input_data})
    payload = result["createDatabaseCluster"]

    # Critical validation step
    if payload.get("error") is not None:
        error_type = payload['error'].get('__typename', 'Unknown')
        error_msg = payload['error'].get('message', 'No message')
        raise Exception(f"Failed to create test cluster: {error_type} - {error_msg}")

    cluster = payload["cluster"]
    # Only track if creation succeeded
    created_clusters.append((cluster["projectName"], cluster["name"]))
    return cluster["name"]
```

#### Scenario: Model creation validation prevents null reference errors

**Given** a test model fixture creates a model
**When** the create mutation returns None (due to error)
**Then** the fixture must not access model fields
**And** must not append None to tracking list
**And** must raise descriptive error for debugging

Example (incorrect - causes TypeError):
```python
result = graphql_client.execute(CREATE_MODEL, variables={"input": input_data})
model = result["createModel"]["model"]  # Could be None
created_models.append((model["projectName"], model["id"]))  # TypeError if model is None
```

Example (correct):
```python
result = graphql_client.execute(CREATE_MODEL, variables={"input": input_data})
payload = result["createModel"]

if payload.get("error") is not None or payload.get("model") is None:
    error = payload.get("error", {})
    raise Exception(f"Model creation failed: {error}")

model = payload["model"]
created_models.append((model["projectName"], model["id"]))
```

### Requirement: default_project fixture must provide expected field names

Tests expect specific fields from the `default_project` fixture. The fixture MUST either provide a `projectName` field or tests MUST consistently use the `name` field. Field naming SHALL be consistent across all test usages to prevent KeyError exceptions.

#### Scenario: Consistent project field access in tests

**Given** the default_project fixture returns project data
**When** tests access project identification
**Then** tests must use the field name provided by fixture
**Or** fixture must alias the field to match test expectations

Option 1 - Update fixture to provide projectName:
```python
@pytest.fixture
def default_project(graphql_client):
    # ... project retrieval logic ...
    project = result["project"]["project"]
    # Add alias for backward compatibility
    if "name" in project and "projectName" not in project:
        project["projectName"] = project["name"]
    return project
```

Option 2 - Update all tests to use "name":
```python
# Before (causes KeyError)
result = graphql_client.execute(GET_MODEL, variable_values={
    "projectName": default_project["projectName"],  # KeyError
    "id": "test-id"
})

# After (correct)
result = graphql_client.execute(GET_MODEL, variable_values={
    "projectName": default_project["name"],  # Works
    "id": "test-id"
})
```

## Implementation Notes

### Files to Modify

1. **tests/design/model/test_model_errors.py**
   - Fix duplicate GraphQL parameters in all queries (lines 30-180)
   - Fix test_cluster fixture validation (lines 200-224)
   - Fix test_model fixture validation and tracking (lines 226-244)
   - Fix all test method query executions (lines 250-400)

2. **tests/design/model/test_model_graphql.py**
   - Fix duplicate GraphQL parameters in queries (lines 32-120)
   - Fix default_project field access (lines 296, 316)
   - Ensure all query executions use correct parameters

3. **tests/design/conftest.py** (if needed)
   - Verify cluster cleanup uses correct parameters
   - Verify model cleanup uses correct parameters
   - Add projectName alias to default_project if choosing Option 1

### Testing Checklist

After implementing fixes:
- [ ] Run `pytest tests/design/model/test_model_errors.py -v` - all pass
- [ ] Run `pytest tests/design/model/test_model_graphql.py -v` - all pass
- [ ] Run `pytest tests/design/model/ -v` - 23/23 pass
- [ ] Verify no "Failed to cleanup" warnings in output
- [ ] Run tests 3 times consecutively - consistent results
- [ ] Check database cleanup - no orphaned resources

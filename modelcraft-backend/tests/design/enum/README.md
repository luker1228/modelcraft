# Enum GraphQL Typed Errors - Integration Tests

This directory contains integration tests for enum GraphQL operations with typed error handling.

## Test File

- `test_enum_graphql.py` - Comprehensive tests for enum CRUD operations with typed errors

## Test Coverage

The test suite validates all typed error scenarios:

### Create Enum Tests
- ✅ Successful creation returns enum data with no error
- ✅ Duplicate enum name returns `EnumAlreadyExists` error with suggestion
- ✅ Invalid project ID returns `ProjectNotFound` error
- ✅ Duplicate option codes return `InvalidEnumInput` error with suggestion
- ✅ Multi-select enum creation

### Get Enum Tests
- ✅ Getting existing enum returns enum data with no error
- ✅ Non-existent enum returns `EnumNotFound` error
- ✅ Invalid project ID returns `ProjectNotFound` error

### Update Enum Tests
- ✅ Successful update returns updated enum data with no error
- ✅ Non-existent enum returns `EnumNotFound` error
- ✅ Invalid options return `InvalidEnumInput` error
- ✅ Updating options (adding new option)

### Delete Enum Tests
- ✅ Successful deletion returns success with no error
- ✅ Non-existent enum returns `EnumNotFound` error
- ✅ Deleting referenced enum returns `CannotDeleteReferencedEnum` error (when implemented)

### List & References Tests
- ✅ List enums returns array (backward compatibility)
- ✅ Get enum references returns array (backward compatibility)

## Running Tests

### Setup

```bash
# From project root
cd tests

# Setup Python environment (if not done)
./setup-python.sh

# Activate virtual environment
source venv/bin/activate
```

### Run All Enum Tests

```bash
# Run all enum tests with verbose output
pytest design/enum/ -v

# Run with detailed output
pytest design/enum/ -vv

# Run with coverage
pytest design/enum/ --cov=design.enum --cov-report=html
```

### Run Specific Tests

```bash
# Run specific test class
pytest design/enum/test_enum_graphql.py::TestEnumTypedErrors -v

# Run specific test method
pytest design/enum/test_enum_graphql.py::TestEnumTypedErrors::test_create_enum_already_exists -v

# Run tests matching pattern
pytest design/enum/ -k "create" -v
pytest design/enum/ -k "error" -v
```

### Run with Server

```bash
# Start server (in another terminal)
cd /root/modelcraft_project/modelcraft-go
task run

# Run tests
cd tests
source venv/bin/activate
pytest design/enum/ -v
```

## Test Fixtures

The tests use these fixtures from `design/conftest.py`:

- `gql_client` - GraphQL client for Design API
- `test_project` - Creates a unique test project for each test (auto-cleanup)
- `created_enums` - Tracks created enums for cleanup (module-scoped)

## Error Types Tested

### EnumAlreadyExists
- **Trigger**: Create enum with duplicate name in same project
- **Fields**: `message`, `suggestion`
- **Example**: `"Enum already exists: Status"`

### EnumNotFound
- **Trigger**: Get/update/delete non-existent enum
- **Fields**: `message`
- **Example**: `"Enum not found: NonExistentEnum"`

### InvalidEnumInput
- **Trigger**: Create/update enum with invalid data (duplicate codes, etc.)
- **Fields**: `message`, `suggestion`
- **Example**: `"Invalid enum options: duplicate code 'active'"`

### CannotDeleteReferencedEnum
- **Trigger**: Delete enum that is referenced by model fields
- **Fields**: `message`, `suggestion`
- **Example**: `"Cannot delete enum 'Status', it is referenced by fields: Order.status"`

### ProjectNotFound
- **Trigger**: Any operation with invalid project ID
- **Fields**: `message`
- **Example**: `"Project not found: invalid-project"`

## Test Data

Test data is built using helpers from `design/common/test_data.py`:

```python
# Build enum input
enum_input = build_enum_input(
    project_id="my-project",
    name="OrderStatus",
    title="Order Status",
    options=[
        build_enum_option("pending", "Pending", 1),
        build_enum_option("completed", "Completed", 2),
    ],
    is_multi_select=False,
)
```

## GraphQL Queries Used

All GraphQL queries and mutations are defined at the top of the test file:

- `CREATE_ENUM` - Create enum with typed errors
- `GET_ENUM` - Get enum with typed errors
- `LIST_ENUMS` - List enums (array response)
- `UPDATE_ENUM` - Update enum with typed errors
- `DELETE_ENUM` - Delete enum with typed errors
- `GET_ENUM_REFERENCES` - Get enum field references (array response)

## Expected Test Results

All tests should pass when the enum GraphQL typed error implementation is correct:

```
design/enum/test_enum_graphql.py::TestEnumTypedErrors::test_create_enum_success PASSED
design/enum/test_enum_graphql.py::TestEnumTypedErrors::test_create_enum_already_exists PASSED
design/enum/test_enum_graphql.py::TestEnumTypedErrors::test_create_enum_invalid_project PASSED
design/enum/test_enum_graphql.py::TestEnumTypedErrors::test_create_enum_invalid_options PASSED
design/enum/test_enum_graphql.py::TestEnumTypedErrors::test_get_enum_success PASSED
design/enum/test_enum_graphql.py::TestEnumTypedErrors::test_get_enum_not_found PASSED
design/enum/test_enum_graphql.py::TestEnumTypedErrors::test_get_enum_invalid_project PASSED
design/enum/test_enum_graphql.py::TestEnumTypedErrors::test_update_enum_success PASSED
design/enum/test_enum_graphql.py::TestEnumTypedErrors::test_update_enum_not_found PASSED
design/enum/test_enum_graphql.py::TestEnumTypedErrors::test_update_enum_invalid_options PASSED
design/enum/test_enum_graphql.py::TestEnumTypedErrors::test_delete_enum_success PASSED
design/enum/test_enum_graphql.py::TestEnumTypedErrors::test_delete_enum_not_found PASSED
design/enum/test_enum_graphql.py::TestEnumTypedErrors::test_list_enums PASSED
design/enum/test_enum_graphql.py::TestEnumTypedErrors::test_enum_references PASSED
design/enum/test_enum_graphql.py::TestEnumTypedErrors::test_create_enum_with_multi_select PASSED
design/enum/test_enum_graphql.py::TestEnumTypedErrors::test_update_enum_options PASSED

=================== 16 passed in X.XXs ===================
```

## Troubleshooting

### Test Failures

If tests fail, check:

1. **Server is running**: `task run` in project root
2. **Database is clean**: Run cleanup scripts if needed
3. **GraphQL schema is current**: `task generate-gql`
4. **Fixtures working**: Test fixtures create/cleanup properly

### Common Issues

**"Project not found" errors**: Ensure `test_project` fixture is working

**Cleanup failures**: Old test data may interfere - check database

**Import errors**: Ensure Python path includes test directory

## Related Documentation

- Implementation: `/openspec/changes/add-enum-graphql-typed-errors/`
- GraphQL Schema: `/api/graph/schema/enum.graphql`
- Error Adapter: `/internal/interfaces/graphql/adapter/enum_error_adapter.go`
- Service Layer: `/internal/app/modeldesign/enum_service.go`

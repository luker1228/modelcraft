# ModelCraft Design-Time Integration Tests

## Test Directory Structure

```
tests/design/
├── project/          # Project domain tests
├── model/            # Model domain tests
├── field/            # Field domain tests
├── cluster/          # Database cluster tests
├── enum/             # Enum domain tests
└── conftest.py       # Pytest fixtures and configuration
```

## Available Test Modules

### Project Tests (`tests/design/project/`)
Tests for project management functionality:
- Project CRUD operations
- Project validation
- Project-Cluster relationships
- Project context isolation

### Model Tests (`tests/design/model/`)
Tests for model design functionality:
- Model CRUD operations
- Model deployment lifecycle
- Model validation
- Field management

### Field Tests (`tests/design/field/`)
Tests for field definition functionality:
- Field type validation
- Field constraints
- Relation definitions

### Cluster Tests (`tests/design/cluster/`)
Tests for database cluster management:
- Cluster CRUD operations
- Connection validation
- Cluster-Project relationships

### Enum Tests (`tests/design/enum/`)
Tests for enum definition functionality:
- Enum CRUD operations
- Enum value management
- Enum validation

## Test Execution Patterns

**Single module:**
```bash
pytest tests/design/project
```

**All design-time tests:**
```bash
pytest tests/design/
```

**Specific test file:**
```bash
pytest tests/design/project/test_project_crud.py
```

**Specific test case:**
```bash
pytest tests/design/project/test_project_crud.py::TestProjectCRUD::test_create_project
```

## HTML Report Location

HTML test reports are generated in:
```
tests/reports/test_report_{module}_{timestamp}.html
```

Example: `tests/reports/test_report_project_20260222_143022.html`

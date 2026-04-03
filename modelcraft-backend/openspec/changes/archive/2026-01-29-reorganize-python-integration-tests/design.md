# Design: Python Integration Test Directory Organization

## Overview

This design document outlines the rationale and patterns for organizing Python integration tests in a mid-size project. The goal is to provide a scalable structure that balances simplicity with future growth.

## Design Principles

### 1. Domain-Driven Organization

**Principle**: Organize tests by business domain, mirroring the main codebase's DDD structure.

**Rationale**:
- Cognitive consistency: Tests follow the same mental model as the application code
- Easy navigation: Developers working on a domain know where to find/add tests
- Clear boundaries: Each domain's tests are isolated and can evolve independently

**Application**:
```
automated/
├── project/      # Tests for project management domain
├── model/        # Tests for model domain (design + runtime)
├── cluster/      # Tests for database cluster domain
└── ...
```

### 2. Shared Infrastructure Centralization

**Principle**: Extract common test utilities to a dedicated location to reduce duplication.

**Rationale**:
- DRY (Don't Repeat Yourself): Avoid duplicating client setup, fixtures, and utilities
- Consistency: Shared code ensures uniform behavior across tests
- Maintainability: Changes to shared logic happen in one place

**Application**:
```
automated/
├── conftest.py           # Pytest shared fixtures (auto-discovered)
└── common/               # Explicit shared utilities
    ├── graphql_client.py # GraphQL client factory
    ├── fixtures.py       # Reusable fixtures
    └── assertions.py     # Custom assertion helpers
```

### 3. Shallow Hierarchy (Max 2 Levels)

**Principle**: Keep directory nesting shallow to avoid "directory hell".

**Rationale**:
- Simplicity: Easier to navigate with fewer levels
- Import paths: Shorter import statements (`from tests.automated.project import ...`)
- Scalability: Mid-size projects (10-30 test files) don't need deep nesting

**Application**:
```
✅ Good: automated/project/test_project_crud.py
❌ Bad:  automated/project/crud/create/test_create_project.py
```

**Future Growth**: If a domain grows to 5+ test files, split within the domain:
```
automated/model/
├── design/
│   ├── test_model_crud.py
│   └── test_field_operations.py
└── runtime/
    ├── test_graphql_queries.py
    └── test_aggregations.py
```

### 4. Separation of Tests and Tools

**Principle**: Distinguish between "test cases" and "test support tools".

**Rationale**:
- Clarity: Test runners (pytest) discover only actual tests
- Organization: Utilities are not tests and should not be in test directories
- Reusability: Tools can be imported by both tests and external scripts

**Application**:
```
tests/
├── automated/          # Test cases (pytest discovers here)
│   └── project/
│       └── test_*.py
└── utils/              # Support tools (not discovered by pytest)
    ├── config.py
    ├── health_check.py
    └── cleanup_test_data.py
```

### 5. Convention over Configuration

**Principle**: Follow pytest conventions for automatic discovery and reduced configuration.

**Rationale**:
- Standard patterns: `test_*.py` naming, `conftest.py` for fixtures
- Less boilerplate: Pytest auto-discovers tests without explicit configuration
- Community alignment: Familiar to anyone experienced with pytest

**Application**:
- File naming: `test_project_crud.py`, not `project_crud_test.py`
- Fixture location: `conftest.py` for shared fixtures, not scattered across files
- Test function naming: `test_create_project_success`, not `validate_project_creation`

## Architectural Patterns

### Pattern 1: Layered Test Structure

```
┌─────────────────────────────────────┐
│  Test Cases (test_*.py)             │  ← Business test logic
├─────────────────────────────────────┤
│  Domain Fixtures (conftest.py)      │  ← Domain-specific setup
├─────────────────────────────────────┤
│  Common Utilities (common/)         │  ← Shared tools
├─────────────────────────────────────┤
│  Base Configuration (utils/config)  │  ← Environment setup
└─────────────────────────────────────┘
```

**Benefits**:
- Clear dependencies (top layers depend on bottom layers)
- Easy to mock/override at any layer
- Testable test code (utilities can have their own tests)

### Pattern 2: Fixture Hierarchy

```python
# tests/automated/conftest.py (root fixtures)
@pytest.fixture(scope="session")
def test_config():
    """Load test configuration once per session"""
    return TestConfig()

@pytest.fixture(scope="session")
def graphql_client(test_config):
    """Create GraphQL client for the session"""
    return create_graphql_client(test_config.get_design_graphql_url())

# tests/automated/project/conftest.py (domain fixtures)
@pytest.fixture(scope="module")
def created_projects(graphql_client):
    """Track created projects for cleanup"""
    projects = []
    yield projects
    # Cleanup after module tests complete
    for project_id in projects:
        delete_project(graphql_client, project_id)
```

**Benefits**:
- Scope management: Session, module, or function-scoped fixtures
- Automatic cleanup: `yield` pattern ensures cleanup happens
- Fixture chaining: Higher-level fixtures depend on lower-level ones

### Pattern 3: Test Data Builders

Instead of inline test data, use builder functions for reusability:

```python
# tests/automated/common/test_data.py
def build_project_input(project_id=None, title=None, **kwargs):
    """Build a project creation input with sensible defaults"""
    return {
        "projectId": project_id or f"test-project-{uuid.uuid4().hex[:8]}",
        "title": title or "Test Project",
        "description": kwargs.get("description", ""),
        "status": kwargs.get("status", "ACTIVE"),
    }

# tests/automated/project/test_project_crud.py
def test_create_project(graphql_client):
    input_data = build_project_input(title="My Project")
    result = create_project(graphql_client, input_data)
    assert result["title"] == "My Project"
```

**Benefits**:
- Reduced duplication: Test data definition in one place
- Flexibility: Override specific fields while using defaults
- Readability: Test intent is clearer (`build_project_input(title="X")` vs inline dict)

## Migration Strategy

### Phase-by-Phase Approach

```
Current State        Transition             Target State
─────────────        ──────────             ────────────
automated/           1. Create structure    automated/
├── test_1.py   →    2. Move files     →   ├── project/
├── test_2.py        3. Update imports     │   └── test_*.py
└── ...              4. Extract shared     ├── model/
                     5. Cleanup            │   └── test_*.py
                                           └── common/
                                               └── utilities
```

**Key Migration Steps**:
1. **Non-breaking addition**: Create new structure alongside old
2. **Incremental migration**: Move one domain at a time
3. **Continuous validation**: Run tests after each migration step
4. **Safe cleanup**: Remove old structure only after full verification

### Import Path Updates

```python
# Before
from config import config

# After
from tests.utils.config import config
```

**Strategy**: Use IDE refactoring tools or regex search-replace for batch updates.

## Trade-offs and Alternatives Considered

### Alternative 1: Flat Structure (Keep Current)

```
automated/
├── test_project_crud.py
├── test_model_crud.py
├── test_cluster_crud.py
└── ... (30+ files)
```

**Rejected because**:
- Does not scale beyond 10-15 files
- Hard to find related tests
- No clear organization principle

### Alternative 2: Test Type Organization

```
automated/
├── unit/
├── integration/
└── e2e/
```

**Rejected because**:
- All current tests are integration tests
- Doesn't help with test discoverability within a type
- Would still need domain organization within `integration/`

### Alternative 3: Feature-Based Organization

```
automated/
├── feature_project_management/
├── feature_model_design/
└── ...
```

**Rejected because**:
- "Feature" is less precise than "domain"
- Doesn't align with codebase structure (DDD domains)
- Less clear boundaries between features vs domains

### Trade-off: `utils/` vs Root-Level Scripts

**Decision**: Move utilities to `tests/utils/`

**Trade-off**:
- ✅ Pro: Clear separation of tools vs tests
- ✅ Pro: `utils/` is a Python package with `__init__.py`
- ⚠️ Con: Slightly longer import paths
- ⚠️ Con: Additional directory level

**Mitigation**: The benefits of clear organization outweigh the minor inconvenience of longer imports.

## Future Evolution Scenarios

### Scenario 1: Growth to 50+ Test Files

If the project grows significantly:

```
automated/
├── project/
│   ├── crud/
│   │   ├── test_create.py
│   │   ├── test_read.py
│   │   └── test_delete.py
│   └── isolation/
│       └── test_resource_isolation.py
└── model/
    ├── design/
    └── runtime/
```

**Strategy**: Add sub-categorization within domains when a domain exceeds 5-7 test files.

### Scenario 2: Addition of New Test Types

If performance or E2E tests are added:

```
tests/
├── automated/          # Integration tests (current)
├── performance/        # Load and stress tests
└── e2e/                # End-to-end user journey tests
```

**Strategy**: Keep distinct test types as siblings to `automated/`, not nested within it.

### Scenario 3: Multi-Service Testing

If the project adopts microservices:

```
tests/
├── automated/
│   ├── service_a/
│   │   ├── project/
│   │   └── model/
│   └── service_b/
│       └── cluster/
└── integration/        # Cross-service tests
```

**Strategy**: Introduce service-level grouping above domain grouping.

## Validation and Success Metrics

### Immediate Success Criteria

- ✅ All existing tests pass after migration
- ✅ Test discovery works (`pytest --collect-only` finds all tests)
- ✅ CI/CD pipelines run without changes
- ✅ Import paths are consistent and correct

### Long-term Success Metrics

- **Discoverability**: New developers can find relevant tests in < 30 seconds
- **Maintainability**: Adding a new test takes < 5 minutes (including finding the right location)
- **Scalability**: Structure supports growth to 30+ test files without reorganization
- **Code reuse**: Shared fixtures and utilities reduce test code by ~20%

## References

- **Pytest Best Practices**: https://docs.pytest.org/en/stable/goodpractices.html
- **Test Organization Patterns**: https://martinfowler.com/articles/practical-test-pyramid.html
- **ModelCraft DDD Architecture**: `docs/00-overview/architecture.md`

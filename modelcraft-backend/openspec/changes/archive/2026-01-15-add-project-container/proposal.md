# Proposal: Add Project Container

## Overview

Introduce a **Project** concept as a top-level container that organizes Clusters, Models, and Enums into logical workspaces. Each project has a human-readable ID for easy identification and reference.

## Motivation

### Current Pain Points
1. **Lack of Organization**: Models, clusters, and enums exist without a clear organizational hierarchy
2. **No Isolation Boundary**: Different teams or applications cannot easily separate their resources
3. **Complex Identification**: Model locator requires specifying `cluster_name.database_name.model_name` which becomes verbose
4. **Global Namespace Pollution**: Enum names must be globally unique, limiting naming flexibility

### Benefits of Project Container
1. **Logical Grouping**: Projects provide a natural boundary for organizing related resources
2. **Simplified Identification**: Projects use human-readable IDs (e.g., "ecommerce", "crm") for easy reference
3. **Namespace Isolation**: Resources within a project can use shorter, context-relevant names
4. **Multi-Tenancy Foundation**: Projects can support future multi-tenant requirements
5. **Better Access Control**: Projects enable future permission management at the project level

## Scope

### In Scope
1. Create Project domain entity with human-readable ID
2. Modify Cluster to belong to a Project
3. Modify Model to belong to a Project (replacing current global scope)
4. Modify Enum to belong to a Project (replacing current global scope)
5. Update database schema with project_id foreign keys
6. Create GraphQL and REST APIs for project CRUD operations
7. **BREAKING**: Require project_id in all APIs (no backward compatibility)
8. **BREAKING**: Change modelId and clusterId to be unique only within project
9. Enforce strict project isolation - no cross-project joins or references

### Out of Scope
1. Multi-tenancy and access control mechanisms (future work)
2. Project-level quotas or resource limits (future work)
3. Project templates or cloning features (future work)
4. Cross-project resource references (EXPLICITLY NOT SUPPORTED)
5. Project archival or soft-delete (can be added later)
6. Backward compatibility with existing data (breaking change)

## Design Decisions

### 1. Project ID Format
- **Decision**: Use human-readable string ID (e.g., "ecommerce", "crm", "mobile-app")
- **Rationale**:
  - Easy to remember and communicate
  - Better developer experience in APIs and URLs
  - Similar to GitHub repository names or Kubernetes namespaces
- **Constraints**:
  - 2-64 characters
  - Lowercase alphanumeric and hyphens only
  - Must start with letter
  - Globally unique across the system

### 2. Hierarchical Model Locator
- **Current**: `cluster_name.database_name.model_name`
- **New**: `project_id.cluster_name.database_name.model_name`
- **Rationale**: Extend existing pattern for consistency
- **BREAKING**: Existing APIs must now include project_id parameter

### 3. Scoped ID Strategy (BREAKING CHANGE)
- **Decision**: Model IDs and Cluster IDs are unique only within their project
- **Rationale**:
  - Simpler implementation - no global UUID generation needed
  - Better performance - shorter IDs, simpler lookups
  - Natural namespace per project
- **Format**:
  - Cluster ID: project-scoped integer or short string
  - Model ID: project-scoped integer or short string
  - Full reference requires project_id context
- **BREAKING**: Existing IDs may conflict if data needs migration

### 4. No Cross-Project References
- **Decision**: Models in different projects CANNOT join or reference each other
- **Rationale**:
  - Enforces clean boundaries and isolation
  - Prevents complex dependency graphs across projects
  - Simpler query planning and execution
- **Constraints**:
  - Relations can only reference models within same project
  - Enum references must be within same project
  - Cluster references must be within same project

### 5. Unique Constraints
- **Project ID**: Globally unique
- **Cluster Name**: Unique within a project (changed from globally unique)
- **Model Name**: Unique within `project_id + cluster_name + database_name`
- **Enum Name**: Unique within a project (changed from globally unique)

### 6. Database Schema Strategy
- Add `project_id` column to:
  - `database_clusters` table
  - `models` table
  - `model_enums` table
- Update unique indexes to include `project_id`
- Add foreign key constraints referencing `projects.id`
- **BREAKING**: Change ID columns to be project-scoped:
  - Cluster: composite PK (project_id, id) or change to (project_id, name)
  - Model: composite PK (project_id, id) or single PK with project_id + auto-increment
- No migration support - fresh start recommended

## Implementation Approach

### Phase 1: Core Domain (Breaking Changes)
1. Create Project domain entity and repository
2. Add project_id to Cluster, Model, Enum entities
3. Update ModelLocator to include project_id
4. **BREAKING**: Remove support for APIs without project_id
5. **BREAKING**: Change ID generation to be project-scoped

### Phase 2: Database Schema (Breaking Changes)
1. Create `projects` table
2. **BREAKING**: Drop and recreate dependent tables with new schema:
   - `database_clusters` with project-scoped IDs
   - `models` with project-scoped IDs
   - `model_enums` with project-scoped IDs
3. Update unique indexes to include `project_id`
4. Add foreign key constraints with CASCADE delete

### Phase 3: API Layer (Breaking Changes)
1. Add GraphQL mutations and queries for project management
2. Add REST endpoints for project CRUD
3. **BREAKING**: Update ALL existing APIs to REQUIRE project_id parameter
4. **BREAKING**: Remove APIs that don't include project context
5. Add validation to prevent cross-project references

### Phase 4: Validation & Testing
1. Unit tests for project domain logic
2. Repository integration tests with project isolation
3. API integration tests requiring project_id
4. Validation tests for cross-project reference prevention

## Dependencies

### Upstream Dependencies
- None (this is a foundational feature)

### Downstream Impact
- **BREAKING**: All features using Models, Clusters, or Enums MUST include project context
- **BREAKING**: All existing APIs will break and require client updates
- Existing changes in `openspec/changes/` will need updates:
  - `add-goja-interceptor-for-operations` - MUST update to require project context
  - `add-schema-based-model-operations` - MUST operate within project scope
- Runtime GraphQL queries MUST include project context
- All model relations MUST be within same project (no cross-project joins)

## Migration Strategy (BREAKING CHANGE)

### No Automatic Migration
- **BREAKING**: This is a clean break - no automatic data migration
- Users must:
  1. Export existing data before upgrade
  2. Create projects manually
  3. Re-import data with project_id specified
  4. Update all API clients to include project_id

### Fresh Database Schema
```sql
-- 1. Create projects table
CREATE TABLE projects (
  id VARCHAR(64) PRIMARY KEY,
  title VARCHAR(255) NOT NULL,
  ...
);

-- 2. Recreate dependent tables with project-scoped IDs
DROP TABLE IF EXISTS model_relations;
DROP TABLE IF EXISTS field_definitions;
DROP TABLE IF EXISTS models;
DROP TABLE IF EXISTS database_clusters;
DROP TABLE IF EXISTS model_enums;

-- Create new tables with project_id and simplified IDs
CREATE TABLE database_clusters (
  project_id VARCHAR(64) NOT NULL,
  id INT AUTO_INCREMENT,
  name VARCHAR(64) NOT NULL,
  ...
  PRIMARY KEY (project_id, id),
  UNIQUE KEY (project_id, name),
  FOREIGN KEY (project_id) REFERENCES projects(id) ON DELETE CASCADE
);

-- Similar for models and enums
```

### API Migration
- **BREAKING**: ALL API endpoints now REQUIRE project_id
- **BREAKING**: Remove all endpoints without project context
- GraphQL queries REQUIRE `projectId` parameter (not optional)
- REST endpoints ONLY support `/api/v1/projects/{project_id}/...` pattern

## Open Questions

1. **Q**: Should project ID be mutable after creation?
   - **Proposed**: No, immutable for simplicity and stability
   - **Rationale**: Changing project ID would require updating all dependent resources
   - **Decision**: Immutable

2. **Q**: What happens when a project is deleted?
   - **Proposed**: Cascade delete all dependent resources (clusters, models, enums)
   - **Rationale**: Clean separation, no orphaned data
   - **Decision**: CASCADE DELETE enforced by foreign keys

3. **Q**: Should model IDs be globally unique UUIDs or project-scoped integers?
   - **Option A**: Project-scoped auto-increment integers (simpler, shorter)
   - **Option B**: Composite key (project_id, auto_increment_id)
   - **Option C**: Continue using global UUIDs
   - **Recommendation**: Option B for simplicity and clarity
   - **Decision needed**: Confirm ID strategy

4. **Q**: How to handle cross-project relation attempts?
   - **Proposed**: Validation error at API layer before database insert
   - **Error message**: "Cannot create relation to model in different project"
   - **Decision**: Validate during relation creation, block at API level

## Success Criteria

1. ✓ Projects can be created with human-readable IDs
2. ✓ Clusters, Models, and Enums belong to projects
3. ✓ All APIs require project_id parameter (no backward compatibility)
4. ✓ Model IDs and Cluster IDs are scoped within projects
5. ✓ Cross-project references are prevented by validation
6. ✓ Unique constraints enforced at project scope
7. ✓ Foreign key CASCADE DELETE works correctly
8. ✓ Integration tests pass with strict project isolation
9. ✓ API clients updated to include project_id in all requests
10. ✓ Documentation clearly marks this as a BREAKING CHANGE

## Timeline

- **Proposal Review**: 1-2 days
- **Implementation**: Phase 1-4 (estimated 5-7 days)
- **Testing & Validation**: 2-3 days
- **Deployment**: 1 day

## Related Work

- Future: Project-level access control
- Future: Project quotas and resource limits
- Future: Project templates and cloning

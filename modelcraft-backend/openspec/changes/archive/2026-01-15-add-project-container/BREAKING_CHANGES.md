# BREAKING CHANGES Summary

## ⚠️ This is a Breaking Change

The Project Container feature introduces **BREAKING CHANGES** that require manual data migration and API client updates. There is **NO backward compatibility**.

## What's Breaking

### 1. API Changes (BREAKING)
- **ALL APIs now REQUIRE `project_id` parameter**
- No optional project_id - it's mandatory
- REST endpoints changed to `/api/v1/projects/{project_id}/models` format
- GraphQL queries require `projectId` argument (not optional)
- Old API calls without project_id will fail with error

### 2. ID Scoping (BREAKING)
- **Model IDs are now project-scoped** (not globally unique)
- **Cluster IDs are now project-scoped** (not globally unique)
- Same ID can exist in different projects
- Lookups require both `project_id` and resource `id`

### 3. Database Schema (BREAKING)
- Tables `models`, `database_clusters`, `model_enums` require restructuring
- Add `project_id` column as part of primary key or composite key
- Unique constraints changed to include `project_id`
- **No automatic migration** - fresh schema recommended

### 4. Project Isolation (NEW CONSTRAINT)
- **Models in different projects CANNOT join**
- **Cross-project references are blocked**
- Relations must be within same project
- Enum references must be within same project
- Cluster references must be within same project

### 5. No "Default" Project
- **Removed backward compatibility with "default" project**
- Users must create projects explicitly
- All resources must belong to a named project
- No automatic migration of existing data

## Migration Path

### For Users Upgrading

1. **Export existing data**
   ```bash
   # Export all models, clusters, enums before upgrade
   curl GET /api/v1/models > models_backup.json
   curl GET /api/v1/clusters > clusters_backup.json
   curl GET /api/v1/enums > enums_backup.json
   ```

2. **Upgrade ModelCraft** (breaks existing data)
   - Deploy new version with project support
   - Database schema will be recreated

3. **Create projects**
   ```bash
   curl POST /api/v1/projects -d '{"id": "my-app", "title": "My Application"}'
   ```

4. **Re-import data with project_id**
   ```bash
   curl POST /api/v1/projects/my-app/models -d '{...}'
   curl POST /api/v1/projects/my-app/clusters -d '{...}'
   curl POST /api/v1/projects/my-app/enums -d '{...}'
   ```

5. **Update all API clients**
   - Add `projectId` to all GraphQL queries
   - Change REST endpoints to include `/projects/{projectId}/`
   - Update model/cluster lookups to include project context

### For Fresh Installations

No migration needed - just create projects and resources normally.

## API Changes Examples

### GraphQL (Before → After)

**Before (BROKEN):**
```graphql
query {
  models {
    id
    name
  }
}
```

**After (REQUIRED):**
```graphql
query {
  models(projectId: "ecommerce") {
    id
    projectId
    name
  }
}
```

### REST (Before → After)

**Before (BROKEN):**
```
GET /api/v1/models
POST /api/v1/models
```

**After (REQUIRED):**
```
GET /api/v1/projects/ecommerce/models
POST /api/v1/projects/ecommerce/models
```

## Database Schema Changes

### Before (Simple UUID-based)

```sql
CREATE TABLE models (
  id VARCHAR(36) PRIMARY KEY,
  name VARCHAR(64),
  ...
  UNIQUE KEY (cluster_name, database_name, name)
);
```

### After (Project-scoped composite key)

```sql
CREATE TABLE projects (
  id VARCHAR(64) PRIMARY KEY,
  title VARCHAR(255),
  ...
);

CREATE TABLE models (
  project_id VARCHAR(64),
  id INT AUTO_INCREMENT,
  name VARCHAR(64),
  ...
  PRIMARY KEY (project_id, id),
  UNIQUE KEY (project_id, cluster_name, database_name, name),
  FOREIGN KEY (project_id) REFERENCES projects(id) ON DELETE CASCADE
);
```

## What's NOT Supported

### ❌ Cross-Project Operations

```graphql
# ❌ NOT ALLOWED
query {
  model(projectId: "ecommerce", name: "Order") {
    relation {
      targetModel {
        # This fails if targetModel is in different project
        projectId  # Could be "crm" - NOT ALLOWED
      }
    }
  }
}
```

**Error**: "Cannot create relation to model in different project"

### ❌ Backward Compatible APIs

```
# ❌ REMOVED - will return 404
GET /api/v1/models

# ❌ REMOVED - will return error
query { models { id } }  # Missing required projectId
```

## Timeline

**Estimated Implementation**: 2-3 weeks
- Week 1: Domain layer + database schema
- Week 2: API layer + validation
- Week 3: Testing + documentation

**Deployment**: Coordinate with users for downtime
- Backup existing data
- Fresh schema deployment
- Provide migration scripts/tools
- Update documentation

## Benefits (Why Breaking Change Is Worth It)

1. **Clean Architecture**: No technical debt from backward compatibility
2. **Better Performance**: Simpler queries with project-scoped IDs
3. **Strong Isolation**: Prevents accidental cross-project data leaks
4. **Simplified Logic**: No "default" project special cases
5. **Future-Proof**: Foundation for multi-tenancy and access control

## Questions?

Review the full proposal:
- `openspec/changes/add-project-container/proposal.md`
- `openspec/changes/add-project-container/design.md`
- `openspec/changes/add-project-container/tasks.md`

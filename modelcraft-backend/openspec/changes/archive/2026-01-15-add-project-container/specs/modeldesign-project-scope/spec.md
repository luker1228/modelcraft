# Spec: Model Design - Project Scope

## Overview

This spec modifies the existing Model Design capability to operate within project scope. Models now belong to projects, and model uniqueness is enforced within the project boundary.

## MODIFIED Requirements

### Requirement: Model Belongs to Project

Every model MUST belong to exactly one project. The project provides the namespace for model organization.

**Changes from Previous:**
- Add `project_id` field to Model entity
- Update ModelLocator to include project_id
- Model uniqueness now scoped to project

**Acceptance Criteria:**
- Model has required `project_id` field
- Model locator includes project_id as first component
- Model full path format: `project_id.cluster_name.database_name.model_name`

#### Scenario: Create model in specific project

**Given** project "ecommerce" exists
**And** cluster "prod-db" exists in project "ecommerce"
**When** user creates model "User" in:
- projectId: "ecommerce"
- clusterName: "prod-db"
- databaseName: "app_db"

**Then** model is created successfully
**And** model.projectId equals "ecommerce"
**And** model full path is "ecommerce.prod-db.app_db.User"

#### Scenario: Create model in non-existent project

**Given** no project exists with ID "nonexistent"
**When** user attempts to create model in project "nonexistent"
**Then** creation fails with error "project not found"

#### Scenario: Create duplicate model name in same project

**Given** project "ecommerce" exists
**And** model "User" exists in "ecommerce.prod-db.app_db.User"
**When** user attempts to create another model "User" in "ecommerce.prod-db.app_db"
**Then** creation fails with error "model name already exists in project"

#### Scenario: Create same model name in different projects

**Given** projects "ecommerce" and "crm" exist
**And** model "User" exists in "ecommerce.prod-db.app_db.User"
**When** user creates model "User" in "crm.prod-db.app_db.User"
**Then** model is created successfully
**And** both models exist independently
**And** models have different project_id values

---

### Requirement: Model Locator Includes Project

ModelLocator MUST include project_id as part of the locator tuple.

**Changes from Previous:**
- Add ProjectID field to ModelLocator struct
- Update GetFullPath to include project_id
- Update Validate to check project_id

**Acceptance Criteria:**
- ModelLocator has ProjectID field
- GetFullPath returns format: `{project_id}.{cluster}.{database}.{model}`
- Validate checks all four components are non-empty

#### Scenario: Construct model locator with project

**Given** model locator is created with:
- ProjectID: "ecommerce"
- ClusterName: "prod-db"
- DatabaseName: "app_db"
- ModelName: "User"

**When** GetFullPath is called
**Then** result is "ecommerce.prod-db.app_db.User"

#### Scenario: Validate model locator requires project

**Given** model locator with:
- ProjectID: "" (empty)
- ClusterName: "prod-db"
- DatabaseName: "app_db"
- ModelName: "User"

**When** Validate is called
**Then** validation fails with error "project ID is required"

---

### Requirement: Query Models by Project

Users MUST be able to query models filtered by project.

**Changes from Previous:**
- Add projectId parameter to list/search operations
- Filter results by project_id in repository layer

**Acceptance Criteria:**
- Can list models in specific project
- Can search models within project scope
- Models from other projects are not visible in query results

#### Scenario: List models in specific project

**Given** models exist:
- "User" in project "ecommerce"
- "Product" in project "ecommerce"
- "Contact" in project "crm"

**When** user lists models with projectId "ecommerce"
**Then** response contains 2 models: "User" and "Product"
**And** response does not contain "Contact"


**Given** models exist in multiple projects
**When** user lists models without specifying projectId
**Then** only models in "default" project are returned
**And** models from other projects are excluded

---

### Requirement: Model Cluster Reference Must Be In Same Project

When creating a model, the referenced cluster MUST exist in the same project.

**New Constraint:**
- Cluster validation checks project_id match
- Cross-project cluster references are not allowed

**Acceptance Criteria:**
- Model creation validates cluster exists in same project
- Attempting to use cluster from different project fails

#### Scenario: Create model with cluster in same project

**Given** project "ecommerce" exists
**And** cluster "prod-db" exists in project "ecommerce"
**When** user creates model in project "ecommerce" referencing cluster "prod-db"
**Then** model is created successfully

#### Scenario: Create model with cluster in different project

**Given** project "ecommerce" and "crm" exist
**And** cluster "prod-db" exists in project "crm"
**When** user attempts to create model in project "ecommerce" referencing cluster "prod-db"
**Then** creation fails with error "cluster not found in project"
**And** error message explains cluster must be in same project

---


## ADDED Requirements

### Requirement: Delete Project Cascades to Models

When a project is deleted, all models in that project MUST be deleted.

**Acceptance Criteria:**
- Deleting project cascades to all models in that project
- Models in other projects are unaffected
- Foreign key constraint enforces cascade

#### Scenario: Delete project removes all models

**Given** project "old-app" exists
**And** models "User", "Product" exist in project "old-app"
**And** model "Contact" exists in project "crm"
**When** user deletes project "old-app" (with force flag)
**Then** project "old-app" is deleted
**And** models "User" and "Product" are deleted
**And** model "Contact" in project "crm" still exists

---

### Requirement: Model IDs Are Project-Scoped

Model IDs MUST be unique only within their project, not globally.

**Acceptance Criteria:**
- Model ID can be an auto-increment integer within project scope
- Two models in different projects can have the same ID
- Model lookup requires both project_id and model_id
- Full model reference format: `{project_id}/{model_id}` or composite key

#### Scenario: Same model ID in different projects

**Given** projects "ecommerce" and "crm" exist
**When** model with ID "1" is created in project "ecommerce"
**And** model with ID "1" is created in project "crm"
**Then** both models exist independently
**And** models are distinct entities
**And** querying model "1" requires project context

#### Scenario: Model lookup requires project context

**Given** model with ID "5" exists in project "ecommerce"
**And** model with ID "5" exists in project "crm"
**When** user queries model "5" in project "ecommerce"
**Then** returns model from "ecommerce" project only
**And** does not return model from "crm" project

---

## Database Schema Changes

### Updated Table: models

```sql
-- Add project_id column
ALTER TABLE `models`
  ADD COLUMN `project_id` VARCHAR(64) NOT NULL DEFAULT 'default' COMMENT '所属项目ID' AFTER `id`;

-- Add foreign key constraint
ALTER TABLE `models`
  ADD CONSTRAINT `fk_model_project`
  FOREIGN KEY (`project_id`)
  REFERENCES `projects` (`id`)
  ON DELETE CASCADE;

-- Update unique index to include project_id
ALTER TABLE `models`
  DROP INDEX `idx_models_name`;

ALTER TABLE `models`
  ADD UNIQUE KEY `idx_models_name` (`project_id`, `cluster_name`, `database_name`, `name`)
  COMMENT '项目内模型名称唯一';

-- Add index for project-based queries
ALTER TABLE `models`
  ADD KEY `idx_project_status` (`project_id`, `status`);
```

## API Changes

### GraphQL

```graphql
# Updated Model type
type Model {
  id: ID!
  projectId: ID!          # NEW FIELD
  name: String!
  # ... other fields
}

# Updated query
extend type Query {
  # New: filter by project
  models(projectId: ID, status: ModelStatus): [Model!]!

  # Updated: require project for lookups
  model(projectId: ID!, name: String!, clusterName: String!, databaseName: String!): Model
}

# Updated mutation
input CreateModelInput {
  projectId: ID!         # NEW REQUIRED FIELD
  name: String!
  clusterName: String!
  databaseName: String!
  # ... other fields
}
```

### REST

```
# New project-scoped endpoints
GET    /api/v1/projects/:projectId/models
POST   /api/v1/projects/:projectId/models
GET    /api/v1/projects/:projectId/models/:modelId
PUT    /api/v1/projects/:projectId/models/:modelId
DELETE /api/v1/projects/:projectId/models/:modelId

# Backward compatible (uses "default" project)
GET    /api/v1/models
POST   /api/v1/models
```

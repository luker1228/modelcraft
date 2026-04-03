# Spec: Enum Definition - Project Scope

## Overview

This spec modifies the existing Enum Definition capability to operate within project scope. Enums now belong to projects, and enum names are unique within each project.

## MODIFIED Requirements

### Requirement: Enum Belongs to Project

Every enum definition MUST belong to exactly one project.

**Changes from Previous:**
- Add `project_id` field to EnumDefinition entity
- Enum name uniqueness scoped to project (not globally unique)
- Enum queries filtered by project

**Acceptance Criteria:**
- Enum has required `project_id` field
- Enum name unique within project, can duplicate across projects
- Enum full identifier format: `project_id.enum_name`

#### Scenario: Create enum in specific project

**Given** project "ecommerce" exists
**When** user creates enum "OrderStatus" in project "ecommerce" with options:
- "pending"
- "confirmed"
- "shipped"
- "delivered"

**Then** enum is created successfully
**And** enum.projectId equals "ecommerce"
**And** enum is visible when querying project "ecommerce"

#### Scenario: Create enum in non-existent project

**Given** no project exists with ID "nonexistent"
**When** user attempts to create enum in project "nonexistent"
**Then** creation fails with error "project not found"

#### Scenario: Create duplicate enum name in same project

**Given** project "ecommerce" exists
**And** enum "OrderStatus" exists in project "ecommerce"
**When** user attempts to create another enum "OrderStatus" in project "ecommerce"
**Then** creation fails with error "enum name already exists in project"

#### Scenario: Create same enum name in different projects

**Given** projects "ecommerce" and "crm" exist
**And** enum "Status" exists in project "ecommerce" with options ["active", "inactive"]
**When** user creates enum "Status" in project "crm" with options ["lead", "customer"]
**Then** enum is created successfully
**And** both enums exist independently
**And** enums have different project_id and different option sets

---

### Requirement: Query Enums by Project

Users MUST be able to query enums filtered by project.

**Changes from Previous:**
- Add projectId parameter to list/search operations
- Filter results by project_id in repository layer

**Acceptance Criteria:**
- Can list enums in specific project
- Enums from other projects are not visible in query results
- Enum lookup by name searches within project scope

#### Scenario: List enums in specific project

**Given** enums exist:
- "OrderStatus" in project "ecommerce"
- "ProductCategory" in project "ecommerce"
- "LeadStatus" in project "crm"

**When** user lists enums with projectId "ecommerce"
**Then** response contains 2 enums: "OrderStatus" and "ProductCategory"
**And** response does not contain "LeadStatus"

#### Scenario: Get enum by name within project

**Given** enum "Status" exists in both projects "ecommerce" and "crm"
**When** user gets enum "Status" in project "ecommerce"
**Then** returns enum from project "ecommerce"
**And** does not return enum from project "crm"


**Given** enums exist in multiple projects
**When** user lists enums without specifying projectId
**Then** only enums in "default" project are returned
**And** enums from other projects are excluded

---

### Requirement: Field Enum References Must Be In Same Project

When a field references an enum, the enum MUST exist in the same project as the model.

**New Constraint:**
- Enum validation checks project_id match with model
- Cross-project enum references are not allowed

**Acceptance Criteria:**
- Field enum reference validates enum exists in same project as model
- Attempting to use enum from different project fails
- Field enum association stores project context

#### Scenario: Create field with enum in same project

**Given** project "ecommerce" exists
**And** model "Order" exists in project "ecommerce"
**And** enum "OrderStatus" exists in project "ecommerce"
**When** user creates field "status" in model "Order" with enum reference "OrderStatus"
**Then** field is created successfully
**And** field can use enum options from "OrderStatus"

#### Scenario: Create field with enum in different project

**Given** projects "ecommerce" and "crm" exist
**And** model "Order" exists in project "ecommerce"
**And** enum "Status" exists in project "crm"
**When** user attempts to create field in model "Order" referencing enum "Status"
**Then** creation fails with error "enum not found in project"
**And** error message explains enum must be in same project as model

---

### Requirement: Update Enum Unique Constraint

Enum name uniqueness MUST be enforced at project scope, not globally.

**Changes from Previous:**
- Remove global unique constraint on enum name
- Add composite unique constraint on (project_id, name)

**Acceptance Criteria:**
- Same enum name can exist in multiple projects
- Duplicate enum names within same project are rejected
- Database constraint enforces uniqueness

#### Scenario: Database enforces project-scoped uniqueness

**Given** database has enum "Status" in project "ecommerce"
**When** direct database insert attempts to add enum "Status" in project "ecommerce"
**Then** database raises unique constraint violation error
**And** insert is rolled back

---


## ADDED Requirements

### Requirement: Delete Project Cascades to Enums

When a project is deleted, all enums in that project MUST be deleted.

**Acceptance Criteria:**
- Deleting project cascades to all enums in that project
- Enums in other projects are unaffected
- Foreign key constraint enforces cascade
- Deleting enum may be restricted if referenced by fields (existing behavior)

#### Scenario: Delete project removes all enums

**Given** project "old-app" exists
**And** enums "Status", "Priority" exist in project "old-app"
**And** enum "Category" exists in project "active-app"
**When** user deletes project "old-app" (with force flag)
**Then** project "old-app" is deleted
**And** enums "Status" and "Priority" are deleted
**And** enum "Category" in project "active-app" still exists

#### Scenario: Prevent project deletion with enums in use

**Given** project "ecommerce" exists
**And** enum "OrderStatus" exists in project "ecommerce"
**And** model "Order" has field referencing enum "OrderStatus"
**When** user attempts to delete project "ecommerce"
**Then** deletion may be prevented due to enum in use
**Or** cascade deletion removes model fields, then enum, then project

---

## Database Schema Changes

### Updated Table: model_enums

```sql
-- Add project_id column
ALTER TABLE `model_enums`
  ADD COLUMN `project_id` VARCHAR(64) NOT NULL DEFAULT 'default' COMMENT '所属项目ID' AFTER `id`;

-- Add foreign key constraint
ALTER TABLE `model_enums`
  ADD CONSTRAINT `fk_enum_project`
  FOREIGN KEY (`project_id`)
  REFERENCES `projects` (`id`)
  ON DELETE CASCADE;

-- Update unique index to include project_id
ALTER TABLE `model_enums`
  DROP INDEX `idx_model_enums_name`;

ALTER TABLE `model_enums`
  ADD UNIQUE KEY `idx_enum_name` (`project_id`, `name`)
  COMMENT '项目内枚举名称唯一';

-- Add index for project-based queries
ALTER TABLE `model_enums`
  ADD KEY `idx_project` (`project_id`);
```

## API Changes

### GraphQL

```graphql
# Updated Enum type
type Enum {
  id: ID!
  projectId: ID!          # NEW FIELD
  name: String!
  title: String!
  description: String
  options: [EnumOption!]!
  isMultiSelect: Boolean!
  createdAt: DateTime!
  updatedAt: DateTime!
}

# Updated query
extend type Query {
  # New: filter by project
  enums(projectId: ID): [Enum!]!

  # Updated: require project for lookups
  enum(projectId: ID!, name: String!): Enum
  enumById(id: ID!): Enum  # ID-based lookup still works
}

# Updated mutation
input CreateEnumInput {
  projectId: ID!         # NEW REQUIRED FIELD
  name: String!
  title: String!
  description: String
  options: [EnumOptionInput!]!
  isMultiSelect: Boolean
}

input UpdateEnumInput {
  id: ID!
  # projectId cannot be changed
  title: String
  description: String
  options: [EnumOptionInput!]
}
```

### REST

```
# New project-scoped endpoints
GET    /api/v1/projects/:projectId/enums
POST   /api/v1/projects/:projectId/enums
GET    /api/v1/projects/:projectId/enums/:enumName
PUT    /api/v1/projects/:projectId/enums/:enumId
DELETE /api/v1/projects/:projectId/enums/:enumId

# Backward compatible (uses "default" project)
GET    /api/v1/enums
POST   /api/v1/enums
GET    /api/v1/enums/:enumId
```

## Repository Layer Changes

### Updated Interface: EnumDefinitionRepository

```go
type EnumDefinitionRepository interface {
    // Updated: GetByName now includes projectId
    GetByName(ctx context.Context, projectId, name string) (*EnumDefinition, error)

    // New: List by project
    ListByProject(ctx context.Context, projectId string) ([]*EnumDefinition, error)

    // New: Check existence in project
    ExistsByNameInProject(ctx context.Context, projectId, name string) (bool, error)

    // Existing methods remain, updated implementations
    Create(ctx context.Context, enum *EnumDefinition) error
    Update(ctx context.Context, enum *EnumDefinition) error
    GetByID(ctx context.Context, id string) (*EnumDefinition, error)
    Delete(ctx context.Context, id string) error
    // ... other methods
}
```

## Validation Changes

### Enum Creation Validation

```go
func (s *EnumAppService) CreateEnum(
    ctx context.Context,
    projectId, name, title, description string,
    options []EnumOption,
    isMultiSelect bool,
) (*EnumDefinition, error) {
    // 1. Validate project exists
    _, err := s.projectRepo.GetByID(ctx, projectId)
    if err != nil {
        return nil, bizerrors.Errorf("project not found: %s", projectId)
    }

    // 2. Check name uniqueness within project
    exists, err := s.enumRepo.ExistsByNameInProject(ctx, projectId, name)
    if err != nil {
        return nil, err
    }
    if exists {
        return nil, bizerrors.Errorf("enum name already exists in project: %s", name)
    }

    // 3. Validate enum options
    if len(options) == 0 {
        return nil, bizerrors.Errorf("enum must have at least one option")
    }

    // 4. Create enum
    // ... rest of implementation
}
```

### Field Enum Reference Validation

```go
func (s *FieldAppService) CreateFieldWithEnum(
    ctx context.Context,
    modelId, fieldName, enumName string,
) (*FieldDefinition, error) {
    // 1. Get model to determine project
    model, err := s.modelRepo.GetByID(ctx, modelId)
    if err != nil {
        return nil, err
    }

    // 2. Validate enum exists in same project
    enum, err := s.enumRepo.GetByName(ctx, model.ProjectID, enumName)
    if err != nil {
        return nil, bizerrors.Errorf("enum not found in project: %s", enumName)
    }

    // 3. Create field with enum reference
    // ... rest of implementation
}
```

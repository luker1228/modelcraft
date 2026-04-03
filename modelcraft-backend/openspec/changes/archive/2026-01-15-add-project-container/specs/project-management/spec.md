# Spec: Project Management

## Overview

The Project Management capability provides CRUD operations for creating, reading, updating, and deleting projects. Projects serve as top-level containers that organize clusters, models, and enums into logical workspaces.

## ADDED Requirements

### Requirement: Create Project with Human-Readable ID

A project MUST be created with a human-readable ID that serves as its unique identifier throughout the system.

**Acceptance Criteria:**
- Project ID is 2-64 characters long
- Project ID contains only lowercase letters, numbers, and hyphens
- Project ID must start with a lowercase letter
- Project ID is globally unique across all projects
- Title is required and non-empty
- Description is optional

#### Scenario: Create project with valid ID

**Given** no project exists with ID "ecommerce"
**When** user creates a project with:
- id: "ecommerce"
- title: "E-Commerce Platform"
- description: "Customer-facing e-commerce application"

**Then** project is created successfully
**And** project can be retrieved by ID "ecommerce"
**And** project has status "active"
**And** createdAt and updatedAt timestamps are set

#### Scenario: Create project with invalid ID format

**Given** user attempts to create a project
**When** project ID is "E-Commerce" (contains uppercase)
**Then** creation fails with error "invalid project ID format"
**And** error message explains the ID format requirements

#### Scenario: Create project with duplicate ID

**Given** project "ecommerce" already exists
**When** user attempts to create another project with ID "ecommerce"
**Then** creation fails with error "project ID already exists"
**And** existing project remains unchanged

#### Scenario: Create project without title

**Given** user attempts to create a project
**When** title is empty or null
**Then** creation fails with error "project title cannot be empty"

---

### Requirement: Retrieve Project by ID

Users MUST be able to retrieve a project by its unique ID.

**Acceptance Criteria:**
- Project can be retrieved by exact ID match
- Retrieval returns complete project details
- Non-existent project returns appropriate error

#### Scenario: Get existing project

**Given** project "ecommerce" exists with title "E-Commerce Platform"
**When** user retrieves project by ID "ecommerce"
**Then** project details are returned
**And** response includes id, title, description, status, createdAt, updatedAt

#### Scenario: Get non-existent project

**Given** no project exists with ID "nonexistent"
**When** user retrieves project by ID "nonexistent"
**Then** operation returns null or error "project not found"

---

### Requirement: List All Projects

Users MUST be able to list all projects with optional filtering by status.

**Acceptance Criteria:**
- Can list all projects regardless of status
- Can filter by status (active, archived)
- Results are ordered by creation date (newest first)
- Empty list returned when no projects match criteria

#### Scenario: List all active projects

**Given** projects exist:
- "ecommerce" (status: active)
- "crm" (status: active)
- "old-app" (status: archived)

**When** user lists projects with status filter "active"
**Then** response contains 2 projects
**And** response includes "ecommerce" and "crm"
**And** response does not include "old-app"

#### Scenario: List projects when none exist

**Given** no projects exist in the system
**When** user lists all projects
**Then** response is an empty array
**And** no error is returned

---

### Requirement: Update Project Metadata

Users MUST be able to update a project's title and description after creation.

**Acceptance Criteria:**
- Project ID cannot be changed
- Title can be updated if non-empty
- Description can be updated or cleared
- updatedAt timestamp is refreshed on update

#### Scenario: Update project title and description

**Given** project "ecommerce" exists with:
- title: "E-Commerce"
- description: "Old description"

**When** user updates project "ecommerce" with:
- title: "E-Commerce Platform"
- description: "New customer-facing platform"

**Then** project is updated successfully
**And** project title is "E-Commerce Platform"
**And** project description is "New customer-facing platform"
**And** updatedAt timestamp is later than original

#### Scenario: Update project with invalid title

**Given** project "ecommerce" exists
**When** user updates project with empty title
**Then** update fails with error "project title cannot be empty"
**And** existing project remains unchanged

#### Scenario: Update non-existent project

**Given** no project exists with ID "nonexistent"
**When** user attempts to update project "nonexistent"
**Then** operation fails with error "project not found"

---

### Requirement: Delete Project

Users MUST be able to delete a project, with safeguards to prevent accidental data loss.

**Acceptance Criteria:**
- Deleting a project with no resources succeeds immediately
- Deleting a project with resources requires explicit confirmation or fails
- Deletion is permanent (hard delete)
- Cascade deletion removes all dependent resources if confirmed

#### Scenario: Delete empty project

**Given** project "test-project" exists
**And** project has no clusters, models, or enums
**When** user deletes project "test-project"
**Then** project is deleted successfully
**And** project can no longer be retrieved

#### Scenario: Attempt to delete project with resources

**Given** project "ecommerce" exists
**And** project has 2 clusters, 5 models, and 3 enums
**When** user attempts to delete project "ecommerce" without force flag
**Then** deletion fails with error "cannot delete project with existing resources"
**And** error message lists resource counts
**And** project remains unchanged

#### Scenario: Force delete project with resources

**Given** project "old-app" exists with resources
**When** user deletes project "old-app" with force flag
**Then** project is deleted successfully
**And** all clusters, models, and enums are cascade deleted
**And** project can no longer be retrieved

---

## REMOVED Requirements

### Requirement: Default Project for Backward Compatibility

**Reason**: Breaking change approach - no backward compatibility needed
**Migration**: Users must create projects explicitly, no automatic "default" project

---

## ADDED Requirements

### Requirement: Prevent Cross-Project References

The system MUST prevent any references or joins between resources in different projects.

**Acceptance Criteria:**
- Model relations can only reference models within the same project
- Enum field references must be within the same project
- API validation blocks cross-project references before database operations
- Clear error messages explain project isolation constraint

#### Scenario: Attempt cross-project model relation

**Given** projects "ecommerce" and "crm" exist
**And** model "Order" exists in project "ecommerce"
**And** model "Customer" exists in project "crm"
**When** user attempts to create relation from "Order" to "Customer"
**Then** creation fails with error "cannot create relation to model in different project"
**And** error message includes both project IDs

#### Scenario: Attempt cross-project enum reference

**Given** project "ecommerce" has model "Product"
**And** project "crm" has enum "Status"
**When** user attempts to add field to "Product" with enum "Status"
**Then** creation fails with error "enum not found in project"
**And** no field is created

---

### Requirement: Project ID Validation

Project IDs MUST follow a strict format to ensure consistency and prevent issues.

**Acceptance Criteria:**
- Minimum length: 2 characters
- Maximum length: 64 characters
- Allowed characters: lowercase letters (a-z), digits (0-9), hyphens (-)
- Must start with a lowercase letter
- Cannot end with a hyphen

#### Scenario: Valid project IDs

**Given** user creates projects with IDs:
- "app"
- "my-app"
- "app123"
- "e-commerce-platform-v2"

**Then** all project creations succeed

#### Scenario: Invalid project IDs

**Given** user attempts to create projects with IDs:
- "A" (too short)
- "App" (uppercase)
- "my_app" (underscore)
- "123app" (starts with digit)
- "app-" (ends with hyphen)
- "a" * 65 (too long)

**Then** all project creations fail with error "invalid project ID format"

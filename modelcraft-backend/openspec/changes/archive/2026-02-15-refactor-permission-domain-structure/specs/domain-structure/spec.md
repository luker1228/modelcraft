## ADDED Requirements

### Requirement: Consolidated Domain Architecture

The system SHALL organize authentication and authorization concepts into three cohesive domains following Domain-Driven Design principles: Identity Domain (user identity), Auth Domain (authentication mechanisms), and Permission Domain (authorization, roles, permissions, and membership).

#### Scenario: Identity Domain structure

- **GIVEN** the codebase domain structure
- **WHEN** inspecting `internal/domain/user/`
- **THEN** the domain contains User entity representing user identity
- **AND** UserRepository interface for persistence
- **AND** no authorization or membership logic exists in this domain

#### Scenario: Auth Domain structure

- **GIVEN** the codebase domain structure
- **WHEN** inspecting `internal/domain/auth/`
- **THEN** the domain contains AuthProvider interface for pluggable authentication
- **AND** UnifiedClaims for normalized JWT claims
- **AND** ProjectAuthConfig for project-level auth settings
- **AND** no authorization or role management logic exists in this domain

#### Scenario: Permission Domain structure

- **GIVEN** the codebase domain structure
- **WHEN** inspecting `internal/domain/permission/`
- **THEN** the domain contains Role entity for role definitions
- **AND** Permission value object for permission representation
- **AND** UserRole entity for user-organization-role bindings AND membership status
- **AND** RoleRepository, PermissionRepository, UserRoleRepository interfaces
- **AND** no separate `internal/domain/role/` or `internal/domain/membership/` directories exist

#### Scenario: Deprecated domains removed

- **GIVEN** the refactoring is complete
- **WHEN** searching for `internal/domain/role` and `internal/domain/membership` directories
- **THEN** neither directory exists
- **AND** all references have been migrated to `internal/domain/permission`

#### Scenario: Clear domain boundaries

- **GIVEN** a developer needs to implement authorization logic
- **WHEN** they review the domain structure
- **THEN** it is clear that `internal/domain/permission/` is the single source for all authorization concerns
- **AND** the domain handles roles, permissions, user-role bindings, and membership lifecycle
- **AND** there is no ambiguity about where to place role or membership logic

### Requirement: Single Role Entity Definition

The system SHALL use `permission.Role` as the single authoritative Role entity throughout the codebase, eliminating the duplicate `role.Role` definition. All code SHALL import and reference `modelcraft/internal/domain/permission.Role`.

#### Scenario: Permission Role entity as single source

- **GIVEN** the refactoring is complete
- **WHEN** searching for `type Role struct` in `internal/domain/`
- **THEN** exactly one definition exists in `internal/domain/permission/role.go`
- **AND** the definition includes: id (int), name (string), description (string), is_system (bool), org_name (string)
- **AND** no definition exists in `internal/domain/role/` (directory deleted)

#### Scenario: Application services use permission.Role

- **GIVEN** application services in `internal/app/`
- **WHEN** role management operations are performed
- **THEN** all services import `modelcraft/internal/domain/permission`
- **AND** all role variables are declared as `*permission.Role`
- **AND** no imports reference `modelcraft/internal/domain/role`

#### Scenario: Repository interfaces use permission.Role

- **GIVEN** repository interfaces and implementations
- **WHEN** role persistence operations are defined
- **THEN** method signatures use `*permission.Role` type
- **AND** RoleRepository interface is defined in `internal/domain/permission/repository.go`

### Requirement: Unified User-Organization-Role Binding

The system SHALL use `permission.UserRole` as the single entity representing both user-role bindings AND membership status, eliminating the separate `membership.Membership` concept. The entity SHALL use `OrgName` (string) as the organization identifier.

#### Scenario: UserRole replaces Membership

- **GIVEN** the refactoring is complete
- **WHEN** checking for membership-related entities
- **THEN** `internal/domain/membership/` directory does not exist
- **AND** `permission.UserRole` includes membership fields: status, invited_by, invited_at, joined_at
- **AND** all membership operations use UserRole entity

#### Scenario: Organization identifier standardization

- **GIVEN** the UserRole entity
- **WHEN** inspecting the organization reference field
- **THEN** the field is named `OrgName` with type `string`
- **AND** no `OrgID` field of type UUID exists
- **AND** all queries use `org_name` for organization filtering

#### Scenario: Application services use UserRole for membership

- **GIVEN** organization management service in `internal/app/organization/`
- **WHEN** listing organization members
- **THEN** the service uses `permission.UserRoleRepository`
- **AND** returns `[]*permission.UserRole` instead of `[]*membership.Membership`
- **AND** no imports reference `internal/domain/membership`

#### Scenario: Backward compatibility during migration

- **GIVEN** existing user_roles records in database
- **WHEN** the schema migration adds membership columns
- **THEN** existing records receive default values: status='active', joined_at=created_at
- **AND** existing functionality continues to work without data loss

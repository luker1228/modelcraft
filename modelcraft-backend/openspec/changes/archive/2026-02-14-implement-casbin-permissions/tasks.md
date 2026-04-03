## 1. Database Schema

- [x] 1.1 Create `roles` table migration with fields: id, name, description, is_system, org_name, created_at, updated_at
- [x] 1.2 Create `user_roles` table migration with foreign keys to users and roles, cascade delete on role deletion
- [x] 1.3 Create `role_permissions` table migration with role_id FK, org_name, obj, act fields
- [x] 1.4 Add unique constraints: UNIQUE(name, org_name) on roles, UNIQUE(user_id, role_id, org_name) on user_roles
- [x] 1.5 Add indexes: INDEX(org_name, role_id) on role_permissions, INDEX(user_id, org_name) on user_roles
- [x] 1.6 Write SQL migration script in db/schema/mysql/04_permissions.sql
- [x] 1.7 Insert 4 system roles (owner, admin, editor, viewer) with is_system=true in migration
- [x] 1.8 Test migration on local database and verify schema integrity

## 2. Domain Layer - Permission Entities

- [x] 2.1 Create internal/domain/permission/role.go with Role entity (ID, Name, Description, IsSystem, OrgName)
- [x] 2.2 Create internal/domain/permission/permission.go with Permission value object (Obj, Act)
- [x] 2.3 Create internal/domain/permission/user_role.go with UserRole entity (UserID, RoleID, OrgName)
- [x] 2.4 Add validation methods: Role.Validate(), Permission.Validate()
- [x] 2.5 Add business rules: IsSystemRole(), CanModify(), CanDelete()
- [x] 2.6 Write unit tests for domain entities and validation

## 3. Infrastructure - Casbin Integration

- [x] 3.1 Create internal/infrastructure/auth/casbin_model.conf with RBAC model configuration
- [x] 3.2 Create internal/infrastructure/auth/system_roles.go with hardcoded system role permissions
- [x] 3.3 Create internal/infrastructure/auth/casbin_enforcer.go with singleton enforcer initialization
- [x] 3.4 Implement LoadSystemRolePermissions() to load hardcoded permissions into enforcer
- [x] 3.5 Implement enforcer caching with EnableCache(true)
- [x] 3.6 Add error handling and logging for enforcer initialization
- [x] 3.7 Write unit tests for enforcer with various permission scenarios
- [x] 3.8 Add integration tests for enforcer with database-backed custom roles

## 4. Infrastructure - Repository Layer

- [x] 4.1 Create internal/domain/permission/repository.go with RoleRepository interface
- [x] 4.2 Implement internal/infrastructure/repository/role_repository.go with sqlc
- [x] 4.3 Add methods: CreateRole(), GetRoleByID(), GetRoleByNameAndOrg(), ListRolesByOrg(), UpdateRole(), DeleteRole()
- [x] 4.4 Create PermissionRepository interface with methods: AddPermission(), RemovePermission(), ListPermissionsByRole()
- [x] 4.5 Implement permission_repository.go with sqlc
- [x] 4.6 Create UserRoleRepository interface with methods: AssignRole(), RevokeRole(), ListUserRoles(), ListRoleUsers()
- [x] 4.7 Implement user_role_repository.go with sqlc
- [x] 4.8 Add query optimization: preload relations, use indexes
- [x] 4.9 Write unit tests for each repository with mock database

## 5. Application Layer - Permission Services

- [x] 5.1 Create internal/app/permission/role_service.go with RoleService
- [x] 5.2 Implement CreateCustomRole(name, description, orgName) with validation (prevent system role names)
- [x] 5.3 Implement UpdateRole(roleID, input) with system role protection
- [x] 5.4 Implement DeleteRole(roleID) with system role protection and cascade cleanup
- [x] 5.5 Implement GetRole(roleID), ListRoles(orgName, includeSystem bool)
- [x] 5.6 Create internal/app/permission/permission_service.go with PermissionService
- [x] 5.7 Implement AddPermissionToRole(roleID, obj, act) with validation
- [x] 5.8 Implement RemovePermissionFromRole(roleID, obj, act)
- [x] 5.9 Implement ListRolePermissions(roleID) - merge hardcoded + database permissions
- [x] 5.10 Create internal/app/permission/user_role_service.go with UserRoleService
- [x] 5.11 Implement AssignRoleToUser(userID, roleID, orgName) with validation
- [x] 5.12 Implement RevokeRoleFromUser(userID, roleID, orgName)
- [x] 5.13 Implement ListUserRoles(userID, orgName), ListRoleUsers(roleID, orgName)
- [x] 5.14 Implement CheckPermission(userID, orgName, obj, act) - core permission check logic
- [x] 5.15 Write unit tests for all service methods with mock repositories
- [x] 5.16 Add integration tests with real database

## 6. GraphQL Directive Implementation

- [x] 6.1 Add directive definition to api/graph/schema/base.graphql: `directive @hasPermission(action: String!) on FIELD_DEFINITION`
- [x] 6.2 Create internal/interfaces/graphql/directives.go with HasPermission() function
- [x] 6.3 Implement user context extraction using GetUserIDFromContext() and GetOrgNameFromContext()
- [x] 6.4 Implement Casbin enforce call: enforcer.Enforce(userID, obj, act)
- [x] 6.5 Add error handling: return permission denied error with required action
- [x] 6.6 Update gqlgen.yml to register @hasPermission directive
- [x] 6.7 Run `task generate-gql` to generate directive code
- [x] 6.8 Add @hasPermission directive to project GraphQL operations (createProject, updateProject, deleteProject)
- [x] 6.9 Add @hasPermission directive to model GraphQL operations (createModel, updateModel, deleteModel, deployModel)
- [x] 6.10 Add @hasPermission directive to cluster GraphQL operations (createCluster, updateCluster, deleteCluster)
- [x] 6.11 Add @hasPermission directive to enum GraphQL operations (createEnum, updateEnum, deleteEnum)
- [x] 6.12 Write integration tests for directive with different roles and permissions
- [x] 6.13 Test directive with missing authentication context (should return error)

## 7. GraphQL Permission Management APIs

- [x] 7.1 Create api/graph/schema/permission.graphql with Role, Permission, UserRole types
- [x] 7.2 Add Query types: roles(orgName: String!), role(id: ID!), userRoles(userID: String!, orgName: String!)
- [x] 7.3 Add Mutation types: createRole, updateRole, deleteRole (with typed error responses)
- [x] 7.4 Add Mutation types: addPermissionToRole, removePermissionFromRole
- [x] 7.5 Add Mutation types: assignRoleToUser, revokeRoleFromUser
- [x] 7.6 Add error types: RoleAlreadyExists, RoleNotFound, SystemRoleCannotBeModified, InvalidPermissionInput
- [x] 7.7 Run `task generate-gql` to generate resolver stubs
- [x] 7.8 Implement internal/interfaces/graphql/permission.resolvers.go with resolver logic
- [x] 7.9 Add validation: prevent system role modification/deletion in resolvers
- [x] 7.10 Add permission checks: only admin/owner can manage roles (use @hasPermission directive)
- [x] 7.11 Write integration tests for role CRUD operations
- [x] 7.12 Write integration tests for permission assignment operations
- [x] 7.13 Write integration tests for user role assignment operations

## 8. Integration with JWT Middleware

- [x] 8.1 Review internal/middleware/jwt_auth.go to confirm user_id and org_name are stored in context
- [x] 8.2 Create internal/infrastructure/auth/permission_checker.go to centralize permission check logic
- [x] 8.3 Implement CheckUserPermission(ctx, userID, orgName, action string) helper function
- [x] 8.4 Query user's roles from user_roles table filtered by org_name
- [x] 8.5 For each role, load permissions: hardcoded if system role, else query role_permissions table
- [x] 8.6 Build Casbin policy rules from aggregated permissions
- [x] 8.7 Call enforcer.Enforce() with parsed obj and act from action string
- [x] 8.8 Cache permission check results per request to avoid redundant queries
- [x] 8.9 Add logging for permission check decisions (allow/deny with reason)
- [x] 8.10 Write unit tests for permission checker with various role combinations

## 9. Documentation

- [x] 9.1 Create docs/04-auth/casbin-permissions.md with architecture overview
- [x] 9.2 Document system roles and their default permissions
- [x] 9.3 Document custom role creation workflow with GraphQL examples
- [x] 9.4 Document user role assignment workflow with GraphQL examples
- [x] 9.5 Document permission format (resource:action) and wildcard support
- [x] 9.6 Document @hasPermission directive usage with schema examples
- [x] 9.7 Update CLAUDE.md with permission system overview and key patterns
- [x] 9.8 Create migration guide from JWT permissions to Casbin roles
- [x] 9.9 Add troubleshooting section for common permission issues

## 10. Testing

- [x] 10.1 Write pytest integration tests in tests/automated/test_permissions.py
- [x] 10.2 Test system role permissions: owner has wildcard, admin has specific permissions
- [x] 10.3 Test custom role creation, update, deletion with API
- [x] 10.4 Test permission assignment to custom roles
- [x] 10.5 Test user role assignment across different tenants (same user, different roles)
- [x] 10.6 Test GraphQL directive enforcement: allow access with correct role, deny without
- [x] 10.7 Test permission denied error responses (check error message format)
- [x] 10.8 Test system role protection: cannot modify/delete system roles
- [x] 10.9 Test cascade deletion: delete role, verify user_roles and role_permissions cleaned up
- [x] 10.10 Test multi-tenant isolation: user in org1 cannot access org2 resources
- [x] 10.11 Test permission union: user with multiple roles gets combined permissions
- [x] 10.12 Performance test: measure latency of permission checks under load
- [x] 10.13 Add test coverage report and ensure >80% coverage for new code

## 11. Deployment & Rollout

- [ ] 11.1 Deploy database migrations to staging environment
- [ ] 11.2 Initialize system roles via migration script
- [ ] 11.3 Deploy application code with Casbin integration (directive not yet enabled)
- [ ] 11.4 Test permission APIs in staging with manual GraphQL queries
- [ ] 11.5 Enable @hasPermission directive on pilot operations (e.g., project mutations)
- [ ] 11.6 Monitor logs for permission check errors and performance
- [ ] 11.7 Gradually roll out directive to all operations over 2-3 days
- [ ] 11.8 Provide migration tool to convert existing JWT permissions to Casbin roles
- [ ] 11.9 Update production deployment documentation
- [ ] 11.10 Train support team on new permission system and troubleshooting
- [ ] 11.11 Deploy to production with feature flag (disabled by default)
- [ ] 11.12 Enable feature flag for selected tenants (pilot phase)
- [ ] 11.13 Monitor production metrics (latency, error rate, permission denials)
- [ ] 11.14 Enable feature flag globally after validation
- [ ] 11.15 Optionally deprecate old JWT permission checks after 2-4 weeks

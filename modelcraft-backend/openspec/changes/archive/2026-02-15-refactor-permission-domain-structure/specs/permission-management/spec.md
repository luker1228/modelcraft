## MODIFIED Requirements

### Requirement: User Role Assignment Data Model

The system SHALL provide a unified data model for binding users to roles AND managing membership lifecycle with fields: id, user_id, role_id, org_name, status, invited_by, invited_at, joined_at, created_at, updated_at. This replaces the previous separate `Membership` concept.

#### Scenario: Extended user_roles table schema

- **GIVEN** the database schema is migrated
- **WHEN** the user_roles table structure is inspected
- **THEN** the table has all previous columns: id (PK), user_id (VARCHAR, FK to users.id), role_id (INT, FK to roles.id), org_name (VARCHAR), created_at (TIMESTAMP)
- **AND** new membership columns exist: status (VARCHAR DEFAULT 'active'), invited_by (VARCHAR nullable), invited_at (TIMESTAMP nullable), joined_at (TIMESTAMP nullable), updated_at (TIMESTAMP)
- **AND** a unique constraint exists on (user_id, role_id, org_name)
- **AND** foreign keys cascade on delete
- **AND** indexes exist on (status), (org_name, status)

#### Scenario: Active member record

- **GIVEN** user "user123" is directly assigned role "admin" (role_id=2) in organization "org1"
- **WHEN** the assignment is stored
- **THEN** a record is created with user_id="user123", role_id=2, org_name="org1", status="active"
- **AND** joined_at is set to the current timestamp
- **AND** invited_by, invited_at are NULL

#### Scenario: Invited member record

- **GIVEN** user "user456" is invited to organization "org1" with role "editor" by user "admin123"
- **WHEN** the invitation is created
- **THEN** a record is created with user_id="user456", role_id=3, org_name="org1", status="invited"
- **AND** invited_by="admin123"
- **AND** invited_at is set to the current timestamp
- **AND** joined_at is NULL

#### Scenario: Invitation acceptance

- **GIVEN** an invited user record exists with status="invited"
- **WHEN** the user accepts the invitation
- **THEN** status changes to "active"
- **AND** joined_at is set to the current timestamp
- **AND** invited_by and invited_at remain unchanged (audit trail)
- **AND** updated_at is updated

#### Scenario: Member suspension

- **GIVEN** an active user-role binding exists
- **WHEN** an admin suspends the member
- **THEN** status changes to "suspended"
- **AND** updated_at is updated
- **AND** the user loses access permissions immediately

#### Scenario: Member role change

- **GIVEN** an active user-role binding exists with role_id=3 (editor)
- **WHEN** an admin changes the role to viewer (role_id=4)
- **THEN** role_id is updated to 4
- **AND** status remains "active"
- **AND** updated_at is updated
- **AND** permissions change immediately

#### Scenario: List organization members

- **GIVEN** organization "org1" has 5 user-role bindings with various statuses
- **WHEN** querying for all members of "org1"
- **THEN** all 5 records are returned
- **AND** records include status, invited_at, joined_at fields
- **AND** results can be filtered by status (e.g., only "active" members)

#### Scenario: Cascade deletion when role is deleted

- **GIVEN** role_id=5 has 3 user-role bindings
- **WHEN** role_id=5 is deleted
- **THEN** all 3 user-role bindings are automatically deleted via CASCADE
- **AND** no orphaned records remain

## ADDED Requirements

### Requirement: UserRole Entity Membership Methods

The UserRole entity SHALL provide methods for managing membership lifecycle: AcceptInvitation(), ChangeRole(), Suspend(), Activate(), and IsActive(). These methods enforce business rules and state transitions.

#### Scenario: Accept invitation method

- **GIVEN** a UserRole entity with status="invited"
- **WHEN** AcceptInvitation() is called
- **THEN** status changes to "active"
- **AND** joined_at is set to current time
- **AND** method returns nil error

#### Scenario: Accept invitation on non-invited status fails

- **GIVEN** a UserRole entity with status="active"
- **WHEN** AcceptInvitation() is called
- **THEN** method returns error "can only accept invitation when status is 'invited'"
- **AND** status remains unchanged

#### Scenario: Change role method

- **GIVEN** a UserRole entity with role_id=2
- **WHEN** ChangeRole(4) is called
- **THEN** role_id changes to 4
- **AND** updated_at is set to current time
- **AND** method returns nil error

#### Scenario: Change role with invalid id fails

- **GIVEN** a UserRole entity
- **WHEN** ChangeRole(0) is called with invalid role_id
- **THEN** method returns error "role_id must be positive"
- **AND** role_id remains unchanged

#### Scenario: Suspend method

- **GIVEN** a UserRole entity with status="active"
- **WHEN** Suspend() is called
- **THEN** status changes to "suspended"
- **AND** updated_at is updated

#### Scenario: Activate method

- **GIVEN** a UserRole entity with status="suspended"
- **WHEN** Activate() is called
- **THEN** status changes to "active"
- **AND** updated_at is updated

#### Scenario: IsActive check

- **GIVEN** a UserRole entity with status="active"
- **WHEN** IsActive() is called
- **THEN** method returns true

#### Scenario: IsActive check on suspended

- **GIVEN** a UserRole entity with status="suspended"
- **WHEN** IsActive() is called
- **THEN** method returns false

### Requirement: Extended UserRoleRepository Interface

The UserRoleRepository interface SHALL include methods for membership management operations in addition to role assignment: CreateInvitation, AcceptInvitation, UpdateMembershipStatus, ListOrgMembers, ListUserMemberships, DeleteMembership, DeleteUserRolesByOrg.

#### Scenario: CreateInvitation repository method

- **GIVEN** a UserRole entity with status="invited"
- **WHEN** CreateInvitation(ctx, invitation) is called
- **THEN** the invitation is persisted to the database
- **AND** invited_by, invited_at fields are stored
- **AND** method returns nil error

#### Scenario: AcceptInvitation repository method

- **GIVEN** an invited user-role record exists in database
- **WHEN** AcceptInvitation(ctx, userID, orgName) is called
- **THEN** the record's status is updated to "active"
- **AND** joined_at is set to current timestamp
- **AND** method returns nil error

#### Scenario: UpdateMembershipStatus repository method

- **GIVEN** an active user-role record exists
- **WHEN** UpdateMembershipStatus(ctx, userID, orgName, "suspended") is called
- **THEN** the record's status is updated to "suspended"
- **AND** updated_at is updated
- **AND** method returns nil error

#### Scenario: ListOrgMembers repository method

- **GIVEN** organization "org1" has 5 members with different statuses
- **WHEN** ListOrgMembers(ctx, "org1") is called
- **THEN** all 5 UserRole records are returned
- **AND** records include all membership fields (status, invited_at, joined_at)

#### Scenario: ListUserMemberships repository method

- **GIVEN** user "user123" has memberships in 3 organizations
- **WHEN** ListUserMemberships(ctx, "user123") is called
- **THEN** all 3 UserRole records are returned
- **AND** records span different org_name values

#### Scenario: DeleteMembership repository method

- **GIVEN** a user-role record with id=10 exists
- **WHEN** DeleteMembership(ctx, 10) is called
- **THEN** the record is deleted from the database
- **AND** method returns nil error

#### Scenario: DeleteUserRolesByOrg repository method

- **GIVEN** organization "org1" has 5 user-role bindings
- **WHEN** DeleteUserRolesByOrg(ctx, "org1") is called
- **THEN** all 5 records are deleted
- **AND** method returns nil error
- **AND** this supports organization deletion cascade

## REMOVED Requirements

None. This change extends existing requirements without removing functionality.

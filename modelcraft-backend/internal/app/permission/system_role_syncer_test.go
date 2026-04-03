package permission

import (
	"errors"
	"modelcraft/internal/domain/permission"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// systemRoleSyncerTestSetup creates a syncer with mock repos for tests.
func systemRoleSyncerTestSetup() (
	*SystemRolePermissionsSyncer,
	*MockRoleRepository,
	*MockPermissionRepository,
) {
	mockRoleRepo := new(MockRoleRepository)
	mockPermRepo := new(MockPermissionRepository)
	syncer := NewSystemRolePermissionsSyncer(mockRoleRepo, mockPermRepo)
	return syncer, mockRoleRepo, mockPermRepo
}

// stubSystemRole returns a fake DB role record for the given system role name.
func stubSystemRole(id int, name string) *permission.Role {
	return &permission.Role{
		ID:       id,
		Name:     name,
		IsSystem: true,
		OrgName:  permission.SystemOrgName,
	}
}

// TestSystemRolePermissionsSyncer_Sync_DeleteThenInsert verifies that Sync calls
// DeletePermissionsByRole followed by AddPermission for each system role permission.
func TestSystemRolePermissionsSyncer_Sync_DeleteThenInsert(t *testing.T) {
	syncer, mockRoleRepo, mockPermRepo := systemRoleSyncerTestSetup()
	ctx := testContext()

	// Stub role lookups for all four system roles.
	roles := map[string]*permission.Role{
		permission.RoleOwner:  stubSystemRole(1, permission.RoleOwner),
		permission.RoleAdmin:  stubSystemRole(2, permission.RoleAdmin),
		permission.RoleEditor: stubSystemRole(3, permission.RoleEditor),
		permission.RoleViewer: stubSystemRole(4, permission.RoleViewer),
	}
	for name, role := range roles {
		mockRoleRepo.On(
			"GetRoleByNameAndOrg",
			ctx, name, permission.SystemOrgName,
		).Return(role, nil)
	}

	// Expect delete for every system role ID.
	for _, role := range roles {
		mockPermRepo.On("DeletePermissionsByRole", ctx, role.ID).Return(nil)
	}

	// Expect AddPermission for every permission of every system role.
	mockPermRepo.On(
		"AddPermission",
		ctx, mock.AnythingOfType("int"), permission.SystemOrgName, mock.AnythingOfType("*permission.Permission"),
	).Return(nil)

	err := syncer.Sync(ctx)

	assert.NoError(t, err)
	mockRoleRepo.AssertExpectations(t)
	mockPermRepo.AssertExpectations(t)
}

// TestSystemRolePermissionsSyncer_Sync_Idempotent verifies that calling Sync twice
// does not return an error and continues to succeed.
func TestSystemRolePermissionsSyncer_Sync_Idempotent(t *testing.T) {
	syncer, mockRoleRepo, mockPermRepo := systemRoleSyncerTestSetup()
	ctx := testContext()

	roles := map[string]*permission.Role{
		permission.RoleOwner:  stubSystemRole(1, permission.RoleOwner),
		permission.RoleAdmin:  stubSystemRole(2, permission.RoleAdmin),
		permission.RoleEditor: stubSystemRole(3, permission.RoleEditor),
		permission.RoleViewer: stubSystemRole(4, permission.RoleViewer),
	}
	for name, role := range roles {
		mockRoleRepo.On(
			"GetRoleByNameAndOrg",
			ctx, name, permission.SystemOrgName,
		).Return(role, nil)
	}
	for _, role := range roles {
		mockPermRepo.On("DeletePermissionsByRole", ctx, role.ID).Return(nil)
	}
	mockPermRepo.On(
		"AddPermission",
		ctx, mock.AnythingOfType("int"), permission.SystemOrgName, mock.AnythingOfType("*permission.Permission"),
	).Return(nil)

	// First call.
	err := syncer.Sync(ctx)
	assert.NoError(t, err)

	// Second call — mocks accept unlimited calls so this should also succeed.
	err = syncer.Sync(ctx)
	assert.NoError(t, err)
}

// TestSystemRolePermissionsSyncer_Sync_CreatesRoleWhenMissing verifies that Sync
// automatically creates system role records that are missing from the database.
func TestSystemRolePermissionsSyncer_Sync_CreatesRoleWhenMissing(t *testing.T) {
	syncer, mockRoleRepo, mockPermRepo := systemRoleSyncerTestSetup()
	ctx := testContext()

	// All four roles return nil (not yet seeded in DB).
	mockRoleRepo.On(
		"GetRoleByNameAndOrg",
		ctx, mock.AnythingOfType("string"), permission.SystemOrgName,
	).Return(nil, nil)

	// Expect CreateRole to be called for each missing system role.
	mockRoleRepo.On(
		"CreateRole",
		ctx, mock.MatchedBy(func(r *permission.Role) bool {
			return r.IsSystem && r.OrgName == permission.SystemOrgName
		}),
	).Return(nil)

	mockPermRepo.On("DeletePermissionsByRole", ctx, mock.AnythingOfType("int")).Return(nil)
	mockPermRepo.On(
		"AddPermission",
		ctx, mock.AnythingOfType("int"), permission.SystemOrgName, mock.AnythingOfType("*permission.Permission"),
	).Return(nil)

	err := syncer.Sync(ctx)

	assert.NoError(t, err)
	mockRoleRepo.AssertExpectations(t)
	mockPermRepo.AssertExpectations(t)
}

// TestSystemRolePermissionsSyncer_Sync_CreateRoleError verifies that Sync returns an
// error when CreateRole fails for a missing system role.
func TestSystemRolePermissionsSyncer_Sync_CreateRoleError(t *testing.T) {
	syncer, mockRoleRepo, mockPermRepo := systemRoleSyncerTestSetup()
	ctx := testContext()

	mockRoleRepo.On(
		"GetRoleByNameAndOrg",
		ctx, mock.AnythingOfType("string"), permission.SystemOrgName,
	).Return(nil, nil).Once()

	createErr := errors.New("db insert failed")
	mockRoleRepo.On(
		"CreateRole",
		ctx, mock.AnythingOfType("*permission.Role"),
	).Return(createErr)

	err := syncer.Sync(ctx)

	assert.Error(t, err)
	mockPermRepo.AssertNotCalled(t, "DeletePermissionsByRole")
	mockPermRepo.AssertNotCalled(t, "AddPermission")
}

// TestSystemRolePermissionsSyncer_Sync_DeleteError verifies that Sync returns an
// error when DeletePermissionsByRole fails, without proceeding to insert.
func TestSystemRolePermissionsSyncer_Sync_DeleteError(t *testing.T) {
	syncer, mockRoleRepo, mockPermRepo := systemRoleSyncerTestSetup()
	ctx := testContext()

	mockRoleRepo.On(
		"GetRoleByNameAndOrg",
		ctx, mock.AnythingOfType("string"), permission.SystemOrgName,
	).Return(stubSystemRole(1, permission.RoleOwner), nil).Once()

	deleteErr := errors.New("db error")
	mockPermRepo.On("DeletePermissionsByRole", ctx, 1).Return(deleteErr)

	err := syncer.Sync(ctx)

	assert.Error(t, err)
	mockPermRepo.AssertNotCalled(t, "AddPermission")
}

// TestSystemRolePermissionsSyncer_Sync_InsertError verifies that Sync returns an
// error when AddPermission fails.
func TestSystemRolePermissionsSyncer_Sync_InsertError(t *testing.T) {
	syncer, mockRoleRepo, mockPermRepo := systemRoleSyncerTestSetup()
	ctx := testContext()

	mockRoleRepo.On(
		"GetRoleByNameAndOrg",
		ctx, mock.AnythingOfType("string"), permission.SystemOrgName,
	).Return(stubSystemRole(1, permission.RoleOwner), nil).Once()

	mockPermRepo.On("DeletePermissionsByRole", ctx, 1).Return(nil)

	insertErr := errors.New("insert failed")
	mockPermRepo.On(
		"AddPermission",
		ctx, 1, permission.SystemOrgName, mock.AnythingOfType("*permission.Permission"),
	).Return(insertErr)

	err := syncer.Sync(ctx)

	assert.Error(t, err)
}

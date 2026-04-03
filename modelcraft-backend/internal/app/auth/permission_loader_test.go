package auth

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	domainPerm "modelcraft/internal/domain/permission"
)

// ========== Mocks ==========

type MockUserRoleRepository struct {
	mock.Mock
}

func (m *MockUserRoleRepository) AssignRole(ctx context.Context, userRole *domainPerm.UserRole) error {
	args := m.Called(ctx, userRole)
	return args.Error(0)
}

func (m *MockUserRoleRepository) RevokeRole(ctx context.Context, userID string, roleID int, orgName string) error {
	args := m.Called(ctx, userID, roleID, orgName)
	return args.Error(0)
}

func (m *MockUserRoleRepository) ListUserRoles(
	ctx context.Context,
	userID string,
	orgName string,
) ([]*domainPerm.UserRole, error) {
	args := m.Called(ctx, userID, orgName)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*domainPerm.UserRole), args.Error(1)
}

func (m *MockUserRoleRepository) ListRoleUsers(
	ctx context.Context,
	roleID int,
	orgName string,
) ([]*domainPerm.UserRole, error) {
	args := m.Called(ctx, roleID, orgName)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*domainPerm.UserRole), args.Error(1)
}

func (m *MockUserRoleRepository) GetUserRole(
	ctx context.Context,
	userID string,
	roleID int,
	orgName string,
) (*domainPerm.UserRole, error) {
	args := m.Called(ctx, userID, roleID, orgName)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domainPerm.UserRole), args.Error(1)
}

func (m *MockUserRoleRepository) DeleteUserRolesByRole(ctx context.Context, roleID int) error {
	args := m.Called(ctx, roleID)
	return args.Error(0)
}

type MockRoleRepository struct {
	mock.Mock
}

func (m *MockRoleRepository) CreateRole(ctx context.Context, role *domainPerm.Role) error {
	args := m.Called(ctx, role)
	return args.Error(0)
}

func (m *MockRoleRepository) GetRoleByID(ctx context.Context, id int) (*domainPerm.Role, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domainPerm.Role), args.Error(1)
}

func (m *MockRoleRepository) GetRoleByNameAndOrg(
	ctx context.Context,
	name string,
	orgName string,
) (*domainPerm.Role, error) {
	args := m.Called(ctx, name, orgName)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*domainPerm.Role), args.Error(1)
}

func (m *MockRoleRepository) ListRolesByOrg(
	ctx context.Context,
	orgName string,
	includeSystem bool,
) ([]*domainPerm.Role, error) {
	args := m.Called(ctx, orgName, includeSystem)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*domainPerm.Role), args.Error(1)
}

func (m *MockRoleRepository) UpdateRole(ctx context.Context, role *domainPerm.Role) error {
	args := m.Called(ctx, role)
	return args.Error(0)
}

func (m *MockRoleRepository) DeleteRole(ctx context.Context, id int) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

type MockPermissionRepository struct {
	mock.Mock
}

func (m *MockPermissionRepository) AddPermission(
	ctx context.Context,
	roleID int,
	orgName string,
	perm *domainPerm.Permission,
) error {
	args := m.Called(ctx, roleID, orgName, perm)
	return args.Error(0)
}

func (m *MockPermissionRepository) RemovePermission(ctx context.Context, roleID int, obj, act string) error {
	args := m.Called(ctx, roleID, obj, act)
	return args.Error(0)
}

func (m *MockPermissionRepository) ListPermissionsByRole(
	ctx context.Context,
	roleID int,
) ([]*domainPerm.Permission, error) {
	args := m.Called(ctx, roleID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*domainPerm.Permission), args.Error(1)
}

func (m *MockPermissionRepository) ListPermissionsByRoleAndOrg(
	ctx context.Context,
	roleID int,
	orgName string,
) ([]*domainPerm.Permission, error) {
	args := m.Called(ctx, roleID, orgName)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).([]*domainPerm.Permission), args.Error(1)
}

func (m *MockPermissionRepository) DeletePermissionsByRole(ctx context.Context, roleID int) error {
	args := m.Called(ctx, roleID)
	return args.Error(0)
}

// ========== Test Helper Functions ==========

func createTestPermissionLoader(
	userRoleRepo domainPerm.UserRoleRepository,
	roleRepo domainPerm.RoleRepository,
	permRepo domainPerm.PermissionRepository,
) *PermissionLoader {
	return NewPermissionLoader(userRoleRepo, roleRepo, permRepo)
}

// ========== Permission Loader Tests ==========

func TestPermissionLoader_LoadUserPermissions_Success(t *testing.T) {
	// Arrange
	ctx := context.Background()
	userID := "user-123"
	orgName := "modelcraft"

	mockUserRoleRepo := new(MockUserRoleRepository)
	mockRoleRepo := new(MockRoleRepository)
	mockPermRepo := new(MockPermissionRepository)

	// User has one custom role
	userRoles := []*domainPerm.UserRole{
		{
			ID:        1,
			UserID:    userID,
			RoleID:    10,
			OrgName:   orgName,
			CreatedAt: time.Now(),
		},
	}

	// Role details
	role := &domainPerm.Role{
		ID:       10,
		Name:     "editor",
		OrgName:  orgName,
		IsSystem: false, // Custom role
	}

	// Role permissions
	permissions := []*domainPerm.Permission{
		domainPerm.NewPermission("model", "read"),
		domainPerm.NewPermission("model", "write"),
		domainPerm.NewPermission("project", "read"),
	}

	mockUserRoleRepo.On("ListUserRoles", ctx, userID, orgName).Return(userRoles, nil)
	mockRoleRepo.On("GetRoleByID", ctx, 10).Return(role, nil)
	mockPermRepo.On("ListPermissionsByRole", ctx, 10).Return(permissions, nil)

	loader := createTestPermissionLoader(mockUserRoleRepo, mockRoleRepo, mockPermRepo)

	// Act
	result, err := loader.LoadUserPermissions(ctx, userID, orgName)

	// Assert
	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Len(t, result, 3)
	assert.Contains(t, result, "model:read")
	assert.Contains(t, result, "model:write")
	assert.Contains(t, result, "project:read")

	mockUserRoleRepo.AssertExpectations(t)
	mockRoleRepo.AssertExpectations(t)
	mockPermRepo.AssertExpectations(t)
}

func TestPermissionLoader_LoadUserPermissions_NoRoles(t *testing.T) {
	// Arrange
	ctx := context.Background()
	userID := "user-123"
	orgName := "modelcraft"

	mockUserRoleRepo := new(MockUserRoleRepository)
	mockRoleRepo := new(MockRoleRepository)
	mockPermRepo := new(MockPermissionRepository)

	// User has no roles
	mockUserRoleRepo.On("ListUserRoles", ctx, userID, orgName).Return([]*domainPerm.UserRole{}, nil)

	loader := createTestPermissionLoader(mockUserRoleRepo, mockRoleRepo, mockPermRepo)

	// Act
	result, err := loader.LoadUserPermissions(ctx, userID, orgName)

	// Assert
	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Empty(t, result)

	mockUserRoleRepo.AssertExpectations(t)
}

func TestPermissionLoader_LoadUserPermissions_DatabaseError(t *testing.T) {
	// Arrange
	ctx := context.Background()
	userID := "user-123"
	orgName := "modelcraft"

	mockUserRoleRepo := new(MockUserRoleRepository)
	mockRoleRepo := new(MockRoleRepository)
	mockPermRepo := new(MockPermissionRepository)

	// Simulate database error
	dbError := errors.New("database connection failed")
	mockUserRoleRepo.On("ListUserRoles", ctx, userID, orgName).Return(nil, dbError)

	loader := createTestPermissionLoader(mockUserRoleRepo, mockRoleRepo, mockPermRepo)

	// Act
	result, err := loader.LoadUserPermissions(ctx, userID, orgName)

	// Assert
	assert.Error(t, err)
	assert.Nil(t, result)
	assert.Contains(t, err.Error(), "failed to query user roles")

	mockUserRoleRepo.AssertExpectations(t)
}

func TestPermissionLoader_LoadUserPermissions_MultipleRoles(t *testing.T) {
	// Arrange
	ctx := context.Background()
	userID := "user-123"
	orgName := "modelcraft"

	mockUserRoleRepo := new(MockUserRoleRepository)
	mockRoleRepo := new(MockRoleRepository)
	mockPermRepo := new(MockPermissionRepository)

	// User has two roles
	userRoles := []*domainPerm.UserRole{
		{ID: 1, UserID: userID, RoleID: 10, OrgName: orgName},
		{ID: 2, UserID: userID, RoleID: 11, OrgName: orgName},
	}

	// Role 1 (editor)
	role1 := &domainPerm.Role{ID: 10, Name: "editor", OrgName: orgName, IsSystem: false}
	perms1 := []*domainPerm.Permission{
		domainPerm.NewPermission("model", "read"),
		domainPerm.NewPermission("model", "write"),
	}

	// Role 2 (viewer)
	role2 := &domainPerm.Role{ID: 11, Name: "viewer", OrgName: orgName, IsSystem: false}
	perms2 := []*domainPerm.Permission{
		domainPerm.NewPermission("model", "read"), // Duplicate - should be deduplicated
		domainPerm.NewPermission("project", "read"),
	}

	mockUserRoleRepo.On("ListUserRoles", ctx, userID, orgName).Return(userRoles, nil)
	mockRoleRepo.On("GetRoleByID", ctx, 10).Return(role1, nil)
	mockRoleRepo.On("GetRoleByID", ctx, 11).Return(role2, nil)
	mockPermRepo.On("ListPermissionsByRole", ctx, 10).Return(perms1, nil)
	mockPermRepo.On("ListPermissionsByRole", ctx, 11).Return(perms2, nil)

	loader := createTestPermissionLoader(mockUserRoleRepo, mockRoleRepo, mockPermRepo)

	// Act
	result, err := loader.LoadUserPermissions(ctx, userID, orgName)

	// Assert
	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Len(t, result, 3) // Deduplicated: model:read, model:write, project:read
	assert.Contains(t, result, "model:read")
	assert.Contains(t, result, "model:write")
	assert.Contains(t, result, "project:read")

	mockUserRoleRepo.AssertExpectations(t)
	mockRoleRepo.AssertExpectations(t)
	mockPermRepo.AssertExpectations(t)
}

func TestPermissionLoader_LoadUserPermissions_RoleNotFound(t *testing.T) {
	// Arrange
	ctx := context.Background()
	userID := "user-123"
	orgName := "modelcraft"

	mockUserRoleRepo := new(MockUserRoleRepository)
	mockRoleRepo := new(MockRoleRepository)
	mockPermRepo := new(MockPermissionRepository)

	// User has a role, but role doesn't exist in database
	userRoles := []*domainPerm.UserRole{
		{ID: 1, UserID: userID, RoleID: 99, OrgName: orgName},
	}

	mockUserRoleRepo.On("ListUserRoles", ctx, userID, orgName).Return(userRoles, nil)
	mockRoleRepo.On("GetRoleByID", ctx, 99).Return(nil, nil) // Role not found

	loader := createTestPermissionLoader(mockUserRoleRepo, mockRoleRepo, mockPermRepo)

	// Act
	result, err := loader.LoadUserPermissions(ctx, userID, orgName)

	// Assert
	// Should not error, but should skip missing role
	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Empty(t, result) // No permissions since role not found

	mockUserRoleRepo.AssertExpectations(t)
	mockRoleRepo.AssertExpectations(t)
}

func TestPermissionLoader_LoadUserPermissions_SystemRole(t *testing.T) {
	// Arrange
	ctx := context.Background()
	userID := "user-123"
	orgName := "modelcraft"

	mockUserRoleRepo := new(MockUserRoleRepository)
	mockRoleRepo := new(MockRoleRepository)
	mockPermRepo := new(MockPermissionRepository)

	// User has owner role (system role)
	userRoles := []*domainPerm.UserRole{
		{ID: 1, UserID: userID, RoleID: 1, OrgName: orgName},
	}

	// Owner is a system role
	role := &domainPerm.Role{
		ID:       1,
		Name:     "owner",
		OrgName:  "",   // System roles have empty org
		IsSystem: true, // System role
	}

	mockUserRoleRepo.On("ListUserRoles", ctx, userID, orgName).Return(userRoles, nil)
	mockRoleRepo.On("GetRoleByID", ctx, 1).Return(role, nil)
	// No permRepo call for system roles - permissions are hardcoded

	loader := createTestPermissionLoader(mockUserRoleRepo, mockRoleRepo, mockPermRepo)

	// Act
	result, err := loader.LoadUserPermissions(ctx, userID, orgName)

	// Assert
	assert.NoError(t, err)
	require.NotNil(t, result)
	// Owner should have wildcard permission "*:*"
	assert.Greater(t, len(result), 0, "System role should have permissions")

	mockUserRoleRepo.AssertExpectations(t)
	mockRoleRepo.AssertExpectations(t)
	// permRepo should NOT be called for system roles
	mockPermRepo.AssertNotCalled(t, "ListPermissionsByRole", mock.Anything, mock.Anything)
}

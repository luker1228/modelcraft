package permission

import (
	"context"
	"modelcraft/internal/domain/permission"
	"modelcraft/pkg/bizerrors"
	"modelcraft/pkg/ctxutils"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// MockRoleRepository is a mock implementation of permission.RoleRepository
type MockRoleRepository struct {
	mock.Mock
}

func (m *MockRoleRepository) CreateRole(ctx context.Context, role *permission.Role) error {
	args := m.Called(ctx, role)
	return args.Error(0)
}

func (m *MockRoleRepository) GetRoleByID(ctx context.Context, id int) (*permission.Role, error) {
	args := m.Called(ctx, id)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	role := args.Get(0).(*permission.Role)
	return role, args.Error(1)
}

func (m *MockRoleRepository) GetRoleByNameAndOrg(
	ctx context.Context,
	name string,
	orgName string,
) (*permission.Role, error) {
	args := m.Called(ctx, name, orgName)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	role := args.Get(0).(*permission.Role)
	return role, args.Error(1)
}

func (m *MockRoleRepository) ListRolesByOrg(
	ctx context.Context,
	orgName string,
	includeSystem bool,
) ([]*permission.Role, error) {
	args := m.Called(ctx, orgName, includeSystem)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	roles := args.Get(0).([]*permission.Role)
	return roles, args.Error(1)
}

func (m *MockRoleRepository) UpdateRole(ctx context.Context, role *permission.Role) error {
	args := m.Called(ctx, role)
	return args.Error(0)
}

func (m *MockRoleRepository) DeleteRole(ctx context.Context, id int) error {
	args := m.Called(ctx, id)
	return args.Error(0)
}

// MockPermissionRepository is a mock implementation of permission.PermissionRepository
type MockPermissionRepository struct {
	mock.Mock
}

func (m *MockPermissionRepository) AddPermission(
	ctx context.Context,
	roleID int,
	orgName string,
	perm *permission.Permission,
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
) ([]*permission.Permission, error) {
	args := m.Called(ctx, roleID)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	permissions := args.Get(0).([]*permission.Permission)
	return permissions, args.Error(1)
}

func (m *MockPermissionRepository) ListPermissionsByRoleAndOrg(
	ctx context.Context,
	roleID int,
	orgName string,
) ([]*permission.Permission, error) {
	args := m.Called(ctx, roleID, orgName)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	permissions := args.Get(0).([]*permission.Permission)
	return permissions, args.Error(1)
}

func (m *MockPermissionRepository) DeletePermissionsByRole(ctx context.Context, roleID int) error {
	args := m.Called(ctx, roleID)
	return args.Error(0)
}

// MockUserRoleRepository is a mock implementation of permission.UserRoleRepository
type MockUserRoleRepository struct {
	mock.Mock
}

func (m *MockUserRoleRepository) AssignRole(ctx context.Context, userRole *permission.UserRole) error {
	args := m.Called(ctx, userRole)
	return args.Error(0)
}

func (m *MockUserRoleRepository) RevokeRole(
	ctx context.Context,
	userID string,
	roleID int,
	orgName string,
) error {
	args := m.Called(ctx, userID, roleID, orgName)
	return args.Error(0)
}

func (m *MockUserRoleRepository) ListUserRoles(
	ctx context.Context,
	userID string,
	orgName string,
) ([]*permission.UserRole, error) {
	args := m.Called(ctx, userID, orgName)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	userRoles := args.Get(0).([]*permission.UserRole)
	return userRoles, args.Error(1)
}

func (m *MockUserRoleRepository) ListRoleUsers(
	ctx context.Context,
	roleID int,
	orgName string,
) ([]*permission.UserRole, error) {
	args := m.Called(ctx, roleID, orgName)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	userRoles := args.Get(0).([]*permission.UserRole)
	return userRoles, args.Error(1)
}

func (m *MockUserRoleRepository) GetUserRole(
	ctx context.Context,
	userID string,
	roleID int,
	orgName string,
) (*permission.UserRole, error) {
	args := m.Called(ctx, userID, roleID, orgName)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	userRole := args.Get(0).(*permission.UserRole)
	return userRole, args.Error(1)
}

func (m *MockUserRoleRepository) DeleteUserRolesByRole(ctx context.Context, roleID int) error {
	args := m.Called(ctx, roleID)
	return args.Error(0)
}

// testContext creates a properly initialized context for testing
func testContext() context.Context {
	ctx := context.Background()
	cv := &ctxutils.HttpRequestContext{
		RequestId: "test-request-id",
		Lang:      "en",
		Method:    "TEST",
		Path:      "/test",
		ClientIP:  "127.0.0.1",
	}
	return ctxutils.NewHttpContext(ctx, cv)
}

// Test CreateCustomRole
func TestRoleService_CreateCustomRole_Success(t *testing.T) {
	// Setup
	mockRoleRepo := new(MockRoleRepository)
	mockPermRepo := new(MockPermissionRepository)
	mockUserRoleRepo := new(MockUserRoleRepository)

	service := NewRoleService(mockRoleRepo, mockPermRepo, mockUserRoleRepo)

	ctx := testContext()
	input := &CreateCustomRoleInput{
		Name:        "data-scientist",
		Description: "Data science team role",
		OrgName:     "test-org",
	}

	// Mock: Role doesn't exist
	mockRoleRepo.On("GetRoleByNameAndOrg", ctx, "data-scientist", "test-org").
		Return(nil, nil)

	// Mock: Create role succeeds
	mockRoleRepo.On("CreateRole", ctx, mock.MatchedBy(func(role *permission.Role) bool {
		return role.Name == "data-scientist" &&
			role.Description == "Data science team role" &&
			role.OrgName == "test-org" &&
			!role.IsSystem
	})).Return(nil)

	// Execute
	result, err := service.CreateCustomRole(ctx, input)

	// Assert
	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, "data-scientist", result.Name)
	assert.Equal(t, "Data science team role", result.Description)
	assert.Equal(t, "test-org", result.OrgName)
	assert.False(t, result.IsSystem)

	mockRoleRepo.AssertExpectations(t)
}

func TestRoleService_CreateCustomRole_SystemRoleName(t *testing.T) {
	// Setup
	mockRoleRepo := new(MockRoleRepository)
	mockPermRepo := new(MockPermissionRepository)
	mockUserRoleRepo := new(MockUserRoleRepository)

	service := NewRoleService(mockRoleRepo, mockPermRepo, mockUserRoleRepo)

	ctx := testContext()
	input := &CreateCustomRoleInput{
		Name:        "owner",
		Description: "Custom owner role",
		OrgName:     "test-org",
	}

	// Execute
	result, err := service.CreateCustomRole(ctx, input)

	// Assert
	assert.Error(t, err)
	assert.Nil(t, result)

	bizErr, ok := err.(*bizerrors.BusinessError)
	assert.True(t, ok)
	assert.True(t, bizErr.Info().IsConflictError())
	assert.Contains(t, err.Error(), "system role name")

	// Repository should not be called
	mockRoleRepo.AssertNotCalled(t, "CreateRole")
}

func TestRoleService_CreateCustomRole_DuplicateName(t *testing.T) {
	// Setup
	mockRoleRepo := new(MockRoleRepository)
	mockPermRepo := new(MockPermissionRepository)
	mockUserRoleRepo := new(MockUserRoleRepository)

	service := NewRoleService(mockRoleRepo, mockPermRepo, mockUserRoleRepo)

	ctx := testContext()
	input := &CreateCustomRoleInput{
		Name:        "data-scientist",
		Description: "Data science team role",
		OrgName:     "test-org",
	}

	// Mock: Role already exists
	existingRole := &permission.Role{
		ID:      10,
		Name:    "data-scientist",
		OrgName: "test-org",
	}
	mockRoleRepo.On("GetRoleByNameAndOrg", ctx, "data-scientist", "test-org").
		Return(existingRole, nil)

	// Execute
	result, err := service.CreateCustomRole(ctx, input)

	// Assert
	assert.Error(t, err)
	assert.Nil(t, result)

	bizErr, ok := err.(*bizerrors.BusinessError)
	assert.True(t, ok)
	assert.True(t, bizErr.Info().IsConflictError())
	assert.Contains(t, err.Error(), "already exists")

	mockRoleRepo.AssertExpectations(t)
	mockRoleRepo.AssertNotCalled(t, "CreateRole")
}

// Test UpdateRole
func TestRoleService_UpdateRole_Success(t *testing.T) {
	// Setup
	mockRoleRepo := new(MockRoleRepository)
	mockPermRepo := new(MockPermissionRepository)
	mockUserRoleRepo := new(MockUserRoleRepository)

	service := NewRoleService(mockRoleRepo, mockPermRepo, mockUserRoleRepo)

	ctx := testContext()
	roleID := 10
	input := &UpdateRoleInput{
		Name:        "senior-data-scientist",
		Description: "Senior data science role",
	}

	// Mock: Get existing role
	existingRole := &permission.Role{
		ID:       10,
		Name:     "data-scientist",
		OrgName:  "test-org",
		IsSystem: false,
	}
	mockRoleRepo.On("GetRoleByID", ctx, roleID).Return(existingRole, nil)

	// Mock: No name conflict
	mockRoleRepo.On("GetRoleByNameAndOrg", ctx, "senior-data-scientist", "test-org").
		Return(nil, nil)

	// Mock: Update succeeds
	mockRoleRepo.On("UpdateRole", ctx, mock.MatchedBy(func(role *permission.Role) bool {
		return role.ID == 10 &&
			role.Name == "senior-data-scientist" &&
			role.Description == "Senior data science role"
	})).Return(nil)

	// Execute
	result, err := service.UpdateRole(ctx, roleID, input)

	// Assert
	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, "senior-data-scientist", result.Name)
	assert.Equal(t, "Senior data science role", result.Description)

	mockRoleRepo.AssertExpectations(t)
}

func TestRoleService_UpdateRole_SystemRole(t *testing.T) {
	// Setup
	mockRoleRepo := new(MockRoleRepository)
	mockPermRepo := new(MockPermissionRepository)
	mockUserRoleRepo := new(MockUserRoleRepository)

	service := NewRoleService(mockRoleRepo, mockPermRepo, mockUserRoleRepo)

	ctx := testContext()
	roleID := 1
	input := &UpdateRoleInput{
		Name:        "super-owner",
		Description: "Modified owner",
	}

	// Mock: Get system role
	systemRole := &permission.Role{
		ID:       1,
		Name:     "owner",
		OrgName:  permission.SystemOrgName,
		IsSystem: true,
	}
	mockRoleRepo.On("GetRoleByID", ctx, roleID).Return(systemRole, nil)

	// Execute
	result, err := service.UpdateRole(ctx, roleID, input)

	// Assert
	assert.Error(t, err)
	assert.Nil(t, result)

	bizErr, ok := err.(*bizerrors.BusinessError)
	assert.True(t, ok)
	assert.True(t, bizErr.Info().IsOperationDeniedError())
	assert.Contains(t, err.Error(), "system role")

	mockRoleRepo.AssertExpectations(t)
	mockRoleRepo.AssertNotCalled(t, "UpdateRole")
}

func TestRoleService_UpdateRole_NotFound(t *testing.T) {
	// Setup
	mockRoleRepo := new(MockRoleRepository)
	mockPermRepo := new(MockPermissionRepository)
	mockUserRoleRepo := new(MockUserRoleRepository)

	service := NewRoleService(mockRoleRepo, mockPermRepo, mockUserRoleRepo)

	ctx := testContext()
	roleID := 999
	input := &UpdateRoleInput{
		Name: "updated-name",
	}

	// Mock: Role not found
	mockRoleRepo.On("GetRoleByID", ctx, roleID).Return(nil, nil)

	// Execute
	result, err := service.UpdateRole(ctx, roleID, input)

	// Assert
	assert.Error(t, err)
	assert.Nil(t, result)

	bizErr, ok := err.(*bizerrors.BusinessError)
	assert.True(t, ok)
	assert.True(t, bizErr.Info().IsNotFoundError())

	mockRoleRepo.AssertExpectations(t)
}

// Test DeleteRole
func TestRoleService_DeleteRole_Success(t *testing.T) {
	// Setup
	mockRoleRepo := new(MockRoleRepository)
	mockPermRepo := new(MockPermissionRepository)
	mockUserRoleRepo := new(MockUserRoleRepository)

	service := NewRoleService(mockRoleRepo, mockPermRepo, mockUserRoleRepo)

	ctx := testContext()
	roleID := 10

	// Mock: Get existing custom role
	customRole := &permission.Role{
		ID:       10,
		Name:     "data-scientist",
		OrgName:  "test-org",
		IsSystem: false,
	}
	mockRoleRepo.On("GetRoleByID", ctx, roleID).Return(customRole, nil)

	// Mock: Delete role (cascade delete will handle user_roles and permissions)
	mockRoleRepo.On("DeleteRole", ctx, roleID).Return(nil)

	// Execute
	err := service.DeleteRole(ctx, roleID)

	// Assert
	assert.NoError(t, err)

	mockRoleRepo.AssertExpectations(t)
	// User role and permission repos should not be called (cascade delete)
}

func TestRoleService_DeleteRole_SystemRole(t *testing.T) {
	// Setup
	mockRoleRepo := new(MockRoleRepository)
	mockPermRepo := new(MockPermissionRepository)
	mockUserRoleRepo := new(MockUserRoleRepository)

	service := NewRoleService(mockRoleRepo, mockPermRepo, mockUserRoleRepo)

	ctx := testContext()
	roleID := 1

	// Mock: Get system role
	systemRole := &permission.Role{
		ID:       1,
		Name:     "owner",
		OrgName:  permission.SystemOrgName,
		IsSystem: true,
	}
	mockRoleRepo.On("GetRoleByID", ctx, roleID).Return(systemRole, nil)

	// Execute
	err := service.DeleteRole(ctx, roleID)

	// Assert
	assert.Error(t, err)

	bizErr, ok := err.(*bizerrors.BusinessError)
	assert.True(t, ok)
	assert.True(t, bizErr.Info().IsOperationDeniedError())
	assert.Contains(t, err.Error(), "system role")

	mockRoleRepo.AssertExpectations(t)
	// Delete operations should not be called
	mockUserRoleRepo.AssertNotCalled(t, "DeleteUserRolesByRole")
	mockPermRepo.AssertNotCalled(t, "DeletePermissionsByRole")
	mockRoleRepo.AssertNotCalled(t, "DeleteRole")
}

// Test GetRole
func TestRoleService_GetRole_Success(t *testing.T) {
	// Setup
	mockRoleRepo := new(MockRoleRepository)
	mockPermRepo := new(MockPermissionRepository)
	mockUserRoleRepo := new(MockUserRoleRepository)

	service := NewRoleService(mockRoleRepo, mockPermRepo, mockUserRoleRepo)

	ctx := testContext()
	roleID := 10

	// Mock: Get role
	expectedRole := &permission.Role{
		ID:          10,
		Name:        "data-scientist",
		Description: "Data science role",
		OrgName:     "test-org",
		IsSystem:    false,
	}
	mockRoleRepo.On("GetRoleByID", ctx, roleID).Return(expectedRole, nil)

	// Execute
	result, err := service.GetRole(ctx, roleID)

	// Assert
	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, expectedRole.ID, result.ID)
	assert.Equal(t, expectedRole.Name, result.Name)

	mockRoleRepo.AssertExpectations(t)
}

func TestRoleService_GetRole_NotFound(t *testing.T) {
	// Setup
	mockRoleRepo := new(MockRoleRepository)
	mockPermRepo := new(MockPermissionRepository)
	mockUserRoleRepo := new(MockUserRoleRepository)

	service := NewRoleService(mockRoleRepo, mockPermRepo, mockUserRoleRepo)

	ctx := testContext()
	roleID := 999

	// Mock: Role not found
	mockRoleRepo.On("GetRoleByID", ctx, roleID).Return(nil, nil)

	// Execute
	result, err := service.GetRole(ctx, roleID)

	// Assert
	assert.Error(t, err)
	assert.Nil(t, result)

	bizErr, ok := err.(*bizerrors.BusinessError)
	assert.True(t, ok)
	assert.True(t, bizErr.Info().IsNotFoundError())

	mockRoleRepo.AssertExpectations(t)
}

// Test ListRoles
func TestRoleService_ListRoles_Success(t *testing.T) {
	// Setup
	mockRoleRepo := new(MockRoleRepository)
	mockPermRepo := new(MockPermissionRepository)
	mockUserRoleRepo := new(MockUserRoleRepository)

	service := NewRoleService(mockRoleRepo, mockPermRepo, mockUserRoleRepo)

	ctx := testContext()
	orgName := "test-org"
	includeSystem := true

	// Mock: List roles
	expectedRoles := []*permission.Role{
		{ID: 1, Name: "owner", IsSystem: true, OrgName: permission.SystemOrgName},
		{ID: 2, Name: "admin", IsSystem: true, OrgName: permission.SystemOrgName},
		{ID: 10, Name: "data-scientist", IsSystem: false, OrgName: "test-org"},
	}
	mockRoleRepo.On("ListRolesByOrg", ctx, orgName, includeSystem).
		Return(expectedRoles, nil)

	// Execute
	result, err := service.ListRoles(ctx, orgName, includeSystem)

	// Assert
	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Len(t, result, 3)
	assert.Equal(t, "owner", result[0].Name)
	assert.Equal(t, "data-scientist", result[2].Name)

	mockRoleRepo.AssertExpectations(t)
}

func TestRoleService_ListRoles_OnlyCustom(t *testing.T) {
	// Setup
	mockRoleRepo := new(MockRoleRepository)
	mockPermRepo := new(MockPermissionRepository)
	mockUserRoleRepo := new(MockUserRoleRepository)

	service := NewRoleService(mockRoleRepo, mockPermRepo, mockUserRoleRepo)

	ctx := testContext()
	orgName := "test-org"
	includeSystem := false

	// Mock: List only custom roles
	expectedRoles := []*permission.Role{
		{ID: 10, Name: "data-scientist", IsSystem: false, OrgName: "test-org"},
		{ID: 11, Name: "ml-engineer", IsSystem: false, OrgName: "test-org"},
	}
	mockRoleRepo.On("ListRolesByOrg", ctx, orgName, includeSystem).
		Return(expectedRoles, nil)

	// Execute
	result, err := service.ListRoles(ctx, orgName, includeSystem)

	// Assert
	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Len(t, result, 2)
	assert.False(t, result[0].IsSystem)
	assert.False(t, result[1].IsSystem)

	mockRoleRepo.AssertExpectations(t)
}

// Test edge cases
func TestRoleService_CreateCustomRole_EmptyName(t *testing.T) {
	// Setup
	mockRoleRepo := new(MockRoleRepository)
	mockPermRepo := new(MockPermissionRepository)
	mockUserRoleRepo := new(MockUserRoleRepository)

	service := NewRoleService(mockRoleRepo, mockPermRepo, mockUserRoleRepo)

	ctx := testContext()
	input := &CreateCustomRoleInput{
		Name:    "",
		OrgName: "test-org",
	}

	// Execute
	result, err := service.CreateCustomRole(ctx, input)

	// Assert
	assert.Error(t, err)
	assert.Nil(t, result)

	bizErr, ok := err.(*bizerrors.BusinessError)
	assert.True(t, ok)
	assert.True(t, bizErr.Info().IsParamInvalidError())
}

func TestRoleService_CreateCustomRole_EmptyOrgName(t *testing.T) {
	// Setup
	mockRoleRepo := new(MockRoleRepository)
	mockPermRepo := new(MockPermissionRepository)
	mockUserRoleRepo := new(MockUserRoleRepository)

	service := NewRoleService(mockRoleRepo, mockPermRepo, mockUserRoleRepo)

	ctx := testContext()
	input := &CreateCustomRoleInput{
		Name:    "data-scientist",
		OrgName: "",
	}

	// Execute
	result, err := service.CreateCustomRole(ctx, input)

	// Assert
	assert.Error(t, err)
	assert.Nil(t, result)

	bizErr, ok := err.(*bizerrors.BusinessError)
	assert.True(t, ok)
	assert.True(t, bizErr.Info().IsParamInvalidError())
}

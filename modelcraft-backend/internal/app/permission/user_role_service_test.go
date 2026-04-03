package permission

import (
	"context"
	"modelcraft/internal/domain/permission"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// MockPermissionVersionManager is a mock implementation of PermissionVersionManager
type MockPermissionVersionManager struct {
	mock.Mock
}

func (m *MockPermissionVersionManager) GetVersion(ctx context.Context, orgName, userID string) (int64, error) {
	args := m.Called(ctx, orgName, userID)
	return args.Get(0).(int64), args.Error(1)
}

func (m *MockPermissionVersionManager) IncrementVersion(ctx context.Context, orgName, userID string) (int64, error) {
	args := m.Called(ctx, orgName, userID)
	return args.Get(0).(int64), args.Error(1)
}

// TestUserRoleService_AssignRoleToUser_IncrementsVersion verifies that assigning a role
// increments the permission version to invalidate cache
func TestUserRoleService_AssignRoleToUser_IncrementsVersion(t *testing.T) {
	t.Run("should increment version after successful role assignment", func(t *testing.T) {
		// Arrange
		mockRoleRepo := new(MockRoleRepository)
		mockUserRoleRepo := new(MockUserRoleRepository)
		mockPermRepo := new(MockPermissionRepository)
		mockVersionManager := new(MockPermissionVersionManager)

		service := NewUserRoleService(
			mockRoleRepo,
			mockUserRoleRepo,
			mockPermRepo,
			mockVersionManager,
		)

		ctx := testContext()
		userID := "user-123"
		roleID := 1
		orgName := "test-org"

		// Mock role exists (system role)
		mockRole := &permission.Role{
			ID:          roleID,
			Name:        "editor",
			Description: "Editor role",
			OrgName:     permission.SystemOrgName, // System role
			IsSystem:    true,
		}
		mockRoleRepo.On("GetRoleByID", ctx, roleID).Return(mockRole, nil)

		// Mock role assignment doesn't exist yet
		mockUserRoleRepo.On("GetUserRole", ctx, userID, roleID, orgName).Return(nil, nil)

		// Mock successful role assignment
		mockUserRoleRepo.On("AssignRole", ctx, mock.AnythingOfType("*permission.UserRole")).Return(nil)

		// Mock version increment
		mockVersionManager.On("IncrementVersion", ctx, orgName, userID).Return(int64(2), nil)

		// Act
		err := service.AssignRoleToUser(ctx, userID, roleID, orgName)

		// Assert
		assert.NoError(t, err)
		mockVersionManager.AssertCalled(t, "IncrementVersion", ctx, orgName, userID)
		mockVersionManager.AssertNumberOfCalls(t, "IncrementVersion", 1)
	})

	t.Run("should not fail assignment if version increment fails", func(t *testing.T) {
		// Arrange
		mockRoleRepo := new(MockRoleRepository)
		mockUserRoleRepo := new(MockUserRoleRepository)
		mockPermRepo := new(MockPermissionRepository)
		mockVersionManager := new(MockPermissionVersionManager)

		service := NewUserRoleService(
			mockRoleRepo,
			mockUserRoleRepo,
			mockPermRepo,
			mockVersionManager,
		)

		ctx := testContext()
		userID := "user-123"
		roleID := 1
		orgName := "test-org"

		// Mock role exists
		mockRole := &permission.Role{
			ID:          roleID,
			Name:        "editor",
			Description: "Editor role",
			OrgName:     permission.SystemOrgName,
			IsSystem:    true,
		}
		mockRoleRepo.On("GetRoleByID", ctx, roleID).Return(mockRole, nil)

		// Mock role assignment doesn't exist
		mockUserRoleRepo.On("GetUserRole", ctx, userID, roleID, orgName).Return(nil, nil)

		// Mock successful role assignment
		mockUserRoleRepo.On("AssignRole", ctx, mock.AnythingOfType("*permission.UserRole")).Return(nil)

		// Mock version increment failure
		mockVersionManager.On("IncrementVersion", ctx, orgName, userID).Return(int64(0), assert.AnError)

		// Act
		err := service.AssignRoleToUser(ctx, userID, roleID, orgName)

		// Assert - assignment should succeed even if version increment fails
		assert.NoError(t, err)
		mockVersionManager.AssertCalled(t, "IncrementVersion", ctx, orgName, userID)
	})

	t.Run("should work with nil version manager for backward compatibility", func(t *testing.T) {
		// Arrange
		mockRoleRepo := new(MockRoleRepository)
		mockUserRoleRepo := new(MockUserRoleRepository)
		mockPermRepo := new(MockPermissionRepository)

		service := NewUserRoleService(
			mockRoleRepo,
			mockUserRoleRepo,
			mockPermRepo,
			nil, // No version manager
		)

		ctx := testContext()
		userID := "user-123"
		roleID := 1
		orgName := "test-org"

		// Mock role exists
		mockRole := &permission.Role{
			ID:          roleID,
			Name:        "editor",
			Description: "Editor role",
			OrgName:     permission.SystemOrgName,
			IsSystem:    true,
		}
		mockRoleRepo.On("GetRoleByID", ctx, roleID).Return(mockRole, nil)

		// Mock role assignment doesn't exist
		mockUserRoleRepo.On("GetUserRole", ctx, userID, roleID, orgName).Return(nil, nil)

		// Mock successful role assignment
		mockUserRoleRepo.On("AssignRole", ctx, mock.AnythingOfType("*permission.UserRole")).Return(nil)

		// Act
		err := service.AssignRoleToUser(ctx, userID, roleID, orgName)

		// Assert - should succeed without version manager
		assert.NoError(t, err)
	})
}

// TestUserRoleService_RevokeRoleFromUser_IncrementsVersion verifies that revoking a role
// also increments the permission version
func TestUserRoleService_RevokeRoleFromUser_IncrementsVersion(t *testing.T) {
	t.Run("should increment version after successful role revocation", func(t *testing.T) {
		// Arrange
		mockRoleRepo := new(MockRoleRepository)
		mockUserRoleRepo := new(MockUserRoleRepository)
		mockPermRepo := new(MockPermissionRepository)
		mockVersionManager := new(MockPermissionVersionManager)

		service := NewUserRoleService(
			mockRoleRepo,
			mockUserRoleRepo,
			mockPermRepo,
			mockVersionManager,
		)

		ctx := testContext()
		userID := "user-123"
		roleID := 1
		orgName := "test-org"

		// Mock role exists
		mockRole := &permission.Role{
			ID:          roleID,
			Name:        "editor",
			Description: "Editor role",
			OrgName:     permission.SystemOrgName,
			IsSystem:    true,
		}
		mockRoleRepo.On("GetRoleByID", ctx, roleID).Return(mockRole, nil)

		// Mock successful role revocation
		mockUserRoleRepo.On("RevokeRole", ctx, userID, roleID, orgName).Return(nil)

		// Mock version increment
		mockVersionManager.On("IncrementVersion", ctx, orgName, userID).Return(int64(2), nil)

		// Act
		err := service.RevokeRoleFromUser(ctx, userID, roleID, orgName)

		// Assert
		assert.NoError(t, err)
		mockVersionManager.AssertCalled(t, "IncrementVersion", ctx, orgName, userID)
		mockVersionManager.AssertNumberOfCalls(t, "IncrementVersion", 1)
	})

	t.Run("should not fail revocation if version increment fails", func(t *testing.T) {
		// Arrange
		mockRoleRepo := new(MockRoleRepository)
		mockUserRoleRepo := new(MockUserRoleRepository)
		mockPermRepo := new(MockPermissionRepository)
		mockVersionManager := new(MockPermissionVersionManager)

		service := NewUserRoleService(
			mockRoleRepo,
			mockUserRoleRepo,
			mockPermRepo,
			mockVersionManager,
		)

		ctx := testContext()
		userID := "user-123"
		roleID := 1
		orgName := "test-org"

		// Mock role exists
		mockRole := &permission.Role{
			ID:          roleID,
			Name:        "editor",
			Description: "Editor role",
			OrgName:     permission.SystemOrgName,
			IsSystem:    true,
		}
		mockRoleRepo.On("GetRoleByID", ctx, roleID).Return(mockRole, nil)

		// Mock successful role revocation
		mockUserRoleRepo.On("RevokeRole", ctx, userID, roleID, orgName).Return(nil)

		// Mock version increment failure
		mockVersionManager.On("IncrementVersion", ctx, orgName, userID).Return(int64(0), assert.AnError)

		// Act
		err := service.RevokeRoleFromUser(ctx, userID, roleID, orgName)

		// Assert - revocation should succeed even if version increment fails
		assert.NoError(t, err)
		mockVersionManager.AssertCalled(t, "IncrementVersion", ctx, orgName, userID)
	})

	t.Run("should work with nil version manager for backward compatibility", func(t *testing.T) {
		// Arrange
		mockRoleRepo := new(MockRoleRepository)
		mockUserRoleRepo := new(MockUserRoleRepository)
		mockPermRepo := new(MockPermissionRepository)

		service := NewUserRoleService(
			mockRoleRepo,
			mockUserRoleRepo,
			mockPermRepo,
			nil, // No version manager
		)

		ctx := testContext()
		userID := "user-123"
		roleID := 1
		orgName := "test-org"

		// Mock role exists
		mockRole := &permission.Role{
			ID:          roleID,
			Name:        "editor",
			Description: "Editor role",
			OrgName:     permission.SystemOrgName,
			IsSystem:    true,
		}
		mockRoleRepo.On("GetRoleByID", ctx, roleID).Return(mockRole, nil)

		// Mock successful role revocation
		mockUserRoleRepo.On("RevokeRole", ctx, userID, roleID, orgName).Return(nil)

		// Act
		err := service.RevokeRoleFromUser(ctx, userID, roleID, orgName)

		// Assert - should succeed without version manager
		assert.NoError(t, err)
	})
}

package role

import (
	"context"
	"fmt"
	"modelcraft/internal/domain/role"
	"modelcraft/pkg/bizerrors"

	"github.com/google/uuid"
)

// RoleAppService provides role management operations.
type RoleAppService struct {
	roleRepo role.RoleRepository
}

// NewRoleAppService creates a new RoleAppService.
func NewRoleAppService(roleRepo role.RoleRepository) *RoleAppService {
	return &RoleAppService{roleRepo: roleRepo}
}

// ListRoles returns all available roles.
func (s *RoleAppService) ListRoles(ctx context.Context) ([]*role.Role, error) {
	roles, err := s.roleRepo.List(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to list roles: %w", err)
	}
	return roles, nil
}

// GetRoleByID returns a role by its ID.
func (s *RoleAppService) GetRoleByID(ctx context.Context, id string) (*role.Role, error) {
	r, err := s.roleRepo.GetByID(ctx, id)
	if err != nil {
		return nil, fmt.Errorf("failed to get role: %w", err)
	}
	if r == nil {
		return nil, bizerrors.NewError(bizerrors.NotFound, "role not found")
	}
	return r, nil
}

// CreateRoleInput holds the input for creating a custom role.
type CreateRoleInput struct {
	Name        string
	Description string
	Permissions []string
}

// CreateRole creates a new custom role.
func (s *RoleAppService) CreateRole(ctx context.Context, input CreateRoleInput) (*role.Role, error) {
	// Check for duplicate name
	existing, err := s.roleRepo.GetByName(ctx, input.Name)
	if err != nil {
		return nil, fmt.Errorf("failed to check role name: %w", err)
	}
	if existing != nil {
		return nil, bizerrors.NewError(bizerrors.Conflict, "role '%s' already exists", input.Name)
	}

	r, err := role.NewRole(uuid.New().String(), input.Name, input.Description, input.Permissions, false)
	if err != nil {
		return nil, bizerrors.NewError(bizerrors.ParamInvalid, "invalid role: %v", err)
	}

	if err := s.roleRepo.Create(ctx, r); err != nil {
		return nil, fmt.Errorf("failed to create role: %w", err)
	}

	return r, nil
}

// DeleteRole deletes a custom role. System roles cannot be deleted.
func (s *RoleAppService) DeleteRole(ctx context.Context, id string) error {
	r, err := s.roleRepo.GetByID(ctx, id)
	if err != nil {
		return fmt.Errorf("failed to get role: %w", err)
	}
	if r == nil {
		return bizerrors.NewError(bizerrors.NotFound, "role not found")
	}

	if !r.CanDelete() {
		return bizerrors.NewError(bizerrors.OperationDenied, "system role '%s' cannot be deleted", r.Name)
	}

	if err := s.roleRepo.Delete(ctx, id); err != nil {
		return fmt.Errorf("failed to delete role: %w", err)
	}

	return nil
}

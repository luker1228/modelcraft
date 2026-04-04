package repository

import (
	"context"
	"modelcraft/internal/domain/role"
	"modelcraft/internal/domain/shared"
	"modelcraft/internal/infrastructure/dbgen"
	"modelcraft/internal/infrastructure/sqlerr"
	"strconv"
)

// SqlRoleRepository is the sqlc-based implementation of role.RoleRepository.
type SqlRoleRepository struct {
	q dbgen.Querier
}

// NewSqlRoleRepository creates a new SqlRoleRepository backed by the given sqlc Querier.
func NewSqlRoleRepository(q dbgen.Querier) role.RoleRepository {
	return &SqlRoleRepository{q: q}
}

// GetByID retrieves a role by its UUID string.
// The domain uses string IDs, but the DB uses int64. This method converts between them.
func (r *SqlRoleRepository) GetByID(ctx context.Context, id string) (*role.Role, error) {
	intID, err := strconv.ParseInt(id, 10, 64)
	if err != nil {
		// Invalid ID format, treat as not found
		return nil, shared.NewNotFoundError("invalid role id format: " + id)
	}

	var row dbgen.Role
	err = sqlerr.QueryWithSQLErrorHandling(func() error {
		var e error
		row, e = r.q.GetRoleByID(ctx, intID)
		return e
	})
	if err != nil {
		return nil, err
	}

	return roleToDomain(row), nil
}

// GetByName retrieves a role by its name.
func (r *SqlRoleRepository) GetByName(ctx context.Context, name string) (*role.Role, error) {
	var row dbgen.Role
	err := sqlerr.QueryWithSQLErrorHandling(func() error {
		var e error
		row, e = r.q.GetRoleByName(ctx, name)
		return e
	})
	if err != nil {
		return nil, err
	}

	return roleToDomain(row), nil
}

// GetSystemRoleByName retrieves a system role by its name (is_system=true).
func (r *SqlRoleRepository) GetSystemRoleByName(ctx context.Context, name string) (*role.Role, error) {
	var row dbgen.Role
	err := sqlerr.QueryWithSQLErrorHandling(func() error {
		var e error
		row, e = r.q.GetSystemRoleByName(ctx, name)
		return e
	})
	if err != nil {
		return nil, err
	}

	return roleToDomain(row), nil
}

// List retrieves all roles.
func (r *SqlRoleRepository) List(ctx context.Context) ([]*role.Role, error) {
	var rows []dbgen.Role
	err := sqlerr.QueryWithSQLErrorHandling(func() error {
		var e error
		rows, e = r.q.ListRoles(ctx)
		return e
	})
	if err != nil {
		return nil, err
	}

	roles := make([]*role.Role, len(rows))
	for i, row := range rows {
		roles[i] = roleToDomain(row)
	}

	return roles, nil
}

// Create persists a new role to the database.
func (r *SqlRoleRepository) Create(ctx context.Context, entity *role.Role) error {
	params := dbgen.CreateRoleParams{
		Name:        entity.Name,
		Description: sqlerr.PtrToNullStr(&entity.Description),
		IsSystem:    entity.IsSystem,
		OrgName:     "__SYSTEM__", // Default org for legacy roles
	}

	return sqlerr.ExecWithErrorHandling(func() error {
		_, err := r.q.CreateRole(ctx, params)
		return err
	})
}

// Update updates an existing role.
func (r *SqlRoleRepository) Update(ctx context.Context, entity *role.Role) error {
	intID, err := strconv.ParseInt(entity.ID, 10, 64)
	if err != nil {
		// Invalid ID format, no-op
		return nil
	}

	params := dbgen.UpdateRoleParams{
		ID:          intID,
		Description: sqlerr.PtrToNullStr(&entity.Description),
	}

	return sqlerr.ExecWithErrorHandling(func() error {
		return r.q.UpdateRole(ctx, params)
	})
}

// Delete removes a role by its ID.
func (r *SqlRoleRepository) Delete(ctx context.Context, id string) error {
	intID, err := strconv.ParseInt(id, 10, 64)
	if err != nil {
		// Invalid ID format, no-op
		return nil
	}

	return sqlerr.ExecWithErrorHandling(func() error {
		return r.q.DeleteRole(ctx, intID)
	})
}

// roleToDomain converts a dbgen.Role row to a domain Role entity.
// Permissions are fixed to an empty slice as the new DB schema stores permissions
// in a separate role_permissions table (managed by Casbin).
func roleToDomain(row dbgen.Role) *role.Role {
	description := ""
	if row.Description.Valid {
		description = row.Description.String
	}

	return &role.Role{
		ID:          strconv.FormatInt(row.ID, 10),
		Name:        row.Name,
		Description: description,
		Permissions: []string{}, // Permissions are managed by Casbin in role_permissions table
		IsSystem:    row.IsSystem,
		CreatedAt:   row.CreatedAt,
		UpdatedAt:   row.UpdatedAt,
	}
}

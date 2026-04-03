package adapter

import (
	"modelcraft/internal/interfaces/graphql/org/generated"
	"strconv"

	domainOrg "modelcraft/internal/domain/organization"
	domainPermission "modelcraft/internal/domain/permission"
	domainRole "modelcraft/internal/domain/role"
)

// UserManagementMapperInstance provides mapping functions for user management types.
var UserManagementMapperInstance = &userManagementMapper{}

type userManagementMapper struct{}

// ConvertOrganizationToGraphQL converts a domain Organization to a GraphQL Organization.
func (m *userManagementMapper) ConvertOrganizationToGraphQL(org *domainOrg.Organization) *generated.Organization {
	if org == nil {
		return nil
	}

	var displayName *string
	if org.DisplayName != "" {
		displayName = &org.DisplayName
	}

	return &generated.Organization{
		ID:          org.Name,
		Name:        org.Name,
		DisplayName: displayName,
		OwnerID:     org.OwnerID,
		Status:      convertOrgStatusToGraphQL(org.Status),
		CreatedAt:   org.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
		UpdatedAt:   org.UpdatedAt.Format("2006-01-02T15:04:05Z07:00"),
	}
}

// ConvertOrganizationsToGraphQL converts a slice of domain Organizations.
func (m *userManagementMapper) ConvertOrganizationsToGraphQL(orgs []*domainOrg.Organization) []*generated.Organization {
	result := make([]*generated.Organization, len(orgs))
	for i, org := range orgs {
		result[i] = m.ConvertOrganizationToGraphQL(org)
	}
	return result
}

// ConvertRoleToGraphQL converts a domain Role to a GraphQL Role.
func (m *userManagementMapper) ConvertRoleToGraphQL(r *domainRole.Role) *generated.Role {
	if r == nil {
		return nil
	}

	var description *string
	if r.Description != "" {
		description = &r.Description
	}

	return &generated.Role{
		ID:          r.ID,
		Name:        r.Name,
		Description: description,
		Permissions: r.Permissions,
		IsSystem:    r.IsSystem,
		CreatedAt:   r.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
		UpdatedAt:   r.UpdatedAt.Format("2006-01-02T15:04:05Z07:00"),
	}
}

// ConvertRolesToGraphQL converts a slice of domain Roles.
func (m *userManagementMapper) ConvertRolesToGraphQL(roles []*domainRole.Role) []*generated.Role {
	result := make([]*generated.Role, len(roles))
	for i, r := range roles {
		result[i] = m.ConvertRoleToGraphQL(r)
	}
	return result
}

// ConvertCasbinRoleToGraphQL converts a Casbin permission Role to a GraphQL Role.
// Note: Casbin roles don't have inline permissions - permissions are queried separately.
func (m *userManagementMapper) ConvertCasbinRoleToGraphQL(r *domainPermission.Role) *generated.Role {
	if r == nil {
		return nil
	}

	var description *string
	if r.Description != "" {
		description = &r.Description
	}

	return &generated.Role{
		ID:          strconv.Itoa(r.ID), // Convert int ID to string for GraphQL
		Name:        r.Name,
		Description: description,
		Permissions: []string{}, // Casbin permissions are stored separately, not in Role entity
		IsSystem:    r.IsSystem,
		CreatedAt:   r.CreatedAt.Format("2006-01-02T15:04:05Z07:00"),
		UpdatedAt:   r.UpdatedAt.Format("2006-01-02T15:04:05Z07:00"),
	}
}

func convertOrgStatusToGraphQL(status domainOrg.OrgStatus) generated.OrganizationStatus {
	switch status {
	case domainOrg.OrgStatusActive:
		return generated.OrganizationStatusActive
	case domainOrg.OrgStatusSuspended:
		return generated.OrganizationStatusSuspended
	case domainOrg.OrgStatusDeleted:
		return generated.OrganizationStatusDeleted
	default:
		return generated.OrganizationStatusActive
	}
}

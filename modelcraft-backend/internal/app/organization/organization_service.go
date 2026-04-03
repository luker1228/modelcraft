package organization

import (
	"context"
	"fmt"
	"modelcraft/internal/app/permission"
	"modelcraft/internal/domain/membership"
	"modelcraft/internal/domain/organization"
	"modelcraft/internal/domain/user"
	"modelcraft/pkg/bizerrors"

	domainPermission "modelcraft/internal/domain/permission"
)

// OrganizationAppService provides organization management operations.
type OrganizationAppService struct {
	orgRepo         organization.OrganizationRepository
	userRepo        user.UserRepository
	membershipRepo  membership.MembershipRepository
	roleRepo        domainPermission.RoleRepository
	userRoleService *permission.UserRoleService
}

// NewOrganizationAppService creates a new OrganizationAppService.
func NewOrganizationAppService(
	orgRepo organization.OrganizationRepository,
	userRepo user.UserRepository,
	membershipRepo membership.MembershipRepository,
	roleRepo domainPermission.RoleRepository,
	userRoleService *permission.UserRoleService,
) *OrganizationAppService {
	return &OrganizationAppService{
		orgRepo:         orgRepo,
		userRepo:        userRepo,
		membershipRepo:  membershipRepo,
		roleRepo:        roleRepo,
		userRoleService: userRoleService,
	}
}

// GetOrganizationByName retrieves an organization by its unique name.
func (s *OrganizationAppService) GetOrganizationByName(
	ctx context.Context,
	name string,
) (*organization.Organization, error) {
	org, err := s.orgRepo.GetByName(ctx, name)
	if err != nil {
		return nil, fmt.Errorf("failed to get organization: %w", err)
	}
	if org == nil {
		return nil, bizerrors.NewError(bizerrors.NotFound, "organization '%s' not found", name)
	}
	return org, nil
}

// ListOrganizationsByUser returns all organizations a user belongs to.
func (s *OrganizationAppService) ListOrganizationsByUser(
	ctx context.Context,
	userExternalID string,
) ([]*organization.Organization, error) {
	u, err := s.userRepo.GetByExternalID(ctx, userExternalID)
	if err != nil {
		return nil, fmt.Errorf("failed to get user: %w", err)
	}
	if u == nil {
		return nil, bizerrors.NewError(bizerrors.NotFound, "user not found")
	}

	orgs, err := s.orgRepo.ListByUser(ctx, u.ID)
	if err != nil {
		return nil, fmt.Errorf("failed to list organizations: %w", err)
	}
	return orgs, nil
}

// UpdateOrganizationDisplayName updates the display name of an organization.
func (s *OrganizationAppService) UpdateOrganizationDisplayName(
	ctx context.Context,
	orgName string,
	displayName string,
) (*organization.Organization, error) {
	org, err := s.orgRepo.GetByName(ctx, orgName)
	if err != nil {
		return nil, fmt.Errorf("failed to get organization: %w", err)
	}
	if org == nil {
		return nil, bizerrors.NewError(bizerrors.NotFound, "organization '%s' not found", orgName)
	}

	org.UpdateDisplayName(displayName)

	if err := s.orgRepo.Update(ctx, org); err != nil {
		return nil, fmt.Errorf("failed to update organization: %w", err)
	}

	return org, nil
}

// ListMembers returns all members of an organization with their roles.
type OrgMember struct {
	Membership *membership.Membership
	Role       *domainPermission.Role
	UserName   string
}

// ListMembers lists all members in an organization with their roles.
func (s *OrganizationAppService) ListMembers(ctx context.Context, orgName string) ([]*OrgMember, error) {
	org, err := s.orgRepo.GetByName(ctx, orgName)
	if err != nil {
		return nil, fmt.Errorf("failed to get organization: %w", err)
	}
	if org == nil {
		return nil, bizerrors.NewError(bizerrors.NotFound, "organization '%s' not found", orgName)
	}

	memberships, err := s.membershipRepo.ListByOrgWithUserName(ctx, orgName)
	if err != nil {
		return nil, fmt.Errorf("failed to list memberships: %w", err)
	}

	result := make([]*OrgMember, 0, len(memberships))
	for _, ms := range memberships {
		// Query user roles from user_roles table
		userRoles, err := s.userRoleService.ListUserRoles(ctx, ms.Membership.UserID, orgName)
		if err != nil {
			return nil, fmt.Errorf("failed to list user roles for userID=%s: %w", ms.Membership.UserID, err)
		}

		// Get the first role (users typically have one primary role per org)
		var role *domainPermission.Role
		if len(userRoles) > 0 {
			role, err = s.roleRepo.GetRoleByID(ctx, userRoles[0].RoleID)
			if err != nil {
				return nil, fmt.Errorf("failed to get role for roleID=%d: %w", userRoles[0].RoleID, err)
			}
		}

		result = append(result, &OrgMember{
			Membership: ms.Membership,
			Role:       role,
			UserName:   ms.UserName,
		})
	}

	return result, nil
}

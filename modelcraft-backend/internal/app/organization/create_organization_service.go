package organization

import (
	"context"
	"fmt"
	"modelcraft/internal/domain/organization"
	"modelcraft/internal/domain/shared"
	"modelcraft/internal/domain/user"
	"modelcraft/internal/infrastructure/dbgen"
	"modelcraft/internal/infrastructure/repository"
	"modelcraft/pkg/bizerrors"
	"modelcraft/pkg/bizutils"
	"modelcraft/pkg/logfacade"
	"regexp"
	"strings"

	domainPermission "modelcraft/internal/domain/permission"
)

// CreateOrganizationInput 创建组织的输入参数
type CreateOrganizationInput struct {
	DisplayName      string // Required: human-readable display name
	OrganizationName string // Optional: slug (auto-generated if empty)
	OwnerUserID      string
	Phone            string // Org 注册手机号
	// IsNewUser indicates the owner user has not been persisted yet (registration flow).
	// When true, resolveUserAndCheckOrg is skipped; OwnerUserID is used directly.
	IsNewUser bool
}

// CreateOrganizationOutput 创建组织的输出结果
type CreateOrganizationOutput struct {
	OrganizationID   string
	OrganizationName string
	DisplayName      string
	OwnerUserID      string
	RoleID           int64
	// AlreadyExisted is true when the user already had an organization and this call was a no-op.
	AlreadyExisted bool
}

// CreateOrganizationService 创建组织的应用服务
// 负责编排组织创建的完整流程：组织、角色分配
type CreateOrganizationService struct {
	txManager repository.TxManager
	userRepo  user.UserRepository
	orgRepo   organization.OrganizationRepository
	roleRepo  domainPermission.RoleRepository
}

// NewCreateOrganizationService 创建组织服务实例
func NewCreateOrganizationService(
	txManager repository.TxManager,
	userRepo user.UserRepository,
	orgRepo organization.OrganizationRepository,
	roleRepo domainPermission.RoleRepository,
) *CreateOrganizationService {
	return &CreateOrganizationService{
		txManager: txManager,
		userRepo:  userRepo,
		orgRepo:   orgRepo,
		roleRepo:  roleRepo,
	}
}

// Execute 执行创建组织的完整流程
func (s *CreateOrganizationService) Execute(
	ctx context.Context, input *CreateOrganizationInput,
) (*CreateOrganizationOutput, error) {
	logger := logfacade.GetLogger(ctx)
	logger.Infof(ctx, "Starting organization creation: displayName=%s, orgName=%s, userID=%s",
		input.DisplayName, input.OrganizationName, input.OwnerUserID)

	orgSlug, displayName, err := s.resolveOrgSlugAndDisplayName(ctx, input)
	if err != nil {
		return nil, err
	}

	existingUser, err := s.validateUser(ctx, input.OwnerUserID)
	if err != nil {
		return nil, err
	}

	if output := s.handleExistingOrganization(ctx, existingUser.ID); output != nil {
		return output, nil
	}

	if err := s.checkOrgNameAvailable(ctx, orgSlug); err != nil {
		return nil, err
	}

	ownerRole, err := s.getOwnerRole(ctx)
	if err != nil {
		return nil, err
	}

	output, err := s.createOrganizationInTransaction(
		ctx, orgSlug, displayName, existingUser.ID, input.Phone, ownerRole.ID,
	)
	if err != nil {
		return nil, err
	}

	if output != nil {
		logger.Infof(ctx, "Organization created successfully: orgID=%s, orgName=%s",
			output.OrganizationID, output.OrganizationName)
	}

	return output, nil
}

// ExecuteWithQuerier executes organization creation logic on the provided transaction querier.
func (s *CreateOrganizationService) ExecuteWithQuerier(
	ctx context.Context, q dbgen.Querier, input *CreateOrganizationInput,
) (*CreateOrganizationOutput, error) {
	logger := logfacade.GetLogger(ctx)
	logger.Infof(ctx, "Starting organization creation: displayName=%s, orgName=%s, userID=%s",
		input.DisplayName, input.OrganizationName, input.OwnerUserID)

	orgSlug, displayName, err := s.resolveOrgSlugAndDisplayName(ctx, input)
	if err != nil {
		return nil, err
	}

	userRepo := repository.NewSqlUserRepository(q)
	orgRepo := repository.NewSqlOrganizationRepository(q)

	ownerUserID := input.OwnerUserID
	if !input.IsNewUser {
		existingUser, alreadyExisted, err := s.resolveUserAndCheckOrg(
			ctx, userRepo, orgRepo, input.OwnerUserID,
		)
		if err != nil {
			return nil, err
		}
		if alreadyExisted != nil {
			return alreadyExisted, nil
		}
		ownerUserID = existingUser.ID
	}

	orgExists, err := orgRepo.ExistsByName(ctx, orgSlug)
	if err != nil {
		logger.Errorf(ctx, err, "Failed to check organization name")
		return nil, bizerrors.Wrap(err, "Failed to check organization name")
	}
	if orgExists {
		logger.Warnf(ctx, "Organization name already exists: orgName=%s", orgSlug)
		return nil, bizerrors.NewError(bizerrors.OrganizationAlreadyExists, orgSlug)
	}

	ownerRole, err := s.ensureSystemRoles(ctx)
	if err != nil {
		return nil, err
	}

	org, txErr := organization.NewOrganization(orgSlug, displayName, ownerUserID, input.Phone)
	if txErr != nil {
		return nil, bizerrors.Wrap(txErr, "Invalid organization data")
	}
	if txErr = orgRepo.Create(ctx, org); txErr != nil {
		if shared.IsDuplicateKeyError(txErr) {
			return nil, bizerrors.NewError(bizerrors.OrganizationAlreadyExists, orgSlug)
		}
		return nil, bizerrors.Wrap(txErr, "Failed to create organization")
	}

	// Role assignment is deferred to the caller so that the user row exists first.
	// Caller should use NewSqlCasbinUserRoleRepository(q) with output.RoleID after
	// creating the user. This avoids FK violations on user_roles.user_id → users.id
	// when org and user are created in the same transaction.

	return &CreateOrganizationOutput{
		OrganizationID:   orgSlug,
		OrganizationName: orgSlug,
		DisplayName:      displayName,
		OwnerUserID:      ownerUserID,
		RoleID:           int64(ownerRole.ID),
	}, nil
}

// resolveUserAndCheckOrg looks up the user and returns early if they already have an org.
func (s *CreateOrganizationService) resolveUserAndCheckOrg(
	ctx context.Context,
	userRepo user.UserRepository,
	orgRepo organization.OrganizationRepository,
	ownerUserID string,
) (*user.User, *CreateOrganizationOutput, error) {
	logger := logfacade.GetLogger(ctx)
	existingUser, err := userRepo.GetByID(ctx, ownerUserID)
	if err != nil {
		if shared.IsNotFoundError(err) {
			return nil, nil, bizerrors.NewError(bizerrors.UserNotFound, ownerUserID)
		}
		return nil, nil, bizerrors.Wrap(err, "Failed to look up user")
	}
	if existingUser == nil {
		return nil, nil, bizerrors.NewError(bizerrors.UserNotFound, ownerUserID)
	}
	logger.Infof(ctx, "User validated: userID=%s", existingUser.ID)

	existingOrgs, err := orgRepo.ListByUser(ctx, existingUser.ID)
	if err != nil {
		return nil, nil, bizerrors.Wrap(err, "Failed to list user organizations")
	}
	if len(existingOrgs) > 0 {
		org := existingOrgs[0]
		return existingUser, &CreateOrganizationOutput{
			OrganizationID:   org.Name,
			OrganizationName: org.Name,
			DisplayName:      org.DisplayName,
			OwnerUserID:      existingUser.ID,
			AlreadyExisted:   true,
		}, nil
	}
	return existingUser, nil, nil
}

// ensureSystemRoles ensures all system roles exist and returns the owner role.
func (s *CreateOrganizationService) ensureSystemRoles(ctx context.Context) (*domainPermission.Role, error) {
	logger := logfacade.GetLogger(ctx)
	var ownerRole *domainPermission.Role
	for _, roleName := range domainPermission.SystemRoles {
		role, roleErr := s.roleRepo.GetRoleByNameAndOrg(ctx, roleName, domainPermission.SystemOrgName)
		if roleErr != nil {
			if !shared.IsNotFoundError(roleErr) {
				return nil, bizerrors.Wrapf(roleErr, "failed to get system role: %s", roleName)
			}
			role = nil
		}
		if role == nil {
			role = &domainPermission.Role{
				Name:     roleName,
				IsSystem: true,
				OrgName:  domainPermission.SystemOrgName,
			}
			if createErr := s.roleRepo.CreateRole(ctx, role); createErr != nil {
				return nil, bizerrors.Wrapf(createErr, "failed to recreate system role: %s", roleName)
			}
		}
		if roleName == domainPermission.RoleOwner {
			ownerRole = role
		}
	}
	_ = logger
	return ownerRole, nil
}

// resolveOrgSlugAndDisplayName resolves and validates orgSlug and displayName.
func (s *CreateOrganizationService) resolveOrgSlugAndDisplayName(
	ctx context.Context, input *CreateOrganizationInput,
) (orgSlug, displayName string, err error) {
	logger := logfacade.GetLogger(ctx)

	if input.OrganizationName != "" {
		if err := validateOrgSlugFormat(input.OrganizationName); err != nil {
			return "", "", err
		}
		orgSlug = input.OrganizationName
	} else {
		orgSlug = bizutils.GenerateSlugWithLength(input.DisplayName, 6, 24)
		logger.Infof(ctx, "Auto-generated slug: %s", orgSlug)
	}

	displayName = strings.TrimSpace(input.DisplayName)
	if displayName == "" {
		displayName = orgSlug
	} else if len(displayName) > 255 {
		return "", "", bizerrors.NewError(bizerrors.ParamInvalid, "display name must not exceed 255 characters")
	}

	return orgSlug, displayName, nil
}

// validateUser validates that the user exists.
func (s *CreateOrganizationService) validateUser(ctx context.Context, userID string) (*user.User, error) {
	logger := logfacade.GetLogger(ctx)
	existingUser, err := s.userRepo.GetByID(ctx, userID)
	if err != nil {
		if shared.IsNotFoundError(err) {
			return nil, bizerrors.NewError(bizerrors.UserNotFound, userID)
		}
		return nil, bizerrors.Wrap(err, "Failed to look up user")
	}
	if existingUser == nil {
		return nil, bizerrors.NewError(bizerrors.UserNotFound, userID)
	}
	logger.Infof(ctx, "User validated: userID=%s", existingUser.ID)
	return existingUser, nil
}

// handleExistingOrganization returns existing org if user already has one (idempotent).
func (s *CreateOrganizationService) handleExistingOrganization(
	ctx context.Context, userID string,
) *CreateOrganizationOutput {
	logger := logfacade.GetLogger(ctx)
	existingOrgs, err := s.orgRepo.ListByUser(ctx, userID)
	if err != nil {
		logger.Errorf(ctx, err, "Failed to list user organizations")
		return nil
	}
	if len(existingOrgs) == 0 {
		return nil
	}
	logger.Infof(ctx, "User already has an organization (idempotent): userID=%s", userID)
	org := existingOrgs[0]
	return &CreateOrganizationOutput{
		OrganizationID:   org.Name,
		OrganizationName: org.Name,
		DisplayName:      org.DisplayName,
		OwnerUserID:      userID,
		AlreadyExisted:   true,
	}
}

// checkOrgNameAvailable checks that the organization name is available.
func (s *CreateOrganizationService) checkOrgNameAvailable(ctx context.Context, orgSlug string) error {
	logger := logfacade.GetLogger(ctx)
	orgExists, err := s.orgRepo.ExistsByName(ctx, orgSlug)
	if err != nil {
		return bizerrors.Wrap(err, "Failed to check organization name")
	}
	if orgExists {
		logger.Warnf(ctx, "Organization name already exists: orgName=%s", orgSlug)
		return bizerrors.NewError(bizerrors.OrganizationAlreadyExists, orgSlug)
	}
	return nil
}

// getOwnerRole ensures system roles exist and returns the owner role.
func (s *CreateOrganizationService) getOwnerRole(ctx context.Context) (*domainPermission.Role, error) {
	logger := logfacade.GetLogger(ctx)
	var ownerRole *domainPermission.Role

	for _, roleName := range domainPermission.SystemRoles {
		role, err := s.roleRepo.GetRoleByNameAndOrg(ctx, roleName, domainPermission.SystemOrgName)
		if err != nil {
			if !shared.IsNotFoundError(err) {
				return nil, bizerrors.Wrapf(err, "failed to get system role: %s", roleName)
			}
			role = nil
		}
		if role == nil {
			logger.Warnf(ctx, "System role %q not found — recreating", roleName)
			role = &domainPermission.Role{
				Name:     roleName,
				IsSystem: true,
				OrgName:  domainPermission.SystemOrgName,
			}
			if createErr := s.roleRepo.CreateRole(ctx, role); createErr != nil {
				return nil, bizerrors.Wrapf(createErr, "failed to recreate system role: %s", roleName)
			}
		}
		if roleName == domainPermission.RoleOwner {
			ownerRole = role
		}
	}

	logger.Infof(ctx, "Owner role ready: roleID=%d", ownerRole.ID)
	return ownerRole, nil
}

// createOrganizationInTransaction executes organization creation in a transaction.
func (s *CreateOrganizationService) createOrganizationInTransaction(
	ctx context.Context, orgSlug, displayName, userID, phone string, roleID int,
) (*CreateOrganizationOutput, error) {
	logger := logfacade.GetLogger(ctx)
	var output *CreateOrganizationOutput

	err := s.txManager.WithTx(ctx, func(ctx context.Context, q dbgen.Querier) error {
		org, txErr := organization.NewOrganization(orgSlug, displayName, userID, phone)
		if txErr != nil {
			return bizerrors.Wrap(txErr, "Invalid organization data")
		}
		orgRepository := repository.NewSqlOrganizationRepository(q)
		if txErr = orgRepository.Create(ctx, org); txErr != nil {
			if shared.IsDuplicateKeyError(txErr) {
				return bizerrors.NewError(bizerrors.OrganizationAlreadyExists, orgSlug)
			}
			return bizerrors.Wrap(txErr, "Failed to create organization")
		}
		logger.Infof(ctx, "Organization created: orgName=%s", orgSlug)

		userRoleRepository := repository.NewSqlCasbinUserRoleRepository(q)
		userRole := &domainPermission.UserRole{UserID: userID, RoleID: roleID, OrgName: orgSlug}
		if txErr = userRoleRepository.AssignRole(ctx, userRole); txErr != nil {
			return bizerrors.Wrap(txErr, "Failed to assign owner role")
		}
		logger.Infof(ctx, "Owner role assigned successfully")

		output = &CreateOrganizationOutput{
			OrganizationID:   orgSlug,
			OrganizationName: orgSlug,
			DisplayName:      displayName,
			OwnerUserID:      userID,
			RoleID:           int64(roleID),
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	return output, nil
}

// validateOrgSlugFormat validates the organization slug format.
func validateOrgSlugFormat(slug string) error {
	if len(slug) < 6 || len(slug) > 24 {
		return bizerrors.NewError(bizerrors.ParamInvalid, fmt.Sprintf(
			"organization slug must be 6-24 characters (got %d)", len(slug)))
	}
	slugPattern := regexp.MustCompile(`^[a-z][a-z0-9_]*$`)
	if !slugPattern.MatchString(slug) {
		return bizerrors.NewError(
			bizerrors.ParamInvalid,
			"organization slug must start with lowercase letter and contain only letters, numbers, and underscores",
		)
	}
	return nil
}

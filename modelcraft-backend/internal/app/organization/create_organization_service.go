package organization

import (
	"context"
	"fmt"
	"modelcraft/internal/domain/membership"
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

	"github.com/google/uuid"

	domainPermission "modelcraft/internal/domain/permission"
)

// CreateOrganizationInput 创建组织的输入参数
type CreateOrganizationInput struct {
	DisplayName      string // Required: human-readable display name
	OrganizationName string // Optional: slug (auto-generated if empty)
	OwnerUserID      string
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
// 负责编排组织创建的完整流程：组织、成员关系、角色分配
type CreateOrganizationService struct {
	txManager      repository.TxManager
	userRepo       user.UserRepository
	orgRepo        organization.OrganizationRepository
	roleRepo       domainPermission.RoleRepository
	membershipRepo membership.MembershipRepository
}

// NewCreateOrganizationService 创建组织服务实例
func NewCreateOrganizationService(
	txManager repository.TxManager,
	userRepo user.UserRepository,
	orgRepo organization.OrganizationRepository,
	roleRepo domainPermission.RoleRepository,
	membershipRepo membership.MembershipRepository,
) *CreateOrganizationService {
	return &CreateOrganizationService{
		txManager:      txManager,
		userRepo:       userRepo,
		orgRepo:        orgRepo,
		roleRepo:       roleRepo,
		membershipRepo: membershipRepo,
	}
}

// Execute 执行创建组织的完整流程
// 流程包括：
// 1. 验证并生成组织 slug（如果未提供）
// 2. 验证用户存在
// 3. 检查组织名是否已存在
// 4. 获取 owner 系统角色
// 5. 在事务中执行：创建组织 -> 创建成员关系 -> 分配角色
func (s *CreateOrganizationService) Execute(
	ctx context.Context, input *CreateOrganizationInput,
) (*CreateOrganizationOutput, error) {
	logger := logfacade.GetLogger(ctx)
	logger.Infof(ctx, "Starting organization creation: displayName=%s, orgName=%s, userID=%s",
		input.DisplayName, input.OrganizationName, input.OwnerUserID)

	// Step 1: Resolve orgSlug and displayName
	orgSlug, displayName, err := s.resolveOrgSlugAndDisplayName(ctx, input)
	if err != nil {
		return nil, err
	}

	// Step 2: Validate user exists
	existingUser, err := s.validateUser(ctx, input.OwnerUserID)
	if err != nil {
		return nil, err
	}

	// Step 3: Check idempotent - return existing org if user already has one
	if output := s.handleExistingOrganization(ctx, existingUser.ID); output != nil {
		return output, nil
	}

	// Step 4: Check organization name doesn't exist
	if err := s.checkOrgNameAvailable(ctx, orgSlug); err != nil {
		return nil, err
	}

	// Step 5: Get owner system role
	ownerRole, err := s.getOwnerRole(ctx)
	if err != nil {
		return nil, err
	}

	// Step 6: Execute transaction
	output, err := s.createOrganizationInTransaction(
		ctx, orgSlug, displayName, existingUser.ID, ownerRole.ID,
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

// resolveOrgSlugAndDisplayName resolves and validates orgSlug and displayName
func (s *CreateOrganizationService) resolveOrgSlugAndDisplayName(
	ctx context.Context, input *CreateOrganizationInput,
) (orgSlug, displayName string, err error) {
	logger := logfacade.GetLogger(ctx)

	// Resolve orgSlug: validate if provided, otherwise auto-generate from displayName
	if input.OrganizationName != "" {
		if err := validateOrgSlugFormat(input.OrganizationName); err != nil {
			logger.Warn(ctx, "Invalid organization slug format", logfacade.Err(err))
			return "", "", err
		}
		orgSlug = input.OrganizationName
	} else {
		orgSlug = bizutils.GenerateSlugWithLength(input.DisplayName, 6, 24)
		logger.Infof(ctx, "Auto-generated organization slug: %s from displayName: %s", orgSlug, input.DisplayName)
	}

	// Resolve displayName: fall back to orgSlug if blank
	displayName = strings.TrimSpace(input.DisplayName)
	if displayName == "" {
		displayName = orgSlug
		logger.Infof(ctx, "Display name not provided, using organization slug: %s", orgSlug)
	} else if len(displayName) > 255 {
		return "", "", bizerrors.NewError(bizerrors.ParamInvalid, "display name must not exceed 255 characters")
	}

	logger.Infof(ctx, "Using organization slug: %s, displayName: %s", orgSlug, displayName)
	return orgSlug, displayName, nil
}

// validateUser validates that the user exists
func (s *CreateOrganizationService) validateUser(ctx context.Context, userID string) (*user.User, error) {
	logger := logfacade.GetLogger(ctx)
	existingUser, err := s.userRepo.GetByID(ctx, userID)
	if err != nil {
		if shared.IsNotFoundError(err) {
			logger.Warnf(ctx, "User not found: userID=%s", userID)
			return nil, bizerrors.NewError(bizerrors.UserNotFound, userID)
		}
		logger.Error(ctx, "Failed to look up user", logfacade.Err(err))
		return nil, bizerrors.Wrap(err, "Failed to look up user")
	}
	if existingUser == nil {
		logger.Warnf(ctx, "User not found: userID=%s", userID)
		return nil, bizerrors.NewError(bizerrors.UserNotFound, userID)
	}
	logger.Infof(ctx, "User validated: userID=%s", existingUser.ID)
	return existingUser, nil
}

// handleExistingOrganization returns existing organization if user already has one (idempotent)
func (s *CreateOrganizationService) handleExistingOrganization(
	ctx context.Context, userID string,
) *CreateOrganizationOutput {
	logger := logfacade.GetLogger(ctx)
	orgCount, err := s.membershipRepo.CountByUser(ctx, userID)
	if err != nil {
		logger.Error(ctx, "Failed to count user organizations", logfacade.Err(err))
		return nil
	}
	if orgCount < 1 {
		return nil
	}

	logger.Infof(ctx, "User already has an organization, returning existing one (idempotent): userID=%s", userID)
	existingOrgs, listErr := s.orgRepo.ListByUser(ctx, userID)
	if listErr != nil {
		logger.Error(ctx, "Failed to list existing user organizations", logfacade.Err(listErr))
		return nil
	}
	if len(existingOrgs) == 0 {
		return nil
	}

	org := existingOrgs[0]
	return &CreateOrganizationOutput{
		OrganizationID:   org.Name,
		OrganizationName: org.Name,
		DisplayName:      org.DisplayName,
		OwnerUserID:      userID,
		AlreadyExisted:   true,
	}
}

// checkOrgNameAvailable checks that the organization name is available
func (s *CreateOrganizationService) checkOrgNameAvailable(ctx context.Context, orgSlug string) error {
	logger := logfacade.GetLogger(ctx)
	orgExists, err := s.orgRepo.ExistsByName(ctx, orgSlug)
	if err != nil {
		logger.Error(ctx, "Failed to check organization name", logfacade.Err(err))
		return bizerrors.Wrap(err, "Failed to check organization name")
	}
	if orgExists {
		logger.Warnf(ctx, "Organization name already exists: orgName=%s", orgSlug)
		return bizerrors.NewError(bizerrors.OrganizationAlreadyExists, orgSlug)
	}
	logger.Infof(ctx, "Organization name available: orgName=%s", orgSlug)
	return nil
}

// getOwnerRole retrieves the owner system role
func (s *CreateOrganizationService) getOwnerRole(ctx context.Context) (*domainPermission.Role, error) {
	logger := logfacade.GetLogger(ctx)
	ownerRole, err := s.roleRepo.GetRoleByNameAndOrg(ctx, "owner", "__SYSTEM__")
	if err != nil {
		logger.Error(ctx, "Failed to get owner system role", logfacade.Err(err))
		return nil, bizerrors.Wrap(err, "Failed to get owner role")
	}
	if ownerRole == nil {
		logger.Error(ctx, "Owner system role not found in database")
		return nil, bizerrors.NewError(bizerrors.SystemError, "Owner system role not found")
	}
	logger.Infof(ctx, "Owner role found: roleID=%d", ownerRole.ID)
	return ownerRole, nil
}

// createOrganizationInTransaction executes the organization creation in a transaction
func (s *CreateOrganizationService) createOrganizationInTransaction(
	ctx context.Context, orgSlug, displayName, userID string, roleID int,
) (*CreateOrganizationOutput, error) {
	logger := logfacade.GetLogger(ctx)
	var output *CreateOrganizationOutput

	err := s.txManager.WithTx(ctx, func(ctx context.Context, q dbgen.Querier) error {
		// Create organization
		org, txErr := organization.NewOrganization(orgSlug, displayName, userID)
		if txErr != nil {
			logger.Error(ctx, "Invalid organization data", logfacade.Err(txErr))
			return bizerrors.Wrap(txErr, "Invalid organization data")
		}
		orgRepository := repository.NewSqlOrganizationRepository(q)
		if txErr = orgRepository.Create(ctx, org); txErr != nil {
			if shared.IsDuplicateKeyError(txErr) {
				logger.Warnf(ctx, "Organization name already taken (concurrent request): orgName=%s", orgSlug)
				return bizerrors.NewError(bizerrors.OrganizationAlreadyExists, orgSlug)
			}
			logger.Error(ctx, "Failed to create organization", logfacade.Err(txErr))
			return bizerrors.Wrap(txErr, "Failed to create organization")
		}
		logger.Infof(ctx, "Organization created: orgName=%s", orgSlug)

		// Create membership
		membershipID := uuid.New().String()
		ms, txErr := membership.NewMembership(membershipID, userID, orgSlug)
		if txErr != nil {
			logger.Error(ctx, "Invalid membership data", logfacade.Err(txErr))
			return bizerrors.Wrap(txErr, "Invalid membership data")
		}
		membershipRepository := repository.NewSqlMembershipRepository(q)
		if txErr = membershipRepository.Create(ctx, ms); txErr != nil {
			logger.Error(ctx, "Failed to create membership", logfacade.Err(txErr))
			return bizerrors.Wrap(txErr, "Failed to create membership")
		}
		logger.Infof(ctx, "Membership created: membershipID=%s", membershipID)

		// Assign owner role
		userRoleRepository := repository.NewSqlCasbinUserRoleRepository(q)
		userRole := &domainPermission.UserRole{UserID: userID, RoleID: roleID, OrgName: orgSlug}
		if txErr = userRoleRepository.AssignRole(ctx, userRole); txErr != nil {
			logger.Error(ctx, "Failed to assign owner role", logfacade.Err(txErr))
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
		logger.Error(ctx, "Transaction failed", logfacade.Err(err))
		return nil, err
	}
	return output, nil
}

// validateOrgSlugFormat validates the organization slug format.
// Requirements:
// - 6-24 characters
// - Only lowercase letters, numbers, and underscores (no hyphens)
// - Must start with a letter
func validateOrgSlugFormat(slug string) error {
	if len(slug) < 6 || len(slug) > 24 {
		return bizerrors.NewError(bizerrors.ParamInvalid, fmt.Sprintf(
			"organization slug must be 6-24 characters (got %d)", len(slug)))
	}

	// Pattern: start with letter, only lowercase letters/numbers/underscores, no hyphens
	slugPattern := regexp.MustCompile(`^[a-z][a-z0-9_]*$`)
	if !slugPattern.MatchString(slug) {
		return bizerrors.NewError(
			bizerrors.ParamInvalid,
			"organization slug must start with lowercase letter and contain only letters, numbers, "+
				"and underscores (no hyphens)",
		)
	}

	return nil
}

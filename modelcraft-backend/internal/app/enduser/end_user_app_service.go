package enduser

import (
	"context"
	"database/sql"
	"modelcraft/internal/domain/shared"
	"modelcraft/pkg/bizerrors"
	"modelcraft/pkg/bizutils"

	domainenduser "modelcraft/internal/domain/enduser"

	infrrepo "modelcraft/internal/infrastructure/repository"
)

// PrivateDBManager 定义私有数据库连接管理器接口。
type PrivateDBManager interface {
	// Deprecated: GetOrInit still accepts projectSlug for backward compatibility,
	// but EndUser is now Org-scoped. The projectSlug parameter will be removed.
	GetOrInit(ctx context.Context, orgName, projectSlug string) (*sql.DB, error)
}

// EndUserManagementAppService 终端用户管理应用服务。
type EndUserManagementAppService struct {
	privateDBManager PrivateDBManager
	txManager        TxManager
}

// NewEndUserManagementAppService 创建终端用户管理应用服务。
func NewEndUserManagementAppService(
	privateDBManager PrivateDBManager,
	txManager TxManager,
) *EndUserManagementAppService {
	return &EndUserManagementAppService{
		privateDBManager: privateDBManager,
		txManager:        txManager,
	}
}

// CreateEndUser 创建终端用户（开发者管理侧）。
func (s *EndUserManagementAppService) CreateEndUser(
	ctx context.Context,
	cmd CreateEndUserCommand,
) (*CreateEndUserResult, error) {
	db, err := s.privateDBManager.GetOrInit(ctx, cmd.OrgName, cmd.ProjectSlug)
	if err != nil {
		return nil, s.convertDBError(ctx, err)
	}

	if err := domainenduser.ValidatePasswordStrength(cmd.Password); err != nil {
		return nil, bizerrors.NewErrorFromContext(ctx, bizerrors.EndUserParamInvalid, err.Error())
	}
	if err := domainenduser.ValidateUsername(cmd.Username); err != nil {
		return nil, bizerrors.NewErrorFromContext(ctx, bizerrors.EndUserParamInvalid, err.Error())
	}

	hashedPwd, err := domainenduser.NewHashedPasswordFromPlain(cmd.Password)
	if err != nil {
		return nil, bizerrors.Wrapf(err, "failed to hash password")
	}

	userID, err := bizutils.GenerateUUIDV7()
	if err != nil {
		return nil, bizerrors.Wrapf(err, "failed to generate end user id")
	}

	user, err := domainenduser.NewEndUser(userID, cmd.OrgName, cmd.Username, cmd.CreatedBy, hashedPwd)
	if err != nil {
		return nil, bizerrors.Wrapf(err, "failed to create end user entity")
	}

	repo := infrrepo.NewSqlEndUserRepository(db, cmd.OrgName, cmd.ProjectSlug)
	if err := repo.Save(ctx, user); err != nil {
		return nil, s.convertRepoError(ctx, err, cmd.Username)
	}

	return &CreateEndUserResult{
		ID:          user.ID,
		Username:    user.Username,
		IsForbidden: user.IsForbidden,
		CreatedBy:   user.CreatedBy,
		CreatedAt:   user.CreatedAt,
		UpdatedAt:   user.UpdatedAt,
	}, nil
}

// ListEndUsers 列出终端用户。
func (s *EndUserManagementAppService) ListEndUsers(
	ctx context.Context,
	cmd ListEndUsersCommand,
) (*ListEndUsersResult, error) {
	db, err := s.privateDBManager.GetOrInit(ctx, cmd.OrgName, cmd.ProjectSlug)
	if err != nil {
		return nil, s.convertDBError(ctx, err)
	}

	first := cmd.First
	if first <= 0 {
		first = 20
	}
	if first > 100 {
		first = 100
	}

	repo := infrrepo.NewSqlEndUserRepository(db, cmd.OrgName, cmd.ProjectSlug)
	users, totalCount, err := repo.ListWithTotal(ctx, domainenduser.ListEndUsersQuery{
		OrgName: cmd.OrgName,
		Search:  cmd.Search,
		First:   first,
		After:   cmd.After,
	})
	if err != nil {
		return nil, bizerrors.ConvertRepositoryError(ctx, err)
	}

	items := make([]*EndUserDTO, 0, len(users))
	for _, u := range users {
		items = append(items, s.toDTO(u))
	}

	var endCursor string
	hasNextPage := len(items) == first
	if len(items) > 0 {
		endCursor = items[len(items)-1].ID
	}

	return &ListEndUsersResult{
		Items:       items,
		TotalCount:  totalCount,
		HasNextPage: hasNextPage,
		EndCursor:   endCursor,
	}, nil
}

// GetEndUser gets a single end-user by ID.
func (s *EndUserManagementAppService) GetEndUser(
	ctx context.Context,
	cmd GetEndUserCommand,
) (*EndUserDTO, error) {
	db, err := s.privateDBManager.GetOrInit(ctx, cmd.OrgName, cmd.ProjectSlug)
	if err != nil {
		return nil, s.convertDBError(ctx, err)
	}

	repo := infrrepo.NewSqlEndUserRepository(db, cmd.OrgName, cmd.ProjectSlug)
	user, err := repo.GetByID(ctx, cmd.OrgName, cmd.UserID)
	if err != nil {
		return nil, bizerrors.ConvertRepositoryError(ctx, err)
	}
	if user == nil {
		return nil, bizerrors.NewErrorFromContext(ctx, bizerrors.EndUserNotFound, cmd.UserID)
	}

	return s.toDTO(user), nil
}

// UpdateEndUserStatus 更新终端用户状态。
func (s *EndUserManagementAppService) UpdateEndUserStatus(
	ctx context.Context,
	cmd UpdateEndUserStatusCommand,
) (*EndUserDTO, error) {
	db, err := s.privateDBManager.GetOrInit(ctx, cmd.OrgName, cmd.ProjectSlug)
	if err != nil {
		return nil, s.convertDBError(ctx, err)
	}

	repo := infrrepo.NewSqlEndUserRepository(db, cmd.OrgName, cmd.ProjectSlug)
	user, err := repo.GetByID(ctx, cmd.OrgName, cmd.UserID)
	if err != nil {
		return nil, bizerrors.ConvertRepositoryError(ctx, err)
	}
	if user == nil {
		return nil, bizerrors.NewErrorFromContext(ctx, bizerrors.EndUserNotFound, cmd.UserID)
	}
	if cmd.IsForbidden && user.IsBuiltin {
		return nil, bizerrors.NewErrorFromContext(ctx, ErrBuiltinUserCannotBeDisabled)
	}

	if cmd.IsForbidden {
		user.Disable()
	} else {
		user.Enable()
	}

	if err := repo.UpdateStatus(ctx, cmd.OrgName, cmd.UserID, user.IsForbidden); err != nil {
		return nil, bizerrors.ConvertRepositoryError(ctx, err)
	}

	updated, err := repo.GetByID(ctx, cmd.OrgName, cmd.UserID)
	if err != nil {
		return nil, bizerrors.ConvertRepositoryError(ctx, err)
	}
	if updated == nil {
		return nil, bizerrors.NewErrorFromContext(ctx, bizerrors.EndUserNotFound, cmd.UserID)
	}

	return s.toDTO(updated), nil
}

// DeleteEndUser 删除终端用户。
func (s *EndUserManagementAppService) DeleteEndUser(
	ctx context.Context,
	cmd DeleteEndUserCommand,
) error {
	db, err := s.privateDBManager.GetOrInit(ctx, cmd.OrgName, cmd.ProjectSlug)
	if err != nil {
		return s.convertDBError(ctx, err)
	}

	userRepo := infrrepo.NewSqlEndUserRepository(db, cmd.OrgName, cmd.ProjectSlug)
	user, err := userRepo.GetByID(ctx, cmd.OrgName, cmd.UserID)
	if err != nil {
		return bizerrors.ConvertRepositoryError(ctx, err)
	}
	if user == nil {
		return bizerrors.NewErrorFromContext(ctx, bizerrors.EndUserNotFound, cmd.UserID)
	}
	if user.IsBuiltin {
		return bizerrors.NewErrorFromContext(ctx, ErrBuiltinUserCannotBeDeleted)
	}

	err = s.txManager.WithTx(ctx, db, func(ctx context.Context, txDB SQLDBTX) error {
		txUserRepo := infrrepo.NewSqlEndUserRepository(txDB, cmd.OrgName, cmd.ProjectSlug)
		txSessionRepo := infrrepo.NewSqlEndUserSessionRepository(txDB, cmd.OrgName, cmd.ProjectSlug)

		if err := txSessionRepo.RevokeAllByUserID(ctx, cmd.UserID); err != nil {
			return bizerrors.ConvertRepositoryError(ctx, err)
		}
		if err := txUserRepo.Delete(ctx, cmd.OrgName, cmd.UserID); err != nil {
			return bizerrors.ConvertRepositoryError(ctx, err)
		}
		return nil
	})
	if err != nil {
		if _, ok := err.(*bizerrors.BusinessError); ok {
			return err
		}
		return bizerrors.ConvertRepositoryError(ctx, err)
	}

	return nil
}

// ResetEndUserPassword 重置终端用户密码（开发者管理侧）。
func (s *EndUserManagementAppService) ResetEndUserPassword(
	ctx context.Context,
	cmd ResetEndUserPasswordCommand,
) error {
	db, err := s.privateDBManager.GetOrInit(ctx, cmd.OrgName, "")
	if err != nil {
		return s.convertDBError(ctx, err)
	}

	if err := domainenduser.ValidatePasswordStrength(cmd.NewPassword); err != nil {
		return bizerrors.NewErrorFromContext(ctx, bizerrors.EndUserParamInvalid, err.Error())
	}

	repo := infrrepo.NewSqlEndUserRepository(db, cmd.OrgName, "")
	user, err := repo.GetByID(ctx, cmd.OrgName, cmd.UserID)
	if err != nil {
		return bizerrors.ConvertRepositoryError(ctx, err)
	}
	if user == nil {
		return bizerrors.NewErrorFromContext(ctx, bizerrors.EndUserNotFound, cmd.UserID)
	}
	if user.IsBuiltin {
		return bizerrors.NewErrorFromContext(ctx, ErrBuiltinUserCannotBeDisabled)
	}

	hashedPwd, err := domainenduser.NewHashedPasswordFromPlain(cmd.NewPassword)
	if err != nil {
		return bizerrors.Wrapf(err, "failed to hash password")
	}

	if err := repo.UpdatePassword(ctx, cmd.OrgName, cmd.UserID, hashedPwd); err != nil {
		return bizerrors.ConvertRepositoryError(ctx, err)
	}

	return nil
}

func (s *EndUserManagementAppService) toDTO(entity *domainenduser.EndUser) *EndUserDTO {
	if entity == nil {
		return nil
	}
	return &EndUserDTO{
		ID:          entity.ID,
		Username:    entity.Username,
		IsForbidden: entity.IsForbidden,
		IsBuiltin:   entity.IsBuiltin,
		CreatedBy:   entity.CreatedBy,
		CreatedAt:   entity.CreatedAt,
		UpdatedAt:   entity.UpdatedAt,
	}
}

// ListAccessibleProjects 获取指定终端用户在当前 Org 下可访问的项目列表。
func (s *EndUserManagementAppService) ListAccessibleProjects(
	ctx context.Context,
	orgName, userID string,
) ([]AccessibleProjectItem, error) {
	db, err := s.privateDBManager.GetOrInit(ctx, orgName, "")
	if err != nil {
		return nil, s.convertDBError(ctx, err)
	}

	repo := infrrepo.NewSqlEndUserRepository(db, orgName, "")
	projects, err := repo.ListAccessibleProjectsByRoleAssignment(ctx, orgName, userID)
	if err != nil {
		return nil, bizerrors.ConvertRepositoryError(ctx, err)
	}

	items := make([]AccessibleProjectItem, 0, len(projects))
	for _, p := range projects {
		items = append(items, AccessibleProjectItem{
			Slug:        p.ProjectSlug,
			Title:       p.ProjectTitle,
			Description: p.ProjectDescription,
			Status:      p.ProjectStatus,
			CreatedAt:   p.ProjectCreatedAt,
			UpdatedAt:   p.ProjectUpdatedAt,
		})
	}
	return items, nil
}

func (s *EndUserManagementAppService) convertDBError(ctx context.Context, err error) error {
	if shared.IsNotFoundError(err) {
		return bizerrors.NewErrorFromContext(ctx, bizerrors.EndUserClusterNotConfigured)
	}
	return bizerrors.ConvertRepositoryError(ctx, err)
}

func (s *EndUserManagementAppService) convertRepoError(ctx context.Context, err error, username string) error {
	if shared.IsRepoError(err, shared.ErrTypeDuplicatedKey) || shared.IsDuplicateKeyError(err) {
		return bizerrors.NewErrorFromContext(ctx, bizerrors.EndUserConflict, username)
	}
	return bizerrors.ConvertRepositoryError(ctx, err)
}

// NewPrivateDBManagerAdapter 将 infrastructure PrivateDBManager 适配为应用层接口。
func NewPrivateDBManagerAdapter(manager *infrrepo.PrivateDBManager) PrivateDBManager {
	return manager
}

// NewEndUserManagementAppServiceWithRepo creates a thin service wrapper backed
// directly by a repository — for unit-testing guards without a real DB.
func NewEndUserManagementAppServiceWithRepo(
	repo domainenduser.EndUserRepository,
) *EndUserManagementAppServiceWithRepoImpl {
	return &EndUserManagementAppServiceWithRepoImpl{repo: repo}
}

// EndUserManagementAppServiceWithRepoImpl is a test-helper service that exposes
// guard logic directly without requiring a PrivateDBManager.
type EndUserManagementAppServiceWithRepoImpl struct {
	repo domainenduser.EndUserRepository
}

// DeleteEndUserDirect deletes a user, enforcing the builtin guard.
func (s *EndUserManagementAppServiceWithRepoImpl) DeleteEndUserDirect(
	ctx context.Context, orgName, userID string,
) error {
	user, err := s.repo.GetByID(ctx, orgName, userID)
	if err != nil {
		return err
	}
	if user == nil {
		return bizerrors.NewError(bizerrors.NotFound, userID)
	}
	if user.IsBuiltin {
		return bizerrors.NewError(ErrBuiltinUserCannotBeDeleted)
	}
	return s.repo.Delete(ctx, orgName, userID)
}

// UpdateEndUserStatusDirect updates status, enforcing the builtin guard.
func (s *EndUserManagementAppServiceWithRepoImpl) UpdateEndUserStatusDirect(
	ctx context.Context, orgName, userID string, isForbidden bool,
) error {
	user, err := s.repo.GetByID(ctx, orgName, userID)
	if err != nil {
		return err
	}
	if user == nil {
		return bizerrors.NewError(bizerrors.NotFound, userID)
	}
	if isForbidden && user.IsBuiltin {
		return bizerrors.NewError(ErrBuiltinUserCannotBeDisabled)
	}
	return s.repo.UpdateStatus(ctx, orgName, userID, isForbidden)
}

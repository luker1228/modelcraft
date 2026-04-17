package enduser

import (
	"context"
	"database/sql"
	"time"

	domainenduser "modelcraft/internal/domain/enduser"
	"modelcraft/internal/domain/shared"
	infrrepo "modelcraft/internal/infrastructure/repository"
	"modelcraft/pkg/bizerrors"
	"modelcraft/pkg/bizutils"
)

// EndUserRepository 定义终端用户仓储接口。
type EndUserRepository interface {
	Save(ctx context.Context, user *EndUserEntity) error
	GetByID(ctx context.Context, id string) (*EndUserEntity, error)
	UpdateStatus(ctx context.Context, id string, isForbidden bool) error
	Delete(ctx context.Context, id string) error
	ListWithTotal(ctx context.Context, query ListEndUsersQuery) ([]*EndUserEntity, int64, error)
}

// EndUserSessionRepository 定义终端用户会话仓储接口。
type EndUserSessionRepository interface {
	RevokeAllByUserID(ctx context.Context, userID string) error
}

// PrivateDBManager 定义私有数据库连接管理器接口。
type PrivateDBManager interface {
	GetOrInit(ctx context.Context, orgName, projectSlug string) (PrivateDBConnection, error)
}

// PrivateDBConnection 私有数据库连接，提供仓储工厂。
type PrivateDBConnection interface {
	EndUserRepository() EndUserRepository
	SessionRepository() EndUserSessionRepository
}

// ListEndUsersQuery 列表查询参数。
type ListEndUsersQuery struct {
	Search string
	First  int
	After  string
}

// EndUserEntity 终端用户实体（应用层 DTO）。
type EndUserEntity struct {
	ID          string
	Username    string
	Password    string
	IsForbidden bool
	CreatedBy   string
	CreatedAt   string
	UpdatedAt   string
}

// EndUserManagementAppService 终端用户管理应用服务。
type EndUserManagementAppService struct {
	privateDBManager PrivateDBManager
}

// NewEndUserManagementAppService 创建终端用户管理应用服务。
func NewEndUserManagementAppService(
	privateDBManager PrivateDBManager,
) *EndUserManagementAppService {
	return &EndUserManagementAppService{privateDBManager: privateDBManager}
}

// CreateEndUser 创建终端用户（开发者管理侧）。
func (s *EndUserManagementAppService) CreateEndUser(
	ctx context.Context,
	cmd CreateEndUserCommand,
) (*CreateEndUserResult, error) {
	conn, err := s.privateDBManager.GetOrInit(ctx, cmd.OrgName, cmd.ProjectSlug)
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

	now := time.Now().UTC().Format(time.RFC3339)
	user := &EndUserEntity{
		ID:          userID,
		Username:    cmd.Username,
		Password:    hashedPwd.Hash,
		IsForbidden: false,
		CreatedBy:   cmd.CreatedBy,
		CreatedAt:   now,
		UpdatedAt:   now,
	}

	repo := conn.EndUserRepository()
	if err := repo.Save(ctx, user); err != nil {
		return nil, s.convertRepoError(ctx, err, cmd.Username)
	}

	return &CreateEndUserResult{
		ID:          user.ID,
		Username:    user.Username,
		IsForbidden: user.IsForbidden,
		CreatedBy:   user.CreatedBy,
	}, nil
}

// ListEndUsers 列出终端用户。
func (s *EndUserManagementAppService) ListEndUsers(
	ctx context.Context,
	cmd ListEndUsersCommand,
) (*ListEndUsersResult, error) {
	conn, err := s.privateDBManager.GetOrInit(ctx, cmd.OrgName, cmd.ProjectSlug)
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

	repo := conn.EndUserRepository()
	users, totalCount, err := repo.ListWithTotal(ctx, ListEndUsersQuery{
		Search: cmd.Search,
		First:  first,
		After:  cmd.After,
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

// UpdateEndUserStatus 更新终端用户状态。
func (s *EndUserManagementAppService) UpdateEndUserStatus(
	ctx context.Context,
	cmd UpdateEndUserStatusCommand,
) (*EndUserDTO, error) {
	conn, err := s.privateDBManager.GetOrInit(ctx, cmd.OrgName, cmd.ProjectSlug)
	if err != nil {
		return nil, s.convertDBError(ctx, err)
	}

	repo := conn.EndUserRepository()
	user, err := repo.GetByID(ctx, cmd.UserID)
	if err != nil {
		return nil, bizerrors.ConvertRepositoryError(ctx, err)
	}
	if user == nil {
		return nil, bizerrors.NewErrorFromContext(ctx, bizerrors.EndUserNotFound, cmd.UserID)
	}

	if err := repo.UpdateStatus(ctx, cmd.UserID, cmd.IsForbidden); err != nil {
		return nil, bizerrors.ConvertRepositoryError(ctx, err)
	}

	updated, err := repo.GetByID(ctx, cmd.UserID)
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
	conn, err := s.privateDBManager.GetOrInit(ctx, cmd.OrgName, cmd.ProjectSlug)
	if err != nil {
		return s.convertDBError(ctx, err)
	}

	userRepo := conn.EndUserRepository()
	sessionRepo := conn.SessionRepository()

	user, err := userRepo.GetByID(ctx, cmd.UserID)
	if err != nil {
		return bizerrors.ConvertRepositoryError(ctx, err)
	}
	if user == nil {
		return bizerrors.NewErrorFromContext(ctx, bizerrors.EndUserNotFound, cmd.UserID)
	}

	if err := sessionRepo.RevokeAllByUserID(ctx, cmd.UserID); err != nil {
		return bizerrors.ConvertRepositoryError(ctx, err)
	}
	if err := userRepo.Delete(ctx, cmd.UserID); err != nil {
		return bizerrors.ConvertRepositoryError(ctx, err)
	}

	return nil
}

func (s *EndUserManagementAppService) toDTO(entity *EndUserEntity) *EndUserDTO {
	if entity == nil {
		return nil
	}
	return &EndUserDTO{
		ID:          entity.ID,
		Username:    entity.Username,
		IsForbidden: entity.IsForbidden,
		CreatedBy:   entity.CreatedBy,
	}
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

// --- Infrastructure adapters (for wiring in routes.go) ---

type privateDBManagerAdapter struct {
	manager *infrrepo.PrivateDBManager
}

// NewPrivateDBManagerAdapter 将 infrastructure PrivateDBManager 适配为应用层接口。
func NewPrivateDBManagerAdapter(manager *infrrepo.PrivateDBManager) PrivateDBManager {
	return &privateDBManagerAdapter{manager: manager}
}

func (a *privateDBManagerAdapter) GetOrInit(
	ctx context.Context,
	orgName, projectSlug string,
) (PrivateDBConnection, error) {
	db, err := a.manager.GetOrInit(ctx, orgName, projectSlug)
	if err != nil {
		return nil, err
	}
	return &privateDBConnectionAdapter{db: db}, nil
}

type privateDBConnectionAdapter struct {
	db *sql.DB
}

func (c *privateDBConnectionAdapter) EndUserRepository() EndUserRepository {
	return &endUserRepositoryAdapter{repo: infrrepo.NewSqlEndUserRepository(c.db)}
}

func (c *privateDBConnectionAdapter) SessionRepository() EndUserSessionRepository {
	return &endUserSessionRepositoryAdapter{repo: infrrepo.NewSqlEndUserSessionRepository(c.db)}
}

type endUserRepositoryAdapter struct {
	repo *infrrepo.SqlEndUserRepository
}

func (a *endUserRepositoryAdapter) Save(ctx context.Context, user *EndUserEntity) error {
	row := &infrrepo.EndUser{
		ID:          user.ID,
		Username:    user.Username,
		Password:    user.Password,
		IsForbidden: user.IsForbidden,
		CreatedBy:   user.CreatedBy,
		CreatedAt:   parseRFC3339OrZero(user.CreatedAt),
		UpdatedAt:   parseRFC3339OrZero(user.UpdatedAt),
	}
	return a.repo.Save(ctx, row)
}

func (a *endUserRepositoryAdapter) GetByID(ctx context.Context, id string) (*EndUserEntity, error) {
	row, err := a.repo.GetByID(ctx, id)
	if err != nil || row == nil {
		return nil, err
	}
	return toEndUserEntity(row), nil
}

func (a *endUserRepositoryAdapter) UpdateStatus(ctx context.Context, id string, isForbidden bool) error {
	return a.repo.UpdateStatus(ctx, id, isForbidden)
}

func (a *endUserRepositoryAdapter) Delete(ctx context.Context, id string) error {
	return a.repo.Delete(ctx, id)
}

func (a *endUserRepositoryAdapter) ListWithTotal(
	ctx context.Context,
	query ListEndUsersQuery,
) ([]*EndUserEntity, int64, error) {
	rows, total, err := a.repo.ListWithTotal(ctx, infrrepo.ListEndUsersQuery{
		Search: query.Search,
		First:  query.First,
		After:  query.After,
	})
	if err != nil {
		return nil, 0, err
	}
	items := make([]*EndUserEntity, 0, len(rows))
	for _, row := range rows {
		items = append(items, toEndUserEntity(row))
	}
	return items, total, nil
}

type endUserSessionRepositoryAdapter struct {
	repo *infrrepo.SqlEndUserSessionRepository
}

func (a *endUserSessionRepositoryAdapter) RevokeAllByUserID(ctx context.Context, userID string) error {
	return a.repo.RevokeAllByUserID(ctx, userID)
}

func toEndUserEntity(row *infrrepo.EndUser) *EndUserEntity {
	if row == nil {
		return nil
	}
	return &EndUserEntity{
		ID:          row.ID,
		Username:    row.Username,
		Password:    row.Password,
		IsForbidden: row.IsForbidden,
		CreatedBy:   row.CreatedBy,
		CreatedAt:   row.CreatedAt.UTC().Format(time.RFC3339),
		UpdatedAt:   row.UpdatedAt.UTC().Format(time.RFC3339),
	}
}

func parseRFC3339OrZero(value string) time.Time {
	if value == "" {
		return time.Time{}
	}
	t, err := time.Parse(time.RFC3339, value)
	if err != nil {
		return time.Time{}
	}
	return t
}

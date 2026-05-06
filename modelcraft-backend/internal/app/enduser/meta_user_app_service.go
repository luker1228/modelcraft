package enduser

import (
	"context"
	"database/sql"
	"fmt"
	"modelcraft/pkg/bizerrors"
	"modelcraft/pkg/ctxutils"
	"strings"
	"time"

	infrrepo "modelcraft/internal/infrastructure/repository"
)

// MetaUserAppService 提供 runtime meta/user 路由专属的受限查询能力。
// 所有方法均从 context 注入 orgName，禁止客户端传入租户字段。
type MetaUserAppService struct {
	privateDBManager PrivateDBManager
}

// NewMetaUserAppService 创建 MetaUserAppService。
func NewMetaUserAppService(privateDBManager PrivateDBManager) *MetaUserAppService {
	return &MetaUserAppService{privateDBManager: privateDBManager}
}

// GetMe 返回当前认证用户的 meta/user 资料。
// userID 来自 JWT 中间件注入（ctxutils.GetUserIDFromContext），orgName/projectSlug 来自 URL 参数中间件注入。
func (s *MetaUserAppService) GetMe(ctx context.Context) (*MetaUserDTO, error) {
	orgName, _ := ctxutils.GetOrgNameFromContext(ctx)
	projectSlug, _ := ctxutils.GetProjectSlugFromContext(ctx)
	userID, _ := ctxutils.GetUserIDFromContext(ctx)
	if orgName == "" || userID == "" {
		return nil, bizerrors.NewErrorFromContext(ctx, bizerrors.EndUserNotFound, "missing org or user context")
	}

	db, err := s.privateDBManager.GetOrInit(ctx, orgName, projectSlug)
	if err != nil {
		return nil, bizerrors.ConvertRepositoryError(ctx, err)
	}

	repo := infrrepo.NewSqlEndUserRepository(db, orgName, projectSlug)
	user, err := repo.GetByID(ctx, orgName, userID)
	if err != nil {
		return nil, bizerrors.ConvertRepositoryError(ctx, err)
	}
	if user == nil {
		return nil, bizerrors.NewErrorFromContext(ctx, bizerrors.EndUserNotFound, userID)
	}

	return metaUserToDTO(user.ID, user.Username, user.CreatedAt), nil
}

// FindOne 按唯一条件（id 或 username）查询单个用户。
// 至少提供 id 或 username 之一；orgName 和 ProjectSlug 来自 context。
// 未找到用户时返回 (nil, nil)，符合仓储层契约。
func (s *MetaUserAppService) FindOne(ctx context.Context, cmd MetaUserFindOneCommand) (*MetaUserDTO, error) {
	if cmd.OrgName == "" {
		return nil, bizerrors.NewErrorFromContext(ctx, bizerrors.EndUserParamInvalid, "orgName is required")
	}
	if cmd.ID == "" && cmd.Username == "" {
		return nil, bizerrors.NewErrorFromContext(ctx, bizerrors.EndUserParamInvalid, "id or username is required")
	}

	db, err := s.privateDBManager.GetOrInit(ctx, cmd.OrgName, cmd.ProjectSlug)
	if err != nil {
		return nil, bizerrors.ConvertRepositoryError(ctx, err)
	}

	repo := infrrepo.NewSqlEndUserRepository(db, cmd.OrgName, cmd.ProjectSlug)

	if cmd.ID != "" {
		user, err := repo.GetByID(ctx, cmd.OrgName, cmd.ID)
		if err != nil {
			return nil, bizerrors.ConvertRepositoryError(ctx, err)
		}
		if user == nil {
			return nil, nil //nolint:nilnil // per contract: not found returns (nil, nil) at repo layer
		}
		return metaUserToDTO(user.ID, user.Username, user.CreatedAt), nil
	}

	user, err := repo.GetByUsername(ctx, cmd.OrgName, cmd.Username)
	if err != nil {
		return nil, bizerrors.ConvertRepositoryError(ctx, err)
	}
	if user == nil {
		return nil, nil //nolint:nilnil // per contract: not found returns (nil, nil) at repo layer
	}
	return metaUserToDTO(user.ID, user.Username, user.CreatedAt), nil
}

// FindMany 执行受限列表查询。
// 分页：take 默认 20，最大 50；skip 默认 0，最大 1000。
// 过滤字段白名单：id、username、createdAt。
// 排序字段白名单：createdAt。
func (s *MetaUserAppService) FindMany( //nolint:lll
	ctx context.Context,
	cmd MetaUserFindManyCommand,
) (*MetaUserFindManyResult, error) {
	if cmd.OrgName == "" {
		return nil, bizerrors.NewErrorFromContext(ctx, bizerrors.EndUserParamInvalid, "orgName is required")
	}

	// 分页边界校验
	take := cmd.Take
	if take <= 0 {
		take = 20
	}
	if take > 50 {
		return nil, bizerrors.NewErrorFromContext(ctx, bizerrors.EndUserParamInvalid,
			fmt.Sprintf("take must be <= 50, got %d", cmd.Take))
	}
	skip := cmd.Skip
	if skip < 0 {
		return nil, bizerrors.NewErrorFromContext(ctx, bizerrors.EndUserParamInvalid,
			fmt.Sprintf("skip must be >= 0, got %d", cmd.Skip))
	}
	if skip > 1000 {
		return nil, bizerrors.NewErrorFromContext(ctx, bizerrors.EndUserParamInvalid,
			fmt.Sprintf("skip must be <= 1000, got %d", cmd.Skip))
	}

	// 排序字段白名单校验（白名单：createdAt）
	orderDir := "ASC"
	for _, ob := range cmd.OrderBy {
		if ob.CreatedAt == nil {
			return nil, bizerrors.NewErrorFromContext(ctx, bizerrors.EndUserParamInvalid,
				"unsupported orderBy field: only createdAt is allowed")
		}
		dir := strings.ToLower(*ob.CreatedAt)
		if dir != "asc" && dir != "desc" {
			return nil, bizerrors.NewErrorFromContext(ctx, bizerrors.EndUserParamInvalid,
				fmt.Sprintf("invalid sort direction: %s (must be asc or desc)", *ob.CreatedAt))
		}
		orderDir = strings.ToUpper(dir)
	}

	db, err := s.privateDBManager.GetOrInit(ctx, cmd.OrgName, cmd.ProjectSlug)
	if err != nil {
		return nil, bizerrors.ConvertRepositoryError(ctx, err)
	}

	items, err := s.runFindMany(ctx, db, cmd.OrgName, cmd.Where, orderDir, skip, take)
	if err != nil {
		return nil, err
	}
	return &MetaUserFindManyResult{Items: items}, nil
}

// runFindMany 执行实际的数据库查询，支持 where/orderBy/skip/limit。
// orderDir 只接受白名单值 "ASC"/"DESC"（由 FindMany 校验保证）。
func (s *MetaUserAppService) runFindMany(
	ctx context.Context,
	db *sql.DB,
	orgName string,
	where *MetaUserFindManyFilter,
	orderDir string,
	skip, take int,
) ([]*MetaUserDTO, error) {
	args := []any{orgName}
	conditions := []string{"org_name = ?"}

	if where != nil {
		conditions, args = applyMetaUserFilter(conditions, args, where)
	}

	// orderDir 已经过白名单校验（"ASC"/"DESC"），可以安全拼接到 SQL。
	//nolint:gosec // orderDir is whitelist-validated to "ASC" or "DESC" before reaching this point
	query := "SELECT id, username, created_at FROM end_user_users WHERE " +
		strings.Join(conditions, " AND ") +
		" ORDER BY created_at " + orderDir + " LIMIT ? OFFSET ?"

	args = append(args, take, skip)

	rows, err := db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	dtos := make([]*MetaUserDTO, 0, take)
	for rows.Next() {
		var (
			id        string
			username  string
			createdAt time.Time
		)
		if err := rows.Scan(&id, &username, &createdAt); err != nil {
			return nil, err
		}
		dtos = append(dtos, metaUserToDTO(id, username, createdAt))
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return dtos, nil
}

// applyMetaUserFilter 将过滤条件追加到 conditions/args（白名单字段：id, username, createdAt）。
func applyMetaUserFilter(conditions []string, args []any, where *MetaUserFindManyFilter) ([]string, []any) {
	if where.IDEq != nil {
		conditions = append(conditions, "id = ?")
		args = append(args, *where.IDEq)
	}
	if len(where.IDIn) > 0 {
		placeholders := make([]string, len(where.IDIn))
		for i, v := range where.IDIn {
			placeholders[i] = "?"
			args = append(args, v)
		}
		conditions = append(conditions, "id IN ("+strings.Join(placeholders, ",")+")")
	}
	if where.UsernameEq != nil {
		conditions = append(conditions, "username = ?")
		args = append(args, *where.UsernameEq)
	}
	if where.UsernameContains != nil {
		conditions = append(conditions, "username LIKE CONCAT('%', ?, '%')")
		args = append(args, *where.UsernameContains)
	}
	if where.UsernameStartsWith != nil {
		conditions = append(conditions, "username LIKE CONCAT(?, '%')")
		args = append(args, *where.UsernameStartsWith)
	}
	if len(where.UsernameIn) > 0 {
		placeholders := make([]string, len(where.UsernameIn))
		for i, v := range where.UsernameIn {
			placeholders[i] = "?"
			args = append(args, v)
		}
		conditions = append(conditions, "username IN ("+strings.Join(placeholders, ",")+")")
	}
	if where.CreatedAtEq != nil {
		conditions = append(conditions, "created_at = ?")
		args = append(args, *where.CreatedAtEq)
	}
	if where.CreatedAtGte != nil {
		conditions = append(conditions, "created_at >= ?")
		args = append(args, *where.CreatedAtGte)
	}
	if where.CreatedAtLte != nil {
		conditions = append(conditions, "created_at <= ?")
		args = append(args, *where.CreatedAtLte)
	}
	return conditions, args
}

func metaUserToDTO(id, username string, createdAt time.Time) *MetaUserDTO {
	return &MetaUserDTO{
		ID:        id,
		Username:  username,
		CreatedAt: createdAt,
	}
}

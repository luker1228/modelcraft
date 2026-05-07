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

// MetaUserAppService 提供 meta/user 路由专属的受限查询能力。
// end_user_users 存储在系统主库，直接使用系统 *sql.DB，不经过用户配置的 cluster。
type MetaUserAppService struct {
	db *sql.DB
}

// NewMetaUserAppService 创建 MetaUserAppService。
// db 应为系统主库连接（repoFactory.SqlDB）。
func NewMetaUserAppService(db *sql.DB) *MetaUserAppService {
	return &MetaUserAppService{db: db}
}

// GetMe 返回当前认证用户的资料（仅 EndUser 可调用）。
func (s *MetaUserAppService) GetMe(ctx context.Context) (*MetaUserDTO, error) {
	orgName, _ := ctxutils.GetOrgNameFromContext(ctx)
	userID, _ := ctxutils.GetUserIDFromContext(ctx)
	if orgName == "" || userID == "" {
		return nil, bizerrors.NewErrorFromContext(ctx, bizerrors.EndUserNotFound, "missing org or user context")
	}

	repo := infrrepo.NewSqlEndUserRepository(s.db, orgName, "")
	user, err := repo.GetByID(ctx, orgName, userID)
	if err != nil {
		return nil, bizerrors.ConvertRepositoryError(ctx, err)
	}
	if user == nil {
		return nil, bizerrors.NewErrorFromContext(ctx, bizerrors.EndUserNotFound, userID)
	}
	if !user.IsActive() {
		return nil, bizerrors.NewErrorFromContext(ctx, bizerrors.EndUserAccountDisabled)
	}
	return metaUserToDTO(user.ID, user.Username, user.CreatedAt), nil
}

// FindOne 按唯一条件（id 或 username）查询单个用户。
func (s *MetaUserAppService) FindOne(ctx context.Context, cmd MetaUserFindOneCommand) (*MetaUserDTO, error) {
	if cmd.OrgName == "" {
		return nil, bizerrors.NewErrorFromContext(ctx, bizerrors.EndUserParamInvalid, "orgName is required")
	}
	if cmd.ID == "" && cmd.Username == "" {
		return nil, bizerrors.NewErrorFromContext(ctx, bizerrors.EndUserParamInvalid, "id or username is required")
	}

	repo := infrrepo.NewSqlEndUserRepository(s.db, cmd.OrgName, "")

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
func (s *MetaUserAppService) FindMany(
	ctx context.Context,
	cmd MetaUserFindManyCommand,
) (*MetaUserFindManyResult, error) {
	if cmd.OrgName == "" {
		return nil, bizerrors.NewErrorFromContext(ctx, bizerrors.EndUserParamInvalid, "orgName is required")
	}

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

	items, err := s.runFindMany(ctx, cmd.OrgName, cmd.Where, orderDir, skip, take)
	if err != nil {
		return nil, err
	}
	return &MetaUserFindManyResult{Items: items}, nil
}

func (s *MetaUserAppService) runFindMany(
	ctx context.Context,
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

	//nolint:gosec // orderDir is whitelist-validated to "ASC" or "DESC"
	query := "SELECT id, username, created_at FROM end_user_users WHERE " +
		strings.Join(conditions, " AND ") +
		" ORDER BY created_at " + orderDir + " LIMIT ? OFFSET ?"

	args = append(args, take, skip)

	rows, err := s.db.QueryContext(ctx, query, args...)
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

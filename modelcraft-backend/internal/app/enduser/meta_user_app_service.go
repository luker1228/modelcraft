package enduser

import (
	"context"
	"database/sql"
	"encoding/base64"
	"fmt"
	"modelcraft/pkg/bizerrors"
	"modelcraft/pkg/ctxutils"
	"strconv"
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

// FindMany 执行受限列表查询（cursor 分页）。
// 排序固定为 created_at DESC, id DESC（最新优先）。
// first 默认 20，最大 50；After 为空字符串表示第一页。
func (s *MetaUserAppService) FindMany(
	ctx context.Context,
	cmd MetaUserFindManyCommand,
) (*MetaUserFindManyResult, error) {
	if cmd.OrgName == "" {
		return nil, bizerrors.NewErrorFromContext(ctx, bizerrors.EndUserParamInvalid, "orgName is required")
	}

	first := cmd.First
	if first <= 0 {
		first = 20
	}
	if first > 50 {
		return nil, bizerrors.NewErrorFromContext(ctx, bizerrors.EndUserParamInvalid,
			fmt.Sprintf("first must be <= 50, got %d", cmd.First))
	}

	var (
		cursorTime time.Time
		cursorID   string
	)
	if cmd.After != "" {
		var err error
		cursorTime, cursorID, err = decodeCursor(cmd.After)
		if err != nil {
			return nil, bizerrors.NewErrorFromContext(ctx, bizerrors.EndUserParamInvalid,
				fmt.Sprintf("invalid cursor: %s", cmd.After))
		}
	}

	// 多取一条，用于判断 hasMore
	items, err := s.runFindMany(ctx, cmd.OrgName, cmd.Where, cursorTime, cursorID, first+1)
	if err != nil {
		return nil, err
	}

	hasMore := len(items) > first
	if hasMore {
		items = items[:first]
	}

	var nextCursor string
	if hasMore && len(items) > 0 {
		last := items[len(items)-1]
		nextCursor = encodeCursor(last.CreatedAt, last.ID)
	}

	return &MetaUserFindManyResult{
		Items:      items,
		NextCursor: nextCursor,
		HasMore:    hasMore,
	}, nil
}

func (s *MetaUserAppService) runFindMany(
	ctx context.Context,
	orgName string,
	where *MetaUserFindManyFilter,
	cursorTime time.Time,
	cursorID string,
	limit int,
) ([]*MetaUserDTO, error) {
	args := []any{orgName}
	conditions := []string{"org_name = ?"}

	if where != nil {
		conditions, args = applyMetaUserFilter(conditions, args, where)
	}

	// cursor 条件：(created_at, id) < (cursorTime, cursorID)，DESC 方向向前翻页
	if !cursorTime.IsZero() && cursorID != "" {
		conditions = append(conditions, "(created_at < ? OR (created_at = ? AND id < ?))")
		args = append(args, cursorTime, cursorTime, cursorID)
	}

	//nolint:gosec // conditions 仅由白名单字段拼接，无用户输入直接进入 SQL 结构
	query := "SELECT id, username, created_at FROM end_user_users WHERE " +
		strings.Join(conditions, " AND ") +
		" ORDER BY created_at DESC, id DESC LIMIT ?"

	args = append(args, limit)

	rows, err := s.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	dtos := make([]*MetaUserDTO, 0, limit)
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

// encodeCursor 将 created_at + id 编码为 Base64 cursor 字符串。
// 格式：Base64("<unixMilli>|<id>")，例如 "1716217200000|abc-uuid"。
func encodeCursor(t time.Time, id string) string {
	raw := fmt.Sprintf("%d|%s", t.UnixMilli(), id)
	return base64.StdEncoding.EncodeToString([]byte(raw))
}

// decodeCursor 解码 Base64 cursor 字符串，返回 created_at 时间和 id。
func decodeCursor(cursor string) (time.Time, string, error) {
	b, err := base64.StdEncoding.DecodeString(cursor)
	if err != nil {
		return time.Time{}, "", fmt.Errorf("base64 decode failed: %w", err)
	}
	parts := strings.SplitN(string(b), "|", 2)
	if len(parts) != 2 {
		return time.Time{}, "", fmt.Errorf("invalid cursor format")
	}
	ms, err := strconv.ParseInt(parts[0], 10, 64)
	if err != nil {
		return time.Time{}, "", fmt.Errorf("invalid cursor timestamp: %w", err)
	}
	t := time.UnixMilli(ms).UTC()
	return t, parts[1], nil
}

func metaUserToDTO(id, username string, createdAt time.Time) *MetaUserDTO {
	return &MetaUserDTO{
		ID:        id,
		Username:  username,
		CreatedAt: createdAt,
	}
}

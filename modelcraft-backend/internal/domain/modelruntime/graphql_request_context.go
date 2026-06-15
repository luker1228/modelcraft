package modelruntime

import (
	"context"

	"github.com/graph-gophers/dataloader/v7"
)

// graphqlRequestContext 持有单次 GraphQL 请求的可变状态。
//
// 生命周期：与一次 graphql.Do 调用相同，通过 context.WithValue 注入。
// 这使得 graphql.Schema（类型结构）可以被安全缓存和跨请求复用，
// 而请求级状态（DB 连接、dataloader 实例）始终隔离。
type graphqlRequestContext struct {
	// ClientRepo 是当前请求使用的客户端数据库连接。
	ClientRepo ClientDatabaseRepository
	// relationLoaders 是 per-(tableName/referenceKey) 的 dataloader 实例 map。
	// 懒初始化：首次访问某个关系时创建，同一请求内复用以聚合批量查询。
	relationLoaders map[string]*dataloader.Loader[string, map[string]any]
	// OrgName 是请求的组织名称
	OrgName string
	// ProjectSlug 是请求的项目标识
	ProjectSlug string
	// CurrentEndUserID 是当前请求的 EndUser ID（从 JWT 提取）。
	// 为空字符串表示请求来自 tenant admin（无 EndUser 身份）。
	CurrentEndUserID string
	// EndUserAdminID 是当前请求用于自动填充 END_USER_REF 字段的默认 owner。
	// 对 tenant admin 请求，它来自标准 user identity。
	EndUserAdminID string
	// EndUserPerms 是当前 EndUser 在本次请求目标 model 上的权限快照。
	// nil 表示请求来自 tenant admin，跳过所有权限检查。
	EndUserPerms *ResolvedModelPermissions
}

// graphqlRequestContextKey 是 graphqlRequestContext 在 context 中的 key 类型。
type graphqlRequestContextKey struct{}

// newGraphqlRequestContext 创建一个新的请求级上下文。
func newGraphqlRequestContext(
	clientRepo ClientDatabaseRepository,
	orgName, projectSlug, currentEndUserID, endUserAdminID string,
	endUserPerms *ResolvedModelPermissions,
) *graphqlRequestContext {
	return &graphqlRequestContext{
		ClientRepo:       clientRepo,
		relationLoaders:  make(map[string]*dataloader.Loader[string, map[string]any]),
		OrgName:          orgName,
		ProjectSlug:      projectSlug,
		CurrentEndUserID: currentEndUserID,
		EndUserAdminID:   endUserAdminID,
		EndUserPerms:     endUserPerms,
	}
}

// WithGraphqlRequestContext 将请求级上下文注入 context，返回新 context。
// 由 App 层在每次 graphql.Do 前调用，确保所有 resolver 闭包均可通过 p.Context 取到。
func WithGraphqlRequestContext(
	ctx context.Context,
	clientRepo ClientDatabaseRepository,
	orgName, projectSlug, currentEndUserID, endUserAdminID string,
	endUserPerms *ResolvedModelPermissions,
) context.Context {
	rctx := newGraphqlRequestContext(clientRepo, orgName, projectSlug, currentEndUserID, endUserAdminID, endUserPerms)
	return context.WithValue(ctx, graphqlRequestContextKey{}, rctx)
}

// getGraphqlRequestContext 从 context 中取出请求级上下文。
// 若不存在（未经 withGraphqlRequestContext 注入），返回 (nil, false)。
func getGraphqlRequestContext(ctx context.Context) (*graphqlRequestContext, bool) {
	rctx, ok := ctx.Value(graphqlRequestContextKey{}).(*graphqlRequestContext)
	return rctx, ok
}

// getOrCreateLoader 懒初始化并返回指定关系的 dataloader 实例。
//
// key = "tableName/referenceKey"，同一请求内相同 key 复用同一 loader，
// 保证 graphql-go 广度优先执行时同层所有 Load() 调用可被聚合为一条 IN 查询。
func (rctx *graphqlRequestContext) getOrCreateLoader(
	tableName, referenceKey string,
) *dataloader.Loader[string, map[string]any] {
	key := tableName + "/" + referenceKey
	if l, ok := rctx.relationLoaders[key]; ok {
		return l
	}
	l := newRelationBatchLoader(rctx.ClientRepo, tableName, referenceKey)
	rctx.relationLoaders[key] = l
	return l
}

// resolveEndUserOwnerID returns the user ID to use for END_USER_REF field injection.
// Priority: CurrentEndUserID (from end-user JWT) > EndUserAdminID (from authenticated user context).
// Returns empty string if neither is available.
func (rctx *graphqlRequestContext) resolveEndUserOwnerID() string {
	if rctx.CurrentEndUserID != "" {
		return rctx.CurrentEndUserID
	}
	return rctx.EndUserAdminID
}

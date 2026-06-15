package modelruntime

import (
	"context"
	"encoding/json"
	"fmt"
	"modelcraft/internal/domain/modeldesign"
	"modelcraft/internal/domain/modelruntime"
	"modelcraft/internal/domain/shared"
	"modelcraft/internal/infrastructure/database/dml"
	"modelcraft/internal/infrastructure/repository"
	"modelcraft/internal/interfaces/http/middleware"
	"modelcraft/pkg/bizerrors"
	"modelcraft/pkg/bizutils"
	"modelcraft/pkg/ctxutils"
	"modelcraft/pkg/logfacade"
	"modelcraft/pkg/requestcontext"

	"github.com/graphql-go/graphql"
)

// GraphqlAppService graphql应用服务
type GraphqlAppService struct {
	modelRepo            modelruntime.ModelRepository
	graphqlSchemaManager *modelruntime.GraphqlSchemaManager
	permService          modelruntime.EndUserPermissionService
	snapshotBuilder      *RLSSnapshotBuilder
}

// GetSchema 获取或构建 GraphQL Schema。
// Schema 仅包含类型结构，不持有请求级状态（clientRepo、dataloader），可安全缓存。
func (s *GraphqlAppService) GetSchema(ctx context.Context, orgName string, modelLocator *modeldesign.ModelLocator,
) (*graphql.Schema, error) {
	logger := logfacade.GetLogger(ctx)
	gschema, err := s.graphqlSchemaManager.GetByName(ctx, modelLocator)
	if err != nil && !shared.IsNotFoundError(err) {
		return nil, err
	}
	if gschema != nil {
		return gschema, nil
	}

	model, err := s.modelRepo.GetByName(ctx, modelLocator)
	if err != nil {
		logger.Errorf(ctx, "get model fail: %v", err)
		if shared.IsNotFoundError(err) {
			return nil, bizerrors.NewError(bizerrors.ModelNotFound, modelLocator.GetFullPath())
		}
		return nil, fmt.Errorf("获取模型失败 %s", modelLocator.GetFullPath())
	}
	if model == nil {
		return nil, bizerrors.NewError(bizerrors.ModelNotFound, modelLocator.GetFullPath())
	}

	logger.Infof(ctx, "model=%s", bizutils.MarshalToStringIgnoreErr(model))
	gschema, err = s.graphqlSchemaManager.NewSchemaFrom(ctx, model)
	if err != nil {
		logger.Errorf(ctx, "generate schema fail: %v", err)
		return nil, fmt.Errorf("生成schema失败 %s", modelLocator.GetFullPath())
	}
	s.graphqlSchemaManager.StoreSchema(ctx, modelLocator, gschema)
	return gschema, nil
}

// Execute 执行graphql查询
func (s *GraphqlAppService) Execute(ctx context.Context, orgName, projectSlug, name, databaseName string,
	cmd ExecuteGraphQLCommand,
) (*graphql.Result, error) {
	logger := logfacade.GetLogger(ctx)

	// Inject request metadata into context
	ctx = requestcontext.WithMetadata(ctx)
	modelLocator, err := modeldesign.NewModelLocator(orgName, projectSlug, databaseName, name)
	if err != nil {
		return nil, err
	}

	// Load model once — modelID needed by snapshot builder and permission resolution
	model, err := s.modelRepo.GetByName(ctx, modelLocator)
	if err != nil {
		logger.Errorf(ctx, "get model fail: %v", err)
		if shared.IsNotFoundError(err) {
			return nil, bizerrors.NewError(bizerrors.ModelNotFound, modelLocator.GetFullPath())
		}
		return nil, fmt.Errorf("获取模型失败 %s", modelLocator.GetFullPath())
	}
	modelID := model.ID
	if modelID == "" {
		modelID = model.Name
	}

	gschema, err := s.GetSchema(ctx, orgName, modelLocator)
	if err != nil {
		return nil, err
	}

	// 创建请求级 DB 连接
	clientSqlDB, err := repository.DefaultClusterManager.GetConnectionWithDatabase(
		ctx, orgName, modelLocator.ProjectSlug, modelLocator.DatabaseName,
	)
	if err != nil {
		logger.Errorf(ctx, "get client sql db fail: %v", err)
		return nil, fmt.Errorf("获取客户端数据库失败 %s", databaseName)
	}
	// Wrap with RLS intercept layer
	clientRepo := dml.NewRLSInterceptDB(dml.NewClientDB(clientSqlDB))

	// 提取 endUserID
	endUserID := ""
	if ctxutils.IsEndUser(ctx) && !ctxutils.GetIsAdminFromContext(ctx) {
		if uid, err := ctxutils.GetUserIDFromContext(ctx); err == nil {
			endUserID = uid
		}
	}

	// 解析 end-user 权限快照
	var endUserPerms *modelruntime.ResolvedModelPermissions
	if endUserID != "" {
		endUserPerms, err = s.permService.Resolve(ctx, orgName, projectSlug, endUserID, modelID)
		if err != nil {
			return nil, err
		}
	}

	// Build RLS snapshot at entry
	identity := middleware.GetEndUserIdentity(ctx)
	isDeveloper := identity == nil || identity.IsDeveloper()
	userCtx := middleware.GetUserContext(ctx)
	snap, err := s.snapshotBuilder.Build(ctx, orgName, projectSlug, modelID, isDeveloper, userCtx)
	if err != nil {
		return nil, err
	}
	if snap != nil && snap.DenyAll {
		return nil, bizerrors.NewError(bizerrors.PermissionDenied, "RLS: no matching policy")
	}
	if snap != nil {
		ctx = modelruntime.WithRLSSnapshot(ctx, snap)
	}

	endUserAdminID, _ := ctxutils.GetUserIDFromContext(ctx)
	reqCtx := modelruntime.WithGraphqlRequestContext(
		ctx, clientRepo, orgName, projectSlug, endUserID, endUserAdminID, endUserPerms,
	)

	// 执行GraphQL查询
	result := graphql.Do(graphql.Params{
		Schema:         *gschema,
		RequestString:  cmd.Query,
		VariableValues: cmd.Variables,
		Context:        reqCtx,
	})
	marshal, err := json.Marshal(result)
	if err != nil {
		return nil, err
	}
	logger.Infof(ctx, "result=%+v", string(marshal))

	return result, nil
}

// NewGraphqlAppService 创建graphql应用服务
func NewGraphqlAppService(
	modelRepo modelruntime.ModelRepository,
	lfkRepo modeldesign.LogicalForeignKeyRepository,
	permService modelruntime.EndUserPermissionService,
	snapshotBuilder *RLSSnapshotBuilder,
) *GraphqlAppService {
	schemaManager := modelruntime.NewGraphqlSchemaManager(modelRepo, lfkRepo)
	return &GraphqlAppService{
		modelRepo:            modelRepo,
		graphqlSchemaManager: schemaManager,
		permService:          permService,
		snapshotBuilder:      snapshotBuilder,
	}
}

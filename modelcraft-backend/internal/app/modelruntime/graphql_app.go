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
	permResolver         *PolicyPermissionResolver
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

	var rlsCtx *modelruntime.RLSContext
	userCtx := middleware.GetUserContext(ctx)
	if userCtx != nil && userCtx.UseAdmin {
		rlsCtx = &modelruntime.RLSContext{
			IsAdmin: true,
		}
	} else {
		rlsCtx, ctx, err = s.resolveNonAdminRLS(ctx, orgName, projectSlug, modelID)
		if err != nil {
			return nil, err
		}
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

	reqCtx := modelruntime.WithGraphqlRequestContext(
		ctx, clientRepo, orgName, projectSlug, rlsCtx,
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
	permResolver *PolicyPermissionResolver,
	snapshotBuilder *RLSSnapshotBuilder,
) *GraphqlAppService {
	schemaManager := modelruntime.NewGraphqlSchemaManager(modelRepo, lfkRepo)
	return &GraphqlAppService{
		modelRepo:            modelRepo,
		graphqlSchemaManager: schemaManager,
		permResolver:         permResolver,
		snapshotBuilder:      snapshotBuilder,
	}
}

func (s *GraphqlAppService) buildRLSContext(
	ctx context.Context,
	orgName, projectSlug, modelID string,
) (*modelruntime.RLSContext, error) {
	logger := logfacade.GetLogger(ctx)

	userCtx := middleware.GetUserContext(ctx)
	rlsCtx := &modelruntime.RLSContext{
		IsAdmin:     userCtx != nil && userCtx.UseAdmin,
		UserContext: userCtx,
	}
	if rlsCtx.IsAdmin {
		return rlsCtx, nil
	}

	endUserID, err := ctxutils.GetEndUserIDFromContext(ctx)
	if err != nil {
		return nil, err
	}
	rlsCtx.EndUserID = endUserID
	rlsCtx.Permissions, err = s.permResolver.ResolveFromV2Policy(
		ctx, orgName, projectSlug, modelID, rlsCtx.UserContext.Roles,
	)
	if err != nil {
		return nil, err
	}
	if rlsCtx.Permissions != nil && rlsCtx.Permissions.IsEmpty() {
		rlsCtx.FastFail = true
		rlsCtx.FastFailReason = "permissions: no permissions granted"
		logger.Debugf(ctx, "RLS fast-fail: model=%s reason=%s", modelID, rlsCtx.FastFailReason)
		return rlsCtx, nil
	}

	snap, err := s.snapshotBuilder.Build(ctx, orgName, projectSlug, modelID, rlsCtx.UserContext, rlsCtx.Permissions)
	if err != nil {
		return nil, err
	}
	rlsCtx.Snapshot = snap
	logger.Infof(ctx, "RLS context built: model=%s endUser=%s policies=%d select=%v update=%v delete=%v create=%v",
		modelID, endUserID, len(rlsCtx.Permissions.Policies),
		!snap.NoSelectPolicy, !snap.NoUpdatePolicy, !snap.NoDeletePolicy, !snap.NoCreatePolicy)

	return rlsCtx, nil
}

// resolveNonAdminRLS builds the RLS context for a non-admin user and attaches
// the snapshot to ctx when present. Returns a FastFail error as PermissionDenied.
func (s *GraphqlAppService) resolveNonAdminRLS(
	ctx context.Context, orgName, projectSlug, modelID string,
) (*modelruntime.RLSContext, context.Context, error) {
	rlsCtx, err := s.buildRLSContext(ctx, orgName, projectSlug, modelID)
	if err != nil {
		return nil, ctx, err
	}
	if rlsCtx.FastFail {
		return nil, ctx, bizerrors.NewError(bizerrors.PermissionDenied, rlsCtx.FastFailReason)
	}
	if rlsCtx.Snapshot != nil {
		ctx = modelruntime.WithRLSSnapshot(ctx, rlsCtx.Snapshot)
	}
	return rlsCtx, ctx, nil
}

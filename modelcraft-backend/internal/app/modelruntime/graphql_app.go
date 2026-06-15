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

	action := mapOperationToAction(cmd.OperationName)

	var rlsCtx *modelruntime.RLSContext
	if ctxutils.GetIsAdminFromContext(ctx) {
		rlsCtx = &modelruntime.RLSContext{
			IsAdmin: true,
			Action:  action,
		}
	} else {
		rlsCtx, err = s.buildRLSContext(ctx, orgName, projectSlug, modelID, action)
		if err != nil {
			return nil, err
		}
		if rlsCtx.FastFail {
			return nil, bizerrors.NewError(bizerrors.PermissionDenied, rlsCtx.FastFailReason)
		}
		if rlsCtx.Snapshot != nil {
			ctx = modelruntime.WithRLSSnapshot(ctx, rlsCtx.Snapshot)
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
	action modelruntime.Action,
) (*modelruntime.RLSContext, error) {
	logger := logfacade.GetLogger(ctx)

	rlsCtx := &modelruntime.RLSContext{
		Action:      action,
		IsAdmin:     ctxutils.GetIsAdminFromContext(ctx),
		UserContext: middleware.GetUserContext(ctx),
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
		ctx, orgName, projectSlug, modelID, rlsCtx.UserContext.Roles, action,
	)
	if err != nil {
		return nil, err
	}
	if rlsCtx.Permissions != nil && rlsCtx.Permissions.IsEmpty() {
		rlsCtx.FastFail = true
		rlsCtx.FastFailReason = "permissions: no permissions granted"
		logger.Debugf(ctx, "RLS fast-fail: model=%s action=%s reason=%s", modelID, action, rlsCtx.FastFailReason)
		return rlsCtx, nil
	}

	snap, err := s.snapshotBuilder.Build(ctx, orgName, projectSlug, modelID, action, rlsCtx.UserContext, rlsCtx.Permissions)
	if err != nil {
		return nil, err
	}
	rlsCtx.Snapshot = snap
	logger.Debugf(ctx, "RLS context built: model=%s action=%s endUser=%s policies=%d using=%v checks=%d",
		modelID, action, endUserID, len(rlsCtx.Permissions.Policies),
		snap.USING != nil, len(snap.CHECKs))

	return rlsCtx, nil
}

// mapOperationToAction converts a GraphQL operation name to the corresponding domain action.
func mapOperationToAction(op string) modelruntime.Action {
	switch {
	case op == "createOne" || op == "createMany":
		return modelruntime.ActionInsert
	case op == "updateOne" || op == "updateMany":
		return modelruntime.ActionUpdate
	case op == "deleteOne" || op == "deleteMany":
		return modelruntime.ActionDelete
	default:
		return modelruntime.ActionSelect // find*, aggregate, count, list*, etc.
	}
}

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
	"modelcraft/pkg/bizerrors"
	"modelcraft/pkg/bizutils"
	"modelcraft/pkg/ctxutils"
	"modelcraft/pkg/logfacade"
	"modelcraft/pkg/requestcontext"

	"github.com/graphql-go/graphql"
	"github.com/graphql-go/graphql/language/ast"
	"github.com/graphql-go/graphql/language/parser"
	"github.com/graphql-go/graphql/language/source"
)

// GraphqlAppService graphql应用服务
type GraphqlAppService struct {
	modelRepo            modelruntime.ModelRepository
	graphqlSchemaManager *modelruntime.GraphqlSchemaManager
	permService          modelruntime.EndUserPermissionService
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

	if err = s.denyManagedModelMutation(ctx, modelLocator, cmd); err != nil {
		return nil, err
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
	clientRepo := dml.NewClientDB(clientSqlDB)

	// 提取 endUserID。
	// Runtime 只认 ctxutils 注入的身份上下文：
	// - end-user（非管理员）: X-User-Type=end_user + X-Is-Admin=false → 走权限检查
	// - end-user（管理员）:   X-User-Type=end_user + X-Is-Admin=true  → 跳过权限检查（与 tenant admin 等同）
	// - tenant admin: 无 end-user 身份（endUserID 为空）
	endUserID := ""
	if ctxutils.IsEndUser(ctx) && !ctxutils.GetIsAdminFromContext(ctx) {
		if uid, err := ctxutils.GetUserIDFromContext(ctx); err == nil {
			endUserID = uid
		}
	}

	// 解析 end-user 权限快照。
	// 仅普通 end-user 请求需要查权限（endUserID != ""），管理员和 tenant admin 跳过，perms 保持 nil。
	var endUserPerms *modelruntime.ResolvedModelPermissions
	if endUserID != "" {
		endUserPerms, err = s.resolveEndUserPerms(ctx, orgName, projectSlug, endUserID, modelLocator)
		if err != nil {
			return nil, err
		}
	}

	endUserAdminID, _ := ctxutils.GetTenantUserIDFromContext(ctx)
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

// resolveEndUserPerms 获取 end-user 权限快照。
// 调用方负责在 endUserID 为空时跳过此方法。
// 需要 model.ID，因此单独加载 model（GetSchema 内部缓存 schema，此处只取 ID）。
func (s *GraphqlAppService) resolveEndUserPerms(
	ctx context.Context,
	orgName, projectSlug, endUserID string,
	modelLocator *modeldesign.ModelLocator,
) (*modelruntime.ResolvedModelPermissions, error) {
	logger := logfacade.GetLogger(ctx)
	model, err := s.modelRepo.GetByName(ctx, modelLocator)
	if err != nil {
		logger.Errorf(ctx, "get model for permission resolve fail: %v", err)
		if shared.IsNotFoundError(err) {
			return nil, bizerrors.NewError(bizerrors.ModelNotFound, modelLocator.GetFullPath())
		}
		return nil, fmt.Errorf("获取模型失败 %s", modelLocator.GetFullPath())
	}
	if model == nil {
		return nil, bizerrors.NewError(bizerrors.ModelNotFound, modelLocator.GetFullPath())
	}
	perms, err := s.permService.Resolve(ctx, orgName, projectSlug, endUserID, model.ID)
	if err != nil {
		logger.Errorf(ctx, "resolve end-user permissions fail: %v", err)
		return nil, fmt.Errorf("解析权限失败")
	}
	return perms, nil
}

func (s *GraphqlAppService) denyManagedModelMutation(
	ctx context.Context,
	modelLocator *modeldesign.ModelLocator,
	cmd ExecuteGraphQLCommand,
) error {
	logger := logfacade.GetLogger(ctx)

	isMutation, err := isMutationOperation(cmd.Query, cmd.OperationName)
	if err != nil {
		return bizerrors.NewError(bizerrors.ParamInvalid, fmt.Sprintf("invalid graphql operation: %v", err))
	}
	if !isMutation {
		return nil
	}

	model, modelErr := s.modelRepo.GetByName(ctx, modelLocator)
	if modelErr != nil {
		logger.Errorf(ctx, "get model fail: %v", modelErr)
		return fmt.Errorf("获取模型失败 %s", modelLocator.GetFullPath())
	}
	if model != nil && model.CreatedVia == modeldesign.ModelCreationSourceImported {
		return bizerrors.NewErrorFromContext(ctx, bizerrors.ManagedModelReadOnly, modelLocator.ModelName)
	}
	return nil
}

func isMutationOperation(query, operationName string) (bool, error) {
	doc, err := parser.Parse(parser.ParseParams{
		Source: source.NewSource(&source.Source{Body: []byte(query), Name: "RuntimeGraphQLRequest"}),
	})
	if err != nil {
		return false, err
	}

	for _, definition := range doc.Definitions {
		opDef, ok := definition.(*ast.OperationDefinition)
		if !ok {
			continue
		}
		if operationName != "" {
			if opDef.Name == nil || opDef.Name.Value != operationName {
				continue
			}
		}
		if opDef.Operation == "mutation" {
			return true, nil
		}
	}
	return false, nil
}

// NewGraphqlAppService 创建graphql应用服务
func NewGraphqlAppService(
	modelRepo modelruntime.ModelRepository,
	lfkRepo modeldesign.LogicalForeignKeyRepository,
	permService modelruntime.EndUserPermissionService,
) *GraphqlAppService {
	schemaManager := modelruntime.NewGraphqlSchemaManager(modelRepo, lfkRepo)
	return &GraphqlAppService{
		modelRepo:            modelRepo,
		graphqlSchemaManager: schemaManager,
		permService:          permService,
	}
}

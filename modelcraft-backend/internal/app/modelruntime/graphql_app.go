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

	// 将请求级状态（clientRepo、dataloader map）注入 context，
	// 所有 resolver 闭包通过 p.Context 读取，与 Schema 类型结构完全解耦。
	// Runtime 只认 ctxutils 注入的身份上下文：
	// - end-user: X-User-Type=end_user + X-User-ID
	// - tenant: 无 end-user 身份（endUserID 为空）
	endUserID := ""
	if ctxutils.IsEndUser(ctx) {
		if uid, err := ctxutils.GetUserIDFromContext(ctx); err == nil {
			endUserID = uid
		}
	}
	reqCtx := modelruntime.WithGraphqlRequestContext(ctx, clientRepo, orgName, projectSlug, endUserID)

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

func (s *GraphqlAppService) denyManagedModelMutation(
	ctx context.Context,
	modelLocator *modeldesign.ModelLocator,
	cmd ExecuteGraphQLCommand,
) error {
	logger := logfacade.GetLogger(ctx)

	isMutation, err := isMutationOperation(cmd.Query, cmd.OperationName)
	if err != nil {
		return bizerrors.Wrapf(err, "invalid graphql operation")
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
) *GraphqlAppService {
	schemaManager := modelruntime.NewGraphqlSchemaManager(modelRepo, lfkRepo)
	return &GraphqlAppService{
		modelRepo:            modelRepo,
		graphqlSchemaManager: schemaManager,
	}
}

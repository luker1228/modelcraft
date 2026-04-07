package modelruntime

import (
	"context"
	"modelcraft/internal/domain/modeldesign"
	"modelcraft/internal/domain/shared"

	"github.com/graphql-go/graphql"
)

// GraphqlSchemaManager GraphQL Schema管理器，用于根据RuntimeModel生成GraphQL Schema
// TODO 这里增加缓存
type GraphqlSchemaManager struct {
	modelRepo ModelRepository
	lfkRepo   modeldesign.LogicalForeignKeyRepository
}

// StoreSchema 存储GraphQL Schema
func (m *GraphqlSchemaManager) StoreSchema(ctx context.Context, modelLocator *modeldesign.ModelLocator,
	gschema *graphql.Schema) {
}

// GetByName 从缓存中获取 GraphQL Schema。
// schema 不存在时返回 shared.ErrRecordNotFound。
func (m *GraphqlSchemaManager) GetByName(ctx context.Context, modelLocator *modeldesign.ModelLocator,
) (*graphql.Schema, error) {
	// TODO: 实现缓存读取逻辑
	return nil, shared.ErrRecordNotFound
}

// NewGraphqlSchemaManager 创建新的GraphQL Schema管理器。
func NewGraphqlSchemaManager(
	modelRepo ModelRepository,
	lfkRepo modeldesign.LogicalForeignKeyRepository,
) *GraphqlSchemaManager {
	return &GraphqlSchemaManager{modelRepo: modelRepo, lfkRepo: lfkRepo}
}

// NewSchemaFrom 根据运行时模型生成GraphQL Schema，包含Query和Mutation操作。
// clientDB 参数已移除：Schema 构建不再依赖请求级 DB 连接，
// 请求级状态通过 graphqlRequestContext 在每次 graphql.Do 时注入。
func (m *GraphqlSchemaManager) NewSchemaFrom(ctx context.Context, model *RuntimeModel,
) (*graphql.Schema, error) {
	resolver := newGraphqlModelResolver(ctx, model, m.modelRepo, m.lfkRepo)
	schema, err := resolver.newGraphqlSchema(ctx)
	if err != nil {
		return nil, err
	}
	return schema, nil
}

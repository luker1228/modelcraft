package modelruntime

import (
	"context"
)

// ClientDatabaseRepository 客户端数据库仓储接口，用于执行GraphQL查询和变更操作。
// 提供CRUD操作和批量操作的支持。
type ClientDatabaseRepository interface {
	// FindUnique 查找唯一的记录。
	FindUnique(ctx context.Context, input *FindUniqueInput) (map[string]any, error)
	// FindFirst 查找第一个匹配的记录。
	FindFirst(ctx context.Context, input *FindFirstInput) (map[string]any, error)
	// FindMany 查找多个匹配的记录。
	FindMany(ctx context.Context, input *FindManyInput) ([]map[string]any, error)
	// ListByCursor executes a keyset cursor pagination query.
	// Returns at most limit+1 rows — the caller checks len(result) > limit to determine hasNextPage.
	ListByCursor(ctx context.Context, input *ListByCursorInput) ([]map[string]any, error)
	// FindManyIn 通过 IN 条件批量查找关联记录，用于解决 N+1 问题。
	// 等价于：SELECT * FROM tableName WHERE referenceKey IN (values...)
	FindManyIn(ctx context.Context, input *FindManyInInput) ([]map[string]any, error)
	// Aggregate 执行聚合查询操作。
	Aggregate(ctx context.Context, input *AggregateInput) (map[string]any, error)
	// Count 执行计数查询操作。
	Count(ctx context.Context, input *CountInput) (map[string]any, error)

	// 变更操作
	// CreateOne 创建单个记录。
	CreateOne(ctx context.Context, input *CreateOneInput) (string, error)
	// UpdateOne 更新单个记录。
	UpdateOne(ctx context.Context, input *UpdateOneInput) (map[string]any, error)
	// DeleteOne 删除单个记录。
	DeleteOne(ctx context.Context, input *DeleteOneInput) (map[string]any, error)
	// CreateMany 批量创建记录。
	CreateMany(ctx context.Context, input *CreateManyInput) (interface{}, error)
	// UpdateMany 批量更新记录。
	UpdateMany(ctx context.Context, input *UpdateManyInput) (interface{}, error)
	// DeleteMany 批量删除记录。
	DeleteMany(ctx context.Context, input *DeleteManyInput) (interface{}, error)
}

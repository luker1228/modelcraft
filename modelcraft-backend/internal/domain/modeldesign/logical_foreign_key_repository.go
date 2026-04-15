package modeldesign

import "context"

// LogicalForeignKeyRepository 逻辑外键仓储接口
type LogicalForeignKeyRepository interface {
	// Save 保存单条逻辑外键记录（用于创建 normal/reverse 两条记录）
	Save(ctx context.Context, lf *LogicalForeignKey) error

	// GetByID 根据 ID 查找单条逻辑外键记录（按主键查询，无需 orgName）
	GetByID(ctx context.Context, id string) (*LogicalForeignKey, error)

	// DeleteByPairID 按 pair_id 删除整个 FK 对（同时删除 normal 和 reverse 两条记录）
	DeleteByPairID(ctx context.Context, orgName, pairID string) error

	// FindByModel 查找指定 org 下模型的所有逻辑外键（WHERE org_name = ? AND model_id = ?）
	FindByModel(ctx context.Context, orgName, modelID string) ([]*LogicalForeignKey, error)

	// FindByPairID 根据 pair_id 查找 FK 对（返回 normal 和 reverse 两条记录）
	FindByPairID(ctx context.Context, orgName, pairID string) ([]*LogicalForeignKey, error)

	// FindByBelongsToField 查找指定 org 下字段作为 belongs_to_fk_id 引用的逻辑外键
	FindByBelongsToField(ctx context.Context, orgName, lfID string) ([]*LogicalForeignKey, error)

	// FindByRelateField 查找指定 org 下字段作为 relate_fk_id 引用的逻辑外键
	FindByRelateField(ctx context.Context, orgName, lfID string) ([]*LogicalForeignKey, error)

	// BindBelongsToFields 将源模型字段批量绑定到 normal 方向 FK 行（写 field_definitions.belongs_to_fk_id）
	BindBelongsToFields(ctx context.Context, orgName, modelID, lfID string, fieldNames []string) error
}

package modeldesign

import (
	"context"
)

// DeployRepo 模型部署仓储接口
type DeployRepo interface {
	// DeployModelToCreate 部署创建模型
	DeployModelToCreate(ctx context.Context, dataModel *DataModel) error
	// DeployModelToDrop 部署删除模型
	DeployModelToDrop(ctx context.Context, dataModel *DataModel) error
	// DeployModelToAddFields 部署添加字段
	DeployModelToAddFields(ctx context.Context, dataModel *DataModel, addFields []*FieldDefinition) error
	// DeployModelToRemoveFields 部署删除字段
	DeployModelToRemoveFields(ctx context.Context, dataModel *DataModel, fieldKeys []string) error
	// CheckTableExists checks whether the underlying database table for the given model already exists.
	// Returns (true, nil) if the table exists, (false, nil) if it does not, or (false, err) on failure.
	CheckTableExists(ctx context.Context, dataModel *DataModel) (bool, error)
}

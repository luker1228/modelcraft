package modelruntime

import (
	"context"
	"modelcraft/internal/domain/modeldesign"
)

// ModelRepository 模型仓储接口，用于访问和管理RuntimeModel。
// 实现了模型的查询和存储操作。
type ModelRepository interface {
	// GetByID 根据ID获取运行时模型。
	GetByID(ctx context.Context, id string) (*RuntimeModel, error)
	// GetByName 根据模型定位器获取运行时模型
	GetByName(ctx context.Context, modelLocator *modeldesign.ModelLocator) (*RuntimeModel, error)
}

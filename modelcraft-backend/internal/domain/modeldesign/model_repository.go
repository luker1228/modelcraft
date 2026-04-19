package modeldesign

import (
	"context"
)

// ModelQueryOptions 模型查询选项
type ModelQueryOptions struct {
	GetFields bool
}

// NewModelQueryOptions 创建新的模型查询选项
func NewModelQueryOptions() *ModelQueryOptions {
	return new(ModelQueryOptions)
}

// WithFields 设置查询时包含字段信息
func (m *ModelQueryOptions) WithFields() *ModelQueryOptions {
	m.GetFields = true
	return m
}

// ApplyOptions 应用选项到配置
func ApplyOptions(opts []*ModelQueryOptions) ModelQueryOptions {
	option := ModelQueryOptions{}
	for _, opt := range opts {
		if opt.GetFields {
			option.GetFields = true
		}
	}
	return option
}

// DeleteFieldRequest 删除字段请求
type DeleteFieldRequest struct {
	ModelId string
	Name    []string
}

// UpdateFieldsStatusRequest 更新字段状态请求
type UpdateFieldsStatusRequest struct {
	ModelId string
	Name    []string
	Status  StatusType
}

// ModelRepository 模型仓储接口
type ModelRepository interface {
	Save(ctx context.Context, orgName string, model *DataModel) error
	Update(ctx context.Context, model *DataModel) error
	UpdateWithVersion(ctx context.Context, model *DataModel, originalVersion int64) (int64, error)
	Delete(ctx context.Context, id string) error
	GetByID(ctx context.Context, id string, opts ...*ModelQueryOptions) (*DataModel, error)
	GetByName(
		ctx context.Context,
		orgName, databaseName, name, projectId string,
		opts ...*ModelQueryOptions,
	) (*DataModel, error)
	FindByDeploymentStatus(ctx context.Context, statuses ...DeploymentStatus) ([]DataModel, error)
	Query(ctx context.Context, queryObj ModelQuery) ([]DataModel, int, error)
	ListDatabaseCatalog(
		ctx context.Context,
		orgName, projectSlug, search string,
		page, pageSize int,
	) ([]string, int, error)

	AddFields(ctx context.Context, orgName string, field []*FieldDefinition) error
	AddRelationField(ctx context.Context, orgName string, field *FieldDefinition) error
	GetFieldByModelID(ctx context.Context, modelID, name string) (*FieldDefinition, error)
	GetFieldsByModelID(ctx context.Context, modelID string) ([]*FieldDefinition, error)
	// GetTailFieldDisplayOrder returns the largest display_order value among fields in the model,
	// or an empty string if no fields exist.
	GetTailFieldDisplayOrder(ctx context.Context, modelID string) (string, error)
	UpdateField(ctx context.Context, field *FieldDefinition) error
	BulkUpdateFields(ctx context.Context, field []*FieldDefinition) error
	UpdateFieldsStatus(ctx context.Context, requests ...UpdateFieldsStatusRequest) error
	DeleteFields(ctx context.Context, modelID string, names []string) error
	BulkDeleteFields(ctx context.Context, requests ...DeleteFieldRequest) error
}

package modeldesign

import "context"

// FieldEnumAssociationRepository 字段枚举关联仓储接口
type FieldEnumAssociationRepository interface {
	// Create 创建字段枚举关联
	Create(ctx context.Context, association *FieldEnumAssociation) error

	// FindByField 根据模型ID和字段名称查找关联
	FindByField(ctx context.Context, modelID, fieldName string) (*FieldEnumAssociation, error)

	// FindByEnumName 根据枚举名称查找所有关联的字段
	FindByEnumName(ctx context.Context, projectID, enumName string) ([]*FieldEnumAssociation, error)

	// FindByModelID 根据模型ID查找所有字段的枚举关联
	FindByModelID(ctx context.Context, modelID string) ([]*FieldEnumAssociation, error)

	// Delete 删除字段枚举关联
	Delete(ctx context.Context, modelID, fieldName string) error

	// DeleteByModelID 删除模型的所有字段枚举关联
	DeleteByModelID(ctx context.Context, modelID string) error
}

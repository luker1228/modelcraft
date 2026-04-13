package modeldesign

import "context"

// FieldEnumRelationRepository 定义字段枚举关联仓储契约。
type FieldEnumRelationRepository interface {
	Create(ctx context.Context, relation *FieldEnumRelation) error
	FindByID(ctx context.Context, orgName, id string) (*FieldEnumRelation, error)
	FindBySourceField(ctx context.Context, orgName, modelID, sourceFieldName string) (*FieldEnumRelation, error)
	FindByLabelField(ctx context.Context, orgName, modelID, labelFieldName string) (*FieldEnumRelation, error)
	ListByModelID(ctx context.Context, orgName, modelID string) ([]*FieldEnumRelation, error)
	CountBySourceField(ctx context.Context, orgName, modelID, sourceFieldName string) (int64, error)
	CountByLabelField(ctx context.Context, orgName, modelID, labelFieldName string) (int64, error)
	Delete(ctx context.Context, orgName, id string) error
}

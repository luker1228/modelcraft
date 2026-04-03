package modeldesign

import (
	"modelcraft/internal/domain/project"
	"time"

	bizerrors "modelcraft/pkg/bizerrors"
)

// FieldEnumAssociation 字段枚举关联实体
// 管理模型字段与枚举定义之间的多对一关系
type FieldEnumAssociation struct {
	ModelID              string    `json:"modelId"`   // 模型ID
	FieldName            string    `json:"fieldName"` // 字段名称
	project.ProjectScope           // 嵌入项目作用域，包含 OrgName 和 ProjectSlug
	EnumName             string    `json:"enumName"`     // 枚举名称
	DatabaseName         string    `json:"databaseName"` // 数据库名称
	CreatedAt            time.Time `json:"createdAt"`
	UpdatedAt            time.Time `json:"updatedAt"`
}

// Validate 验证关联记录
func (fea *FieldEnumAssociation) Validate() error {
	if fea.ModelID == "" {
		return bizerrors.New("model_id cannot be empty")
	}
	if fea.FieldName == "" {
		return bizerrors.New("field_name cannot be empty")
	}
	if err := fea.ProjectScope.Validate(); err != nil {
		return err
	}
	if fea.EnumName == "" {
		return bizerrors.New("enum_name cannot be empty")
	}
	if fea.DatabaseName == "" {
		return bizerrors.New("database_name cannot be empty")
	}
	return nil
}

// NewFieldEnumAssociation 创建新的字段枚举关联
func NewFieldEnumAssociation(
	modelID, fieldName, orgName, projectSlug, enumName, databaseName string,
) (*FieldEnumAssociation, error) {
	now := time.Now()
	assoc := &FieldEnumAssociation{
		ModelID:   modelID,
		FieldName: fieldName,
		ProjectScope: project.ProjectScope{
			OrgName:     orgName,
			ProjectSlug: projectSlug,
		},
		EnumName:     enumName,
		DatabaseName: databaseName,
		CreatedAt:    now,
		UpdatedAt:    now,
	}

	if err := assoc.Validate(); err != nil {
		return nil, err
	}

	return assoc, nil
}

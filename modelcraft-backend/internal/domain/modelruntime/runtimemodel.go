package modelruntime

import (
	"modelcraft/internal/domain/modeldesign"
	"modelcraft/pkg/bizerrors"

	"github.com/graphql-go/graphql"
)

// RuntimeModel 运行时模型定义，包含模型的基本信息和字段定义。
// 这是GraphQL Schema生成的核心数据结构。
type RuntimeModel struct {
	OrgName      string                   `json:"orgName"`
	ProjectSlug  string                   `json:"projectSlug"`
	Name         string                   `json:"name"`
	Title        string                   `json:"title"`
	Description  string                   `json:"description"`
	DatabaseName string                   `json:"databaseName"`
	DisplayField *string                  `json:"displayField"` // 用于 _displayName 解析的字段名
	Fields       map[string]*RuntimeField `json:"fields"`
}

func (m *RuntimeModel) getUniqueField() []*RuntimeField {
	uniqueFields := make([]*RuntimeField, 0, 3)
	for _, field := range m.Fields {
		if field.IsUnique {
			uniqueFields = append(uniqueFields, field)
		}
	}

	return uniqueFields
}

// RuntimeField 运行时字段定义，别名为FieldDefinition
type RuntimeField = modeldesign.FieldDefinition

func getGraphqlTypeBy(formatType modeldesign.FormatType) (scalar *graphql.Scalar, err error) {
	switch formatType {
	case modeldesign.FormatString:
		scalar = graphql.String
	case modeldesign.FormatUUID:
		scalar = graphql.ID
	case modeldesign.FormatNumber, modeldesign.FormatDecimal:
		scalar = graphql.Float
	case modeldesign.FormatInteger:
		scalar = graphql.Int
	case modeldesign.FormatBoolean:
		scalar = graphql.Boolean
	case modeldesign.FormatDateTime:
		scalar = graphql.DateTime
	case modeldesign.FormatDate:
		scalar = GraphQLDate
	case modeldesign.FormatTime:
		scalar = GraphQLTime
	case modeldesign.FormatEnum:
		scalar = graphql.String
	default:
		return nil, bizerrors.Errorf("unknown fmtType: %s", formatType)
	}
	return
}

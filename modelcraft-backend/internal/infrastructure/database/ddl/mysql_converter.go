package ddl

import (
	"context"
	"fmt"
	"modelcraft/pkg/bizerrors"
	"modelcraft/pkg/bizutils"
	"modelcraft/pkg/ddlfactory"
	"modelcraft/pkg/logfacade"

	entity "modelcraft/internal/domain/modeldesign"
)

// DDLConverter Schema转换器，用于生成DDL语句
type DDLConverter struct {
	ctx          context.Context
	mysqlBuilder *ddlfactory.MySQLDDLBuilder
}

// NewDdlConverter 创建Schema转换器
func newDdlConverter(ctx context.Context) *DDLConverter {
	return &DDLConverter{
		ctx:          ctx,
		mysqlBuilder: ddlfactory.NewMySQLDDLBuilder(ctx),
	}
}

// GenerateCreateTableDDL 生成创建表的DDL
func (s *DDLConverter) GenerateCreateTableDDL(model *entity.DataModel) (string, error) {
	// 遍历所有字段生成列定义
	fieldEntitys := make([]*ddlfactory.FieldEntity, 0, len(model.Fields))
	for _, field := range model.Fields {
		fieldEntity, err := s.generateColumnDefinition(field)
		if err != nil {
			return "", fmt.Errorf("生成字段%s的列定义失败: %w", field.Name, err)
		}
		fieldEntitys = append(fieldEntitys, fieldEntity)
	}
	logfacade.GetLogger(s.ctx).Infof(context.Background(),
		"fieldEntitys: %s", bizutils.MarshalToStringIgnoreErr(fieldEntitys))

	ddl, err := s.mysqlBuilder.BuildCreateTable(&ddlfactory.TableEntity{
		Name:    model.ModelName,
		Fields:  fieldEntitys,
		Engine:  "InnoDB",
		Charset: "utf8mb4",
		Comment: model.Title,
	})
	if err != nil {
		return "", err
	}

	return ddl, nil
}

// generateColumnDefinition 生成单个列的定义
func (s *DDLConverter) generateColumnDefinition(fieldDef *entity.FieldDefinition) (*ddlfactory.FieldEntity, error) {
	if fieldDef.IsRelationField() {
		return nil, fmt.Errorf("关系字段不支持生成列定义")
	}

	// 使用TypeMapper获取MySQL类型
	typeMapper := entity.NewMySQLTypeMapper()
	mysqlType, err := typeMapper.MapToMySQL(fieldDef)
	if err != nil {
		return nil, bizerrors.NewError(bizerrors.SystemError,
			fmt.Sprintf("映射字段类型失败: %v", err))
	}

	mysqlField := ddlfactory.FieldEntity{
		Name: fieldDef.Name,
	}

	// 解析MySQL类型字符串，设置Type和Length
	if err := s.parseAndSetMySQLType(&mysqlField, mysqlType); err != nil {
		return nil, err
	}

	mysqlField.Nullable = !fieldDef.NonNull // NonNull为true表示不可为空，所以Nullable应该为false
	// 设置约束
	if fieldDef.IsPrimary {
		mysqlField.Primary = true
		mysqlField.Nullable = false
	}
	if fieldDef.Description != "" {
		mysqlField.Comment = fieldDef.Description
	}

	return &mysqlField, nil
}

// parseAndSetMySQLType 解析MySQL类型字符串并设置到FieldEntity
func (s *DDLConverter) parseAndSetMySQLType(field *ddlfactory.FieldEntity, mysqlType string) error {
	// 处理带长度的类型，如 VARCHAR(100), DECIMAL(10,2)
	switch {
	case len(mysqlType) >= 7 && mysqlType[:7] == "VARCHAR":
		field.Type = ddlfactory.VARCHAR
		// 从字符串中提取长度
		if len(mysqlType) > 7 {
			// 解析 VARCHAR(100)
			var length int
			_, _ = fmt.Sscanf(mysqlType, "VARCHAR(%d)", &length)
			field.Length = &length
		}
	case len(mysqlType) >= 4 && mysqlType[:4] == "CHAR":
		field.Type = ddlfactory.CHAR
		if len(mysqlType) > 4 {
			var length int
			_, _ = fmt.Sscanf(mysqlType, "CHAR(%d)", &length)
			field.Length = &length
		}
	case mysqlType == "TEXT":
		field.Type = ddlfactory.TEXT
	case mysqlType == "MEDIUMTEXT":
		// MEDIUMTEXT不在ddlfactory中，使用TEXT
		field.Type = ddlfactory.TEXT
	case mysqlType == "LONGTEXT":
		// LONGTEXT不在ddlfactory中，使用TEXT
		field.Type = ddlfactory.TEXT
	case mysqlType == "INT":
		field.Type = ddlfactory.INT
	case mysqlType == "DOUBLE":
		field.Type = ddlfactory.DOUBLE
	case len(mysqlType) > 7 && mysqlType[:7] == "DECIMAL":
		field.Type = ddlfactory.DECIMAL
		var precision, scale int
		_, _ = fmt.Sscanf(mysqlType, "DECIMAL(%d,%d)", &precision, &scale)
		field.Precision = &precision
		field.Scale = &scale
	case mysqlType == "TINYINT(1)" || mysqlType == "BOOL":
		field.Type = ddlfactory.BOOL
	case mysqlType == "DATE":
		field.Type = ddlfactory.DATE
	case mysqlType == "DATETIME":
		field.Type = ddlfactory.DATETIME
	case mysqlType == "TIME":
		// TIME不在ddlfactory中，使用VARCHAR(8)存储 HH:MM:SS
		field.Type = ddlfactory.VARCHAR
		length := 8
		field.Length = &length
	case mysqlType == "JSON":
		field.Type = ddlfactory.JSON
	default:
		return fmt.Errorf("不支持的MySQL类型: %s", mysqlType)
	}

	return nil
}

// GenerateDropTableDDL 生成删除表的DDL语句
func (s *DDLConverter) GenerateDropTableDDL(model *entity.DataModel) (string, error) {
	table, err := s.mysqlBuilder.BuildDropTable(model.ModelName)
	if err != nil {
		return "", err
	}
	return table, nil
}

// GenerateAddColumns 生成添加列的DDL语句（批量添加，生成一条SQL）
func (s *DDLConverter) GenerateAddColumns(model *entity.DataModel, fields []*entity.FieldDefinition) (string, error) {
	if len(fields) == 0 {
		return "", fmt.Errorf("字段列表不能为空")
	}

	// 转换所有字段定义
	fieldEntities := make([]*ddlfactory.FieldEntity, 0, len(fields))
	for _, fieldDef := range fields {
		field, err := s.generateColumnDefinition(fieldDef)
		if err != nil {
			return "", fmt.Errorf("生成字段%s的列定义失败: %w", fieldDef.Name, err)
		}
		fieldEntities = append(fieldEntities, field)
	}

	// 使用批量添加方法生成一条SQL语句
	ddl, err := s.mysqlBuilder.BuildAddColumns(model.ModelName, fieldEntities)
	if err != nil {
		return "", fmt.Errorf("生成批量添加列DDL失败: %w", err)
	}

	// 返回包含单条SQL的切片
	return ddl, nil
}

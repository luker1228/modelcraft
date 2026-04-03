package ddlfactory

import (
	"context"
	"fmt"
	"modelcraft/pkg/logfacade"
	"strings"
)

// MySQLDDLBuilder MySQL DDL构建器
type MySQLDDLBuilder struct {
	ctx context.Context
}

// NewMySQLDDLBuilder 创建MySQL DDL构建器
func NewMySQLDDLBuilder(ctx context.Context) *MySQLDDLBuilder {
	return &MySQLDDLBuilder{
		ctx: ctx,
	}
}

// BuildDropTable 构建删除表的DDL语句
func (b *MySQLDDLBuilder) BuildDropTable(tableName string) (string, error) {
	if tableName == "" {
		return "", fmt.Errorf("表名不能为空")
	}

	// 构建DROP TABLE语句
	sql := fmt.Sprintf("DROP TABLE IF EXISTS `%s`", tableName)
	return sql, nil
}

// BuildAddColumn 构建添加列的DDL语句（使用在线DDL，不锁表）
func (b *MySQLDDLBuilder) BuildAddColumn(tableName string, field *FieldEntity) (string, error) {
	if tableName == "" {
		return "", fmt.Errorf("表名不能为空")
	}

	// 构建字段定义
	fieldDef, err := b.buildFieldDefinition(field)
	if err != nil {
		return "", fmt.Errorf("构建字段 %s 定义失败: %w", field.Name, err)
	}

	// 构建ALTER TABLE ADD COLUMN语句，使用在线DDL
	// ALGORITHM=INPLACE: 尽可能使用原地算法，避免重建整个表
	// LOCK=NONE: 不锁表，允许并发读写
	sql := fmt.Sprintf("ALTER TABLE `%s` ADD COLUMN %s, ALGORITHM=INPLACE, LOCK=NONE", tableName, fieldDef)
	return sql, nil
}

// BuildAddColumns 构建批量添加列的DDL语句（使用在线DDL，不锁表）
// 将多个字段的添加操作合并为一条SQL语句，提高执行效率
func (b *MySQLDDLBuilder) BuildAddColumns(tableName string, fields []*FieldEntity) (string, error) {
	if tableName == "" {
		return "", fmt.Errorf("表名不能为空")
	}

	if len(fields) == 0 {
		return "", fmt.Errorf("字段列表不能为空")
	}

	// 构建所有字段定义
	fieldDefs := make([]string, 0, len(fields))
	for _, field := range fields {
		fieldDef, err := b.buildFieldDefinition(field)
		if err != nil {
			return "", fmt.Errorf("构建字段 %s 定义失败: %w", field.Name, err)
		}
		fieldDefs = append(fieldDefs, fmt.Sprintf("ADD COLUMN %s", fieldDef))
	}

	// 构建ALTER TABLE语句，使用在线DDL
	// ALGORITHM=INPLACE: 尽可能使用原地算法，避免重建整个表
	// LOCK=NONE: 不锁表，允许并发读写
	sql := fmt.Sprintf("ALTER TABLE `%s` %s, ALGORITHM=INPLACE, LOCK=NONE",
		tableName,
		strings.Join(fieldDefs, ", "))
	return sql, nil
}

// BuildCreateTable 构建创建表的DDL语句
func (b *MySQLDDLBuilder) BuildCreateTable(table *TableEntity) (string, error) {
	if table.Name == "" {
		return "", fmt.Errorf("表名不能为空")
	}

	if len(table.Fields) == 0 {
		return "", fmt.Errorf("字段列表不能为空")
	}

	logger := logfacade.GetLogger(b.ctx)
	var sql strings.Builder

	// CREATE TABLE 语句开始
	sql.WriteString(fmt.Sprintf("CREATE TABLE IF NOT EXISTS `%s` (\n", table.Name))

	// 构建字段定义
	fieldDefs := make([]string, 0, len(table.Fields))
	var primaryKeys []string

	for _, field := range table.Fields {
		fieldDef, err := b.buildFieldDefinition(field)
		if err != nil {
			logger.Errorf(b.ctx, "field=%#v", field)
			return "", fmt.Errorf("构建字段 %v 定义失败: %w", field.Name, err)
		}
		fieldDefs = append(fieldDefs, "  "+fieldDef)

		// 收集主键字段
		if field.Primary {
			primaryKeys = append(primaryKeys, field.Name)
		}
	}

	// 添加主键约束
	if len(primaryKeys) > 0 {
		pkDef := fmt.Sprintf("  PRIMARY KEY (`%s`)", strings.Join(primaryKeys, "`, `"))
		fieldDefs = append(fieldDefs, pkDef)
	}

	// 连接所有字段定义
	sql.WriteString(strings.Join(fieldDefs, ",\n"))
	sql.WriteString("\n)")

	// 添加表选项
	b.buildTableOptions(&sql, table)

	return sql.String(), nil
}

// buildFieldDefinition 构建字段定义
func (b *MySQLDDLBuilder) buildFieldDefinition(field *FieldEntity) (string, error) {
	var def strings.Builder

	// 字段名
	def.WriteString(fmt.Sprintf("`%s` ", field.Name))

	// 数据类型
	typeStr, err := b.buildDataType(field)
	if err != nil {
		return "", err
	}
	def.WriteString(typeStr)

	// NULL/NOT NULL
	if field.Nullable {
		def.WriteString(" NULL")
	} else {
		def.WriteString(" NOT NULL")
	}

	// 默认值
	if field.DefaultValue != nil {
		def.WriteString(fmt.Sprintf(" DEFAULT %s", b.formatDefaultValue(*field.DefaultValue, field.Type)))
	}

	// 自增
	if field.AutoIncrement {
		def.WriteString(" AUTO_INCREMENT")
	}

	// 唯一约束
	if field.Unique && !field.Primary {
		def.WriteString(" UNIQUE")
	}

	// 注释
	if field.Comment != "" {
		def.WriteString(fmt.Sprintf(" COMMENT '%s'", b.escapeString(field.Comment)))
	}

	return def.String(), nil
}

// buildDataType 构建数据类型
func (b *MySQLDDLBuilder) buildDataType(field *FieldEntity) (string, error) {
	switch field.Type {
	case TINYINT, SMALLINT, INT, BIGINT, BOOL:
		return string(field.Type), nil

	case FLOAT, DOUBLE:
		if field.Precision != nil && field.Scale != nil {
			return fmt.Sprintf("%s(%d,%d)", field.Type, *field.Precision, *field.Scale), nil
		}
		return string(field.Type), nil

	case DECIMAL:
		if field.Precision != nil {
			if field.Scale != nil {
				return fmt.Sprintf("%s(%d,%d)", field.Type, *field.Precision, *field.Scale), nil
			}
			return fmt.Sprintf("%s(%d)", field.Type, *field.Precision), nil
		} else {
			return "", fmt.Errorf("DECIMAL类型必须指定精度")
		}

	case CHAR, VARCHAR:
		if field.Length == nil {
			return "", fmt.Errorf("%s类型必须指定长度", field.Type)
		}
		return fmt.Sprintf("%s(%d)", field.Type, *field.Length), nil

	case TEXT, DATE, DATETIME, TIMESTAMP, JSON:
		return string(field.Type), nil

	default:
		return "", fmt.Errorf("不支持的数据类型: %s", field.Type)
	}
}

// formatDefaultValue 格式化默认值
func (b *MySQLDDLBuilder) formatDefaultValue(value string, dataType MySQLDataType) string {
	// 特殊函数值不需要引号
	specialValues := []string{
		"CURRENT_TIMESTAMP", "NOW()", "NULL",
	}

	for _, special := range specialValues {
		if strings.ToUpper(value) == special {
			return value
		}
	}

	// 数值类型不需要引号
	switch dataType {
	case TINYINT, SMALLINT, INT, BIGINT, DECIMAL, FLOAT, DOUBLE:
		return value
	default:
		// 字符串类型需要引号
		return fmt.Sprintf("'%s'", b.escapeString(value))
	}
}

// buildTableOptions 构建表选项
func (b *MySQLDDLBuilder) buildTableOptions(sql *strings.Builder, table *TableEntity) {
	options := make([]string, 0, 3)

	// 存储引擎
	engine := table.Engine
	if engine == "" {
		engine = "InnoDB" // 默认引擎
	}
	options = append(options, fmt.Sprintf("ENGINE=%s", engine))

	// 字符集
	charset := table.Charset
	if charset == "" {
		charset = "utf8mb4" // 默认字符集
	}
	options = append(options, fmt.Sprintf("DEFAULT CHARSET=%s", charset))

	// 表注释
	if table.Comment != "" {
		options = append(options, fmt.Sprintf("COMMENT='%s'", b.escapeString(table.Comment)))
	}

	if len(options) > 0 {
		sql.WriteString(" ")
		sql.WriteString(strings.Join(options, " "))
	}
}

// IntPtr 返回给定整数值的指针
// 用于创建可选的整数字段
func IntPtr(i int) *int {
	return &i
}

// escapeString 转义字符串中的特殊字符
func (b *MySQLDDLBuilder) escapeString(s string) string {
	// 简单的转义实现，实际项目中可能需要更完善的转义
	s = strings.ReplaceAll(s, "'", "''")
	s = strings.ReplaceAll(s, "\\", "\\\\")
	return s
}

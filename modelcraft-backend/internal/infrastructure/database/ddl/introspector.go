package ddl

import (
	"context"
	"database/sql"
	"fmt"
	"strings"

	"github.com/pkg/errors"
)

// SchemaIntrospector Schema内省器接口
type SchemaIntrospector interface {
	IntrospectTable(ctx context.Context, db *sql.DB, tableName string) (*TableDefinition, error)
}

// MySQLIntrospector MySQL Schema内省器
type MySQLIntrospector struct{}

// NewSchemaIntrospector 创建Schema内省器实例
func NewSchemaIntrospector() SchemaIntrospector {
	return &MySQLIntrospector{}
}

// columnInfo 从INFORMATION_SCHEMA查询的列信息
type columnInfo struct {
	ColumnName             string
	DataType               string
	CharacterMaximumLength sql.NullInt64
	NumericPrecision       sql.NullInt64
	NumericScale           sql.NullInt64
	IsNullable             string
	ColumnDefault          sql.NullString
	Extra                  string
	ColumnComment          string
}

// IntrospectTable 从数据库内省表结构
func (i *MySQLIntrospector) IntrospectTable(
	ctx context.Context,
	db *sql.DB,
	tableName string,
) (*TableDefinition, error) {
	// 获取当前数据库名
	var dbName string
	if err := db.QueryRowContext(ctx, "SELECT DATABASE()").Scan(&dbName); err != nil {
		return nil, errors.Wrap(err, "failed to get current database name")
	}

	if dbName == "" {
		return nil, fmt.Errorf("no database selected")
	}

	// 1. 查询列信息
	columns, err := i.queryColumns(ctx, db, dbName, tableName)
	if err != nil {
		return nil, err
	}

	if len(columns) == 0 {
		return nil, fmt.Errorf("table %s not found or has no columns", tableName)
	}

	// 2. 查询主键信息
	primaryKeys, err := i.queryPrimaryKeys(ctx, db, dbName, tableName)
	if err != nil {
		return nil, err
	}

	// 3. 查询表注释
	tableComment, err := i.queryTableComment(ctx, db, dbName, tableName)
	if err != nil {
		return nil, err
	}

	// 构建TableDefinition
	tableDef := &TableDefinition{
		TableName:   tableName,
		Columns:     make([]ColumnDefinition, 0, len(columns)),
		PrimaryKeys: primaryKeys,
		Indexes:     []IndexDefinition{}, // 暂时不提取索引信息
		Comment:     tableComment,
	}

	// 转换列信息
	for _, col := range columns {
		colDef := i.convertColumnInfo(col)
		tableDef.Columns = append(tableDef.Columns, colDef)
	}

	return tableDef, nil
}

// queryColumns 查询列信息
func (i *MySQLIntrospector) queryColumns(
	ctx context.Context,
	db *sql.DB,
	dbName, tableName string,
) ([]columnInfo, error) {
	query := `
		SELECT
			COLUMN_NAME,
			DATA_TYPE,
			CHARACTER_MAXIMUM_LENGTH,
			NUMERIC_PRECISION,
			NUMERIC_SCALE,
			IS_NULLABLE,
			COLUMN_DEFAULT,
			EXTRA,
			COLUMN_COMMENT
		FROM INFORMATION_SCHEMA.COLUMNS
		WHERE TABLE_SCHEMA = ? AND TABLE_NAME = ?
		ORDER BY ORDINAL_POSITION
	`

	rows, err := db.QueryContext(ctx, query, dbName, tableName)
	if err != nil {
		return nil, errors.Wrap(err, "failed to query column information")
	}
	defer rows.Close()

	var columns []columnInfo
	for rows.Next() {
		var col columnInfo
		if err := rows.Scan(
			&col.ColumnName,
			&col.DataType,
			&col.CharacterMaximumLength,
			&col.NumericPrecision,
			&col.NumericScale,
			&col.IsNullable,
			&col.ColumnDefault,
			&col.Extra,
			&col.ColumnComment,
		); err != nil {
			return nil, errors.Wrap(err, "failed to scan column information")
		}
		columns = append(columns, col)
	}

	if err := rows.Err(); err != nil {
		return nil, errors.Wrap(err, "failed to iterate column information")
	}

	return columns, nil
}

// queryPrimaryKeys 查询主键信息
func (i *MySQLIntrospector) queryPrimaryKeys(
	ctx context.Context,
	db *sql.DB,
	dbName, tableName string,
) ([]string, error) {
	query := `
		SELECT COLUMN_NAME
		FROM INFORMATION_SCHEMA.KEY_COLUMN_USAGE
		WHERE TABLE_SCHEMA = ?
		  AND TABLE_NAME = ?
		  AND CONSTRAINT_NAME = 'PRIMARY'
		ORDER BY ORDINAL_POSITION
	`

	rows, err := db.QueryContext(ctx, query, dbName, tableName)
	if err != nil {
		return nil, errors.Wrap(err, "failed to query primary key information")
	}
	defer rows.Close()

	var primaryKeys []string
	for rows.Next() {
		var pk string
		if err := rows.Scan(&pk); err != nil {
			return nil, errors.Wrap(err, "failed to scan primary key")
		}
		primaryKeys = append(primaryKeys, pk)
	}

	if err := rows.Err(); err != nil {
		return nil, errors.Wrap(err, "failed to query primary key information")
	}

	return primaryKeys, nil
}

// queryTableComment 查询表注释
func (i *MySQLIntrospector) queryTableComment(
	ctx context.Context,
	db *sql.DB,
	dbName, tableName string,
) (string, error) {
	query := `
		SELECT TABLE_COMMENT
		FROM INFORMATION_SCHEMA.TABLES
		WHERE TABLE_SCHEMA = ? AND TABLE_NAME = ?
	`

	var tableComment string
	if err := db.QueryRowContext(ctx, query, dbName, tableName).Scan(&tableComment); err != nil {
		return "", errors.Wrap(err, "failed to query table comment")
	}

	return tableComment, nil
}

// convertColumnInfo 转换列信息到ColumnDefinition
func (i *MySQLIntrospector) convertColumnInfo(col columnInfo) ColumnDefinition {
	colDef := ColumnDefinition{
		Name:          col.ColumnName,
		DataType:      strings.ToUpper(col.DataType),
		Nullable:      col.IsNullable == "YES",
		AutoIncrement: strings.Contains(strings.ToUpper(col.Extra), "AUTO_INCREMENT"),
		Comment:       col.ColumnComment,
	}

	// 设置长度
	if col.CharacterMaximumLength.Valid {
		colDef.Length = col.CharacterMaximumLength.Int64
	}

	// 设置精度和小数位
	if col.NumericPrecision.Valid {
		colDef.Precision = int(col.NumericPrecision.Int64)
	}
	if col.NumericScale.Valid {
		colDef.Scale = int(col.NumericScale.Int64)
	}

	// 设置默认值
	if col.ColumnDefault.Valid {
		defaultValue := col.ColumnDefault.String
		colDef.DefaultValue = &defaultValue
	}

	return colDef
}

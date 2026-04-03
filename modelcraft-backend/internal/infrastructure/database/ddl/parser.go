package ddl

import (
	"fmt"
	"strings"

	"github.com/pingcap/errors"
	"github.com/pingcap/tidb/pkg/parser"
	"github.com/pingcap/tidb/pkg/parser/ast"
	_ "github.com/pingcap/tidb/pkg/parser/test_driver"
)

// TableDefinition 表定义结构
type TableDefinition struct {
	TableName   string
	Columns     []ColumnDefinition
	PrimaryKeys []string
	Indexes     []IndexDefinition
	Comment     string
}

// ColumnDefinition 列定义结构
type ColumnDefinition struct {
	Name          string
	DataType      string
	Length        int64
	Precision     int
	Scale         int
	Nullable      bool
	DefaultValue  *string
	AutoIncrement bool
	Comment       string
}

// IndexDefinition 索引定义结构
type IndexDefinition struct {
	Name    string
	Columns []string
	Unique  bool
}

// DDLParser DDL解析器接口
type DDLParser interface {
	ParseCreateTable(ddl string) (*TableDefinition, error)
}

// TiDBDDLParser 基于TiDB的DDL解析器
type TiDBDDLParser struct {
	parser *parser.Parser
}

// NewDDLParser 创建DDL解析器实例
func NewDDLParser() DDLParser {
	return &TiDBDDLParser{
		parser: parser.New(),
	}
}

// ParseCreateTable 解析CREATE TABLE语句
func (p *TiDBDDLParser) ParseCreateTable(ddl string) (*TableDefinition, error) {
	stmts, _, err := p.parser.Parse(ddl, "", "")
	if err != nil {
		return nil, errors.Wrap(err, "failed to parse DDL statement")
	}

	if len(stmts) == 0 {
		return nil, fmt.Errorf("no statement found in DDL")
	}

	createTableStmt, ok := stmts[0].(*ast.CreateTableStmt)
	if !ok {
		return nil, fmt.Errorf("not a CREATE TABLE statement")
	}

	tableDef := &TableDefinition{
		TableName:   createTableStmt.Table.Name.String(),
		Columns:     []ColumnDefinition{},
		PrimaryKeys: []string{},
		Indexes:     []IndexDefinition{},
	}

	// 提取表注释
	for _, option := range createTableStmt.Options {
		if option.Tp == ast.TableOptionComment {
			tableDef.Comment = option.StrValue
		}
	}

	// 提取列定义
	for _, col := range createTableStmt.Cols {
		colDef := p.extractColumnDefinition(col)
		tableDef.Columns = append(tableDef.Columns, colDef)
	}

	// 提取约束（主键、索引等）
	for _, constraint := range createTableStmt.Constraints {
		switch constraint.Tp {
		case ast.ConstraintPrimaryKey:
			// 主键
			for _, key := range constraint.Keys {
				tableDef.PrimaryKeys = append(tableDef.PrimaryKeys, key.Column.Name.String())
			}
		case ast.ConstraintKey, ast.ConstraintIndex:
			// 普通索引
			indexDef := IndexDefinition{
				Name:    constraint.Name,
				Columns: []string{},
				Unique:  false,
			}
			for _, key := range constraint.Keys {
				indexDef.Columns = append(indexDef.Columns, key.Column.Name.String())
			}
			tableDef.Indexes = append(tableDef.Indexes, indexDef)
		case ast.ConstraintUniq, ast.ConstraintUniqKey, ast.ConstraintUniqIndex:
			// 唯一索引
			indexDef := IndexDefinition{
				Name:    constraint.Name,
				Columns: []string{},
				Unique:  true,
			}
			for _, key := range constraint.Keys {
				indexDef.Columns = append(indexDef.Columns, key.Column.Name.String())
			}
			tableDef.Indexes = append(tableDef.Indexes, indexDef)
		}
	}

	return tableDef, nil
}

// extractColumnDefinition 从AST列定义中提取列信息
func (p *TiDBDDLParser) extractColumnDefinition(col *ast.ColumnDef) ColumnDefinition {
	colDef := ColumnDefinition{
		Name:          col.Name.Name.String(),
		Nullable:      true, // 默认可空
		AutoIncrement: false,
	}

	// 提取数据类型
	if col.Tp != nil {
		colDef.DataType = strings.ToUpper(col.Tp.InfoSchemaStr())

		// 提取长度/精度/小数位
		if col.Tp.GetFlen() > 0 {
			colDef.Length = int64(col.Tp.GetFlen())
		}
		if col.Tp.GetDecimal() >= 0 {
			colDef.Precision = col.Tp.GetFlen()
			colDef.Scale = col.Tp.GetDecimal()
		}
	}

	// 提取列选项（NOT NULL, DEFAULT, AUTO_INCREMENT, COMMENT等）
	for _, option := range col.Options {
		switch option.Tp {
		case ast.ColumnOptionNotNull:
			colDef.Nullable = false
		case ast.ColumnOptionNull:
			colDef.Nullable = true
		case ast.ColumnOptionAutoIncrement:
			colDef.AutoIncrement = true
		case ast.ColumnOptionDefaultValue:
			if option.Expr != nil {
				// 尝试提取默认值
				defaultStr := p.extractDefaultValue(option.Expr)
				colDef.DefaultValue = &defaultStr
			}
		case ast.ColumnOptionComment:
			if option.Expr != nil {
				colDef.Comment = p.extractStringValue(option.Expr)
			}
		case ast.ColumnOptionPrimaryKey:
			// 列级主键标记（会在constraint中重复处理，这里不处理）
		}
	}

	return colDef
}

// extractDefaultValue 提取默认值表达式
func (p *TiDBDDLParser) extractDefaultValue(expr ast.ExprNode) string {
	switch v := expr.(type) {
	case *ast.FuncCallExpr:
		// 函数调用，如 CURRENT_TIMESTAMP
		return strings.ToUpper(v.FnName.String())
	default:
		// 其他类型，返回文本表示
		return expr.Text()
	}
}

// extractStringValue 提取字符串值（用于注释等）
func (p *TiDBDDLParser) extractStringValue(expr ast.ExprNode) string {
	// 直接使用Text()方法获取字符串表示
	return expr.Text()
}

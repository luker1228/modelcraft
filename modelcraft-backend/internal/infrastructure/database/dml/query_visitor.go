package dml

import (
	"modelcraft/internal/domain/query"
	"modelcraft/pkg/bizerrors"
	"reflect"

	"github.com/doug-martin/goqu/v9"
	"github.com/doug-martin/goqu/v9/exp"
)

// QueryVisitor 查询访问者接口 - 访问者模式的核心接口
type QueryVisitor interface {
	VisitField(node *FieldNode) (goqu.Expression, error)
	VisitLogical(node *LogicalNode) (goqu.Expression, error)
	VisitComparison(node *ComparisonNode) (goqu.Expression, error)
}

// GoquWhereVisitor goqu where条件访问者 - 将查询树转换为goqu表达式
type GoquWhereVisitor struct{}

// NewGoquWhereVisitor 创建goqu访问者
// 返回:
//   - *GoquWhereVisitor: goqu访问者实例
func NewGoquWhereVisitor() *GoquWhereVisitor {
	return &GoquWhereVisitor{}
}

// VisitField 访问字段节点 - 处理简单的字段条件
// 参数:
//   - node: 字段节点
//
// 返回:
//   - goqu.Expression: goqu表达式
//   - error: 错误信息
func (v *GoquWhereVisitor) VisitField(node *FieldNode) (goqu.Expression, error) {
	if node.Field == "" {
		return nil, bizerrors.Errorf("field name cannot be empty")
	}

	// 简单的等值条件，如 {"age": 30} -> age = 30
	return goqu.C(node.Field).Eq(node.Value), nil
}

// VisitLogical 访问逻辑节点 - 处理逻辑操作符
// 参数:
//   - node: 逻辑节点
//
// 返回:
//   - goqu.Expression: goqu表达式
//   - error: 错误信息
func (v *GoquWhereVisitor) VisitLogical(node *LogicalNode) (goqu.Expression, error) {
	if len(node.Children) == 0 {
		return nil, bizerrors.Errorf("logical operator %s requires at least one child", node.Operator)
	}

	// 递归处理所有子节点
	expressions := make([]goqu.Expression, 0, len(node.Children))
	for _, child := range node.Children {
		expr, err := child.Accept(v)
		if err != nil {
			return nil, bizerrors.Errorf("failed to process child node: %w", err)
		}
		expressions = append(expressions, expr)
	}

	// 根据操作符类型组合表达式
	switch node.Operator {
	case query.LogicalOperatorAND:
		return goqu.And(expressions...), nil
	case query.LogicalOperatorOR:
		return goqu.Or(expressions...), nil
	case query.LogicalOperatorNOT:
		// NOT operator should have exactly one child
		if len(expressions) != 1 {
			return nil, bizerrors.Errorf("NOT operator requires exactly one child, got %d", len(expressions))
		}
		// Use goqu's Not() function to negate the expression
		return goqu.L("NOT ?", expressions[0]), nil
	default:
		return nil, bizerrors.Errorf("unsupported logical operator: %s", node.Operator)
	}
}

// VisitComparison 访问比较节点 - 处理比较操作符
// 参数:
//   - node: 比较节点
//
// 返回:
//   - goqu.Expression: goqu表达式
//   - error: 错误信息
func (v *GoquWhereVisitor) VisitComparison(node *ComparisonNode) (goqu.Expression, error) {
	if node.Field == "" {
		return nil, bizerrors.Errorf("field name cannot be empty")
	}
	column := goqu.C(node.Field)
	if len(node.OpAndValuePair) == 0 {
		return nil, bizerrors.Errorf("opAndValuePair cannot be empty")
	}

	exprList := make([]goqu.Expression, 0, len(node.OpAndValuePair))
	for op, value := range node.OpAndValuePair {
		expr, err := v.convertSingleExpr(op, value, column)
		if err != nil {
			return nil, err
		}
		exprList = append(exprList, expr)
	}
	return goqu.And(exprList...), nil
}

// operatorHandler 操作符处理器函数类型
type operatorHandler func(column exp.IdentifierExpression, value any) (goqu.Expression, error)

// operatorHandlers 操作符处理器映射 - 通过map替代switch-case以降低圈复杂度
var operatorHandlers = map[string]operatorHandler{
	query.FieldEquals:     handleEquals,
	query.FieldNot:        handleNot,
	query.FieldGt:         handleGt,
	query.FieldGte:        handleGte,
	query.FieldLt:         handleLt,
	query.FieldLte:        handleLte,
	query.FieldIn:         handleIn,
	query.FieldContains:   handleContains,
	query.FieldStartsWith: handleStartsWith,
	query.FieldEndsWith:   handleEndsWith,
}

// convertArrayValue 将任意类型转换为数组
func convertArrayValue(value any) ([]any, error) {
	v := reflect.ValueOf(value)
	if v.Kind() != reflect.Array && v.Kind() != reflect.Slice {
		return nil, bizerrors.Errorf("requires array value, got %T", value)
	}
	result := make([]any, v.Len())
	for i := 0; i < v.Len(); i++ {
		result[i] = v.Index(i).Interface()
	}
	return result, nil
}

// assertStringValue 断言值为字符串
func assertStringValue(value any) (string, error) {
	strValue, ok := value.(string)
	if !ok {
		return "", bizerrors.Errorf("requires string value, got %T", value)
	}
	return strValue, nil
}

// 比较操作符处理器
func handleEquals(column exp.IdentifierExpression, value any) (goqu.Expression, error) {
	return column.Eq(value), nil
}

func handleNot(column exp.IdentifierExpression, value any) (goqu.Expression, error) {
	return column.Neq(value), nil
}

func handleGt(column exp.IdentifierExpression, value any) (goqu.Expression, error) {
	return column.Gt(value), nil
}

func handleGte(column exp.IdentifierExpression, value any) (goqu.Expression, error) {
	return column.Gte(value), nil
}

func handleLt(column exp.IdentifierExpression, value any) (goqu.Expression, error) {
	return column.Lt(value), nil
}

func handleLte(column exp.IdentifierExpression, value any) (goqu.Expression, error) {
	return column.Lte(value), nil
}

// 数组操作符处理器
func handleIn(column exp.IdentifierExpression, value any) (goqu.Expression, error) {
	result, err := convertArrayValue(value)
	if err != nil {
		return nil, bizerrors.Errorf("in operator %w", err)
	}
	return column.In(result...), nil
}

// 字符串操作符处理器
func handleContains(column exp.IdentifierExpression, value any) (goqu.Expression, error) {
	strValue, err := assertStringValue(value)
	if err != nil {
		return nil, bizerrors.Errorf("contains operator %w", err)
	}
	return column.Like("%" + strValue + "%"), nil
}

func handleStartsWith(column exp.IdentifierExpression, value any) (goqu.Expression, error) {
	strValue, err := assertStringValue(value)
	if err != nil {
		return nil, bizerrors.Errorf("startsWith operator %w", err)
	}
	return column.Like(strValue + "%"), nil
}

func handleEndsWith(column exp.IdentifierExpression, value any) (goqu.Expression, error) {
	strValue, err := assertStringValue(value)
	if err != nil {
		return nil, bizerrors.Errorf("endsWith operator %w", err)
	}
	return column.Like("%" + strValue), nil
}

// convertSingleExpr 使用处理器映射将比较操作符转换为goqu表达式 - 圈复杂度已降低至2
func (v *GoquWhereVisitor) convertSingleExpr(
	op string,
	value any,
	column exp.IdentifierExpression,
) (goqu.Expression, error) {
	handler, exists := operatorHandlers[op]
	if !exists {
		return nil, bizerrors.Errorf("unsupported comparison operator: %s", op)
	}
	return handler(column, value)
}

// ConvertToGoquExpression 将查询节点转换为goqu表达式的便捷函数
// 参数:
//   - node: 查询节点
//
// 返回:
//   - goqu.Expression: goqu表达式
//   - error: 错误信息
func ConvertToGoquExpression(node QueryNode) (goqu.Expression, error) {
	if node == nil {
		return nil, bizerrors.Errorf("query node cannot be nil")
	}

	visitor := NewGoquWhereVisitor()
	return node.Accept(visitor)
}

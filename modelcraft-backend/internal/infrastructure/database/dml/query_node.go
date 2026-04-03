package dml

import (
	"fmt"
	"modelcraft/internal/domain/query"

	"github.com/doug-martin/goqu/v9"
)

// QueryNode 查询节点接口 - 组合模式的核心接口
type QueryNode interface {
	Accept(visitor QueryVisitor) (goqu.Expression, error)
}

// FieldNode 字段节点 - 处理简单的字段条件，如 {"age": 30}
type FieldNode struct {
	Field string
	Value interface{}
}

// NewFieldNode 创建字段节点
// 参数:
//   - field: 字段名称
//   - value: 字段值
//
// 返回:
//   - *FieldNode: 字段节点实例
func NewFieldNode(field string, value interface{}) *FieldNode {
	return &FieldNode{
		Field: field,
		Value: value,
	}
}

// Accept 实现访问者模式
// 参数:
//   - visitor: 查询访问者
//
// 返回:
//   - goqu.Expression: goqu表达式
//   - error: 错误信息
func (n *FieldNode) Accept(visitor QueryVisitor) (goqu.Expression, error) {
	return visitor.VisitField(n)
}

// LogicalNode 逻辑节点 - 处理逻辑操作符，如 {"$or": [...]}
type LogicalNode struct {
	Operator string      // $and, $or
	Children []QueryNode // 子节点
}

// NewLogicalNode 创建逻辑节点
// 参数:
//   - operator: 逻辑操作符（$and, $or）
//   - children: 子节点列表
//
// 返回:
//   - *LogicalNode: 逻辑节点实例
func NewLogicalNode(operator string, children []QueryNode) *LogicalNode {
	return &LogicalNode{
		Operator: operator,
		Children: children,
	}
}

// Accept 实现访问者模式
// 参数:
//   - visitor: 查询访问者
//
// 返回:
//   - goqu.Expression: goqu表达式
//   - error: 错误信息
func (n *LogicalNode) Accept(visitor QueryVisitor) (goqu.Expression, error) {
	return visitor.VisitLogical(n)
}

// EntryType 条目类型 - 表示操作符和值的映射关系
type EntryType map[string]any

// ComparisonNode 比较节点 - 处理比较操作符，如 {"age": {"gt": 18, "lt"20}}
type ComparisonNode struct {
	Field          string
	OpAndValuePair map[string]any // modelruntime.
}

// NewComparisonNode 创建比较节点
// 参数:
//   - field: 字段名称
//   - opAndValuePair: 操作符和值的映射
//
// 返回:
//   - *ComparisonNode: 比较节点实例
//   - error: 错误信息
func NewComparisonNode(field string, opAndValuePair EntryType) (*ComparisonNode, error) {
	for op := range opAndValuePair {
		err := ValidateOperator(op)
		if err != nil {
			return nil, err
		}
	}
	return &ComparisonNode{
		Field:          field,
		OpAndValuePair: opAndValuePair,
	}, nil
}

// Accept 实现访问者模式
// 参数:
//   - visitor: 查询访问者
//
// 返回:
//   - goqu.Expression: goqu表达式
//   - error: 错误信息
func (n *ComparisonNode) Accept(visitor QueryVisitor) (goqu.Expression, error) {
	return visitor.VisitComparison(n)
}

// ValidateOperator 验证操作符是否有效
// 参数:
//   - op: 操作符字符串
//
// 返回:
//   - error: 错误信息，如果操作符无效
func ValidateOperator(op string) error {
	if !query.IsLogicalOperator(op) && !query.IsComparisonOperator(op) {
		return fmt.Errorf("unsupported operator: %s", op)
	}
	return nil
}

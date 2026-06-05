package dml

import (
	"modelcraft/internal/domain/query"
	"modelcraft/pkg/bizerrors"

	"github.com/doug-martin/goqu/v9"
)

// QueryParser 查询解析器 - 将map[string]any转换为查询树
type QueryParser struct {
	maxDepth int // 最大递归深度，防止栈溢出
}

// NewQueryParser 创建查询解析器
// 返回:
//   - *QueryParser: 查询解析器实例
func NewQueryParser() *QueryParser {
	return &QueryParser{
		maxDepth: 10, // 默认最大深度为10
	}
}

// SetMaxDepth 设置最大递归深度
// 参数:
//   - depth: 最大递归深度
func (p *QueryParser) SetMaxDepth(depth int) {
	p.maxDepth = depth
}

// Parse 解析查询条件为查询树
// 参数:
//   - where: 查询条件映射
//
// 返回:
//   - QueryNode: 查询节点
//   - error: 错误信息
func (p *QueryParser) Parse(where map[string]any) (QueryNode, error) {
	return p.parseWithDepth(where, 0)
}

// parseWithDepth 带深度检查的递归解析
func (p *QueryParser) parseWithDepth(where map[string]any, depth int) (QueryNode, error) {
	if depth > p.maxDepth {
		return nil, bizerrors.Errorf("query nesting too deep, max depth is %d", p.maxDepth)
	}

	if len(where) == 0 {
		return nil, bizerrors.Errorf("empty where condition")
	}

	// 如果只有一个条件，直接处理
	if len(where) == 1 {
		for key, value := range where {
			return p.parseSingleCondition(key, value, depth)
		}
	}

	// 多个条件，默认使用 AND 逻辑
	children := make([]QueryNode, 0, len(where))
	for key, value := range where {
		child, err := p.parseSingleCondition(key, value, depth)
		if err != nil {
			return nil, bizerrors.Errorf("failed to parse condition %s: %w", key, err)
		}
		children = append(children, child)
	}

	return NewLogicalNode(query.LogicalOperatorAND, children), nil
}

// parseSingleCondition 解析单个条件
func (p *QueryParser) parseSingleCondition(key string, value interface{}, depth int) (QueryNode, error) {
	// 检查是否为逻辑操作符
	if query.IsLogicalOperator(key) {
		return p.parseLogicalOperator(key, value, depth)
	}

	// 检查值是否为复杂条件（包含比较操作符）
	if valueMap, ok := value.(map[string]interface{}); ok {
		return p.parseComplexCondition(key, valueMap, depth)
	}

	// 简单的字段条件
	return NewFieldNode(key, value), nil
}

// parseLogicalOperator 解析逻辑操作符
func (p *QueryParser) parseLogicalOperator(operator string, value interface{}, depth int) (QueryNode, error) {
	// NOT accepts either a single object { NOT: { field: ... } } or an array (Prisma-style).
	// AND / OR always require an array.
	// Normalise: if NOT receives a plain map, wrap it in a slice.
	if operator == string(query.LogicalOperatorNOT) {
		if valueMap, ok := value.(map[string]interface{}); ok {
			value = []interface{}{valueMap}
		}
	}

	// 逻辑操作符的值必须是数组
	valueSlice, ok := value.([]interface{})
	if !ok {
		return nil, bizerrors.Errorf("logical operator %s requires array value, got %T", operator, value)
	}

	if len(valueSlice) == 0 {
		return nil, bizerrors.Errorf("logical operator %s requires at least one condition", operator)
	}

	children := make([]QueryNode, 0, len(valueSlice))
	for i, item := range valueSlice {
		itemMap, ok := item.(map[string]interface{})
		if !ok {
			return nil, bizerrors.Errorf("logical operator %s item %d must be object, got %T", operator, i, item)
		}

		child, err := p.parseWithDepth(itemMap, depth+1)
		if err != nil {
			return nil, bizerrors.Errorf("failed to parse logical operator %s item %d: %w", operator, i, err)
		}
		children = append(children, child)
	}

	return NewLogicalNode(operator, children), nil
}

// parseComplexCondition 解析复杂条件（包含比较操作符）
func (p *QueryParser) parseComplexCondition(
	field string,
	conditions map[string]interface{},
	depth int,
) (QueryNode, error) {
	if len(conditions) == 0 {
		return nil, bizerrors.Errorf("empty conditions for field %s", field)
	}

	return NewComparisonNode(field, conditions)
}

// ParseQuery 便捷函数 - 解析查询并转换为goqu表达式
// 参数:
//   - where: 查询条件映射
//
// 返回:
//   - QueryNode: 查询节点
//   - error: 错误信息
func ParseQuery(where map[string]any) (QueryNode, error) {
	parser := NewQueryParser()
	return parser.Parse(where)
}

// ParseAndConvert 便捷函数 - 解析查询并直接转换为goqu表达式
// 参数:
//   - where: 查询条件映射
//
// 返回:
//   - goqu.Expression: goqu表达式
//   - error: 错误信息
func ParseAndConvert(where map[string]any) (goqu.Expression, error) {
	// 处理空条件
	if len(where) == 0 {
		return nil, bizerrors.Errorf("empty where condition")
	}

	// 解析为查询树
	node, err := ParseQuery(where)
	if err != nil {
		return nil, bizerrors.Errorf("failed to parse query: %w", err)
	}

	// 转换为goqu表达式
	expr, err := ConvertToGoquExpression(node)
	if err != nil {
		return nil, bizerrors.Errorf("failed to convert to goqu expression: %w", err)
	}

	return expr, nil
}

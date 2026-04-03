// Package query provides Prisma-style query condition builders for constructing GraphQL queries.
//
// This package implements a fluent API for building type-safe query conditions with support for:
//   - Logical operators (AND, OR, NOT)
//   - Comparison operators (equals, not, in, lt, lte, gt, gte)
//   - String operators (contains, startsWith, endsWith)
//   - Case-insensitive string matching
//   - Reserved keyword validation
//
// # Prisma-Style API
//
// The API follows Prisma ORM conventions for query conditions:
//   - Logical operators use uppercase: AND, OR, NOT
//   - Field operators use camelCase: equals, in, contains, startsWith, etc.
//   - All operator names are reserved and cannot be used as field names
//
// # Semantic Distinction: not vs NOT
//
// This package maintains a clear distinction between field-level and logical-level negation:
//   - not (lowercase): Field-level operator meaning "field does not equal value"
//   - NOT (uppercase): Logical operator meaning "negate the entire condition block"
//
// Examples:
//
//	Field("status").Not("draft")  // Field-level: status != "draft"
//	LogicalNot(Field("status").Eq("draft"))  // Logical: NOT (status = "draft")
//
// # Examples
//
// Simple equality:
//
//	cond := Field("username").Eq("john")
//	// Produces: {"username": {"equals": "john"}}
//
// Logical combinations:
//
//	cond := And(
//	    Field("age").Gt(18),
//	    Field("status").Eq("active"),
//	)
//	// Produces: {"AND": [{"age": {"gt": 18}}, {"status": {"equals": "active"}}]}
//
// Logical NOT:
//
//	cond := LogicalNot(Field("title").Contains("SQL"))
//	// Produces: {"NOT": {"title": {"contains": "SQL"}}}
//
// Replacing notIn with NOT:
//
//	// Old (removed): Field("status").NotIn("draft", "archived")
//	// New: LogicalNot(Field("status").In("draft", "archived"))
//	cond := LogicalNot(Field("status").In("draft", "archived"))
//	// Produces: {"NOT": {"status": {"in": ["draft", "archived"]}}}
//
// String matching:
//
//	cond := Field("email").Contains("@example.com")
//	// Produces: {"email": {"contains": "@example.com"}}
//
// # Reserved Keywords
//
// All operator names are reserved keywords and cannot be used as field names:
//   - Logical: AND, OR, NOT
//   - Comparison: equals, not, in, lt, lte, gt, gte
//   - String: contains, startsWith, endsWith
//   - Modifier: mode
//
// Use GetReservedKeywords() to get the complete list and IsReservedKeyword() to validate field names.
package query

import "strings"

// Condition 表示查询条件的类型，用于构建GraphQL查询条件
type Condition map[string]any

// ============================================================================
// Condition 类型方法
// ============================================================================

// And 将当前条件与其他条件进行AND组合
func (c Condition) And(other ...Condition) Condition {
	conditions := make([]Condition, 0, len(other)+1)
	conditions = append(conditions, c)
	conditions = append(conditions, other...)
	return And(conditions...)
}

// Or 将当前条件与其他条件进行OR组合
func (c Condition) Or(other ...Condition) Condition {
	conditions := make([]Condition, 0, len(other)+1)
	conditions = append(conditions, c)
	conditions = append(conditions, other...)
	return Or(conditions...)
}

// Not 对当前条件进行逻辑否定
func (c Condition) Not() Condition {
	return Condition{LogicalOperatorNOT: c}
}

// ToMap 将Condition转换为map[string]any，用于兼容性
func (c Condition) ToMap() map[string]any {
	return map[string]any(c)
}

// IsEmpty 检查条件是否为空
func (c Condition) IsEmpty() bool {
	return len(c) == 0
}

// Clone 创建条件的副本
func (c Condition) Clone() Condition {
	clone := make(Condition, len(c))
	for k, v := range c {
		clone[k] = v
	}
	return clone
}

// Merge 合并其他条件到当前条件（会修改当前条件）
func (c Condition) Merge(other Condition) Condition {
	for k, v := range other {
		c[k] = v
	}
	return c
}

// 字段条件比较操作符常量
const (
	// 通用比较操作符
	// FieldEquals 相等比较
	FieldEquals = "equals"
	// FieldNot 否定比较
	FieldNot = "not"
	// FieldIn 包含在列表中
	FieldIn = "in"
	// 字符串字段特有操作符
	// FieldContains 包含子字符串
	FieldContains = "contains"
	// FieldStartsWith 以字符串开头
	FieldStartsWith = "startsWith"
	// FieldEndsWith 以字符串结尾
	FieldEndsWith = "endsWith"
	// FieldMode 查询模式（区分大小写等）
	FieldMode = "mode"
	// 数值字段特有操作符
	// FieldLt 小于
	FieldLt = "lt"
	// FieldLte 小于或等于
	FieldLte = "lte"
	// FieldGt 大于
	FieldGt = "gt"
	// FieldGte 大于或等于
	FieldGte = "gte"
	// 查询模式枚举值
	// QueryModeDefault 默认查询模式
	QueryModeDefault = "default"
	// QueryModeInsensitive 不区分大小写的查询模式
	QueryModeInsensitive = "insensitive"
	// QueryModeName 查询模式枚举类型名
	QueryModeName = "QueryMode"
	// 逻辑操作符
	// LogicalOperatorAND AND逻辑操作符（Prisma风格：大写）
	LogicalOperatorAND = "AND"
	// LogicalOperatorOR OR逻辑操作符（Prisma风格：大写）
	LogicalOperatorOR = "OR"
	// LogicalOperatorNOT NOT逻辑操作符（Prisma风格：大写）- 用于否定整个条件块
	LogicalOperatorNOT = "NOT"
)

// IsLogicalOperator 检查是否为逻辑操作符（AND、OR或NOT）
func IsLogicalOperator(op string) bool {
	return op == LogicalOperatorAND || op == LogicalOperatorOR || op == LogicalOperatorNOT
}

// IsComparisonOperator 检查是否为比较操作符
func IsComparisonOperator(op string) bool {
	switch op {
	case FieldEquals, FieldNot,
		FieldGt, FieldGte,
		FieldLt, FieldLte,
		FieldContains, FieldStartsWith, FieldEndsWith,
		FieldIn:
		return true
	default:
		return false
	}
}

// ============================================================================
// 构造函数和工厂方法
// ============================================================================

// NewCondition 创建一个新的空条件
func NewCondition() Condition {
	return make(Condition)
}

// NewConditionFrom 从map[string]any创建条件
func NewConditionFrom(m map[string]any) Condition {
	return Condition(m)
}

// Field 创建字段条件构建器
func Field(name string) *FieldBuilder {
	return &FieldBuilder{fieldName: name}
}

// FieldBuilder 字段条件构建器
type FieldBuilder struct {
	fieldName string
}

// Eq 字段等于条件
func (fb *FieldBuilder) Eq(val any) Condition {
	return Condition{fb.fieldName: Eq(val)}
}

// Not 字段不等于条件
func (fb *FieldBuilder) Not(val any) Condition {
	return Condition{fb.fieldName: Not(val)}
}

// In 字段包含条件
func (fb *FieldBuilder) In(vals ...any) Condition {
	return Condition{fb.fieldName: In(vals...)}
}

// Gt 字段大于条件
func (fb *FieldBuilder) Gt(val any) Condition {
	return Condition{fb.fieldName: Gt(val)}
}

// Gte 字段大于等于条件
func (fb *FieldBuilder) Gte(val any) Condition {
	return Condition{fb.fieldName: Gte(val)}
}

// Lt 字段小于条件
func (fb *FieldBuilder) Lt(val any) Condition {
	return Condition{fb.fieldName: Lt(val)}
}

// Lte 字段小于等于条件
func (fb *FieldBuilder) Lte(val any) Condition {
	return Condition{fb.fieldName: Lte(val)}
}

// Contains 字段包含子字符串条件
func (fb *FieldBuilder) Contains(val string) Condition {
	return Condition{fb.fieldName: Contains(val)}
}

// StartsWith 字段以字符串开头条件
func (fb *FieldBuilder) StartsWith(val string) Condition {
	return Condition{fb.fieldName: StartsWith(val)}
}

// EndsWith 字段以字符串结尾条件
func (fb *FieldBuilder) EndsWith(val string) Condition {
	return Condition{fb.fieldName: EndsWith(val)}
}

// Between 字段范围条件
func (fb *FieldBuilder) Between(min, max any) Condition {
	return Condition{fb.fieldName: Between(min, max)}
}

// IsNull 字段为空条件
func (fb *FieldBuilder) IsNull() Condition {
	return Condition{fb.fieldName: IsNull()}
}

// IsNotNull 字段不为空条件
func (fb *FieldBuilder) IsNotNull() Condition {
	return Condition{fb.fieldName: IsNotNull()}
}

// ============================================================================
// 便捷条件生成函数 - 类似ORM客户端的使用方式
// ============================================================================

// Eq 生成相等条件 - 等价于 field = value
func Eq(val any) Condition {
	return Condition{FieldEquals: val}
}

// Not 生成不等条件 - 等价于 field != value
func Not(val any) Condition {
	return Condition{FieldNot: val}
}

// In 生成包含条件 - 等价于 field IN (values...)
func In(vals ...any) Condition {
	return Condition{FieldIn: vals}
}

// Gt 生成大于条件 - 等价于 field > value
func Gt(val any) Condition {
	return Condition{FieldGt: val}
}

// Gte 生成大于等于条件 - 等价于 field >= value
func Gte(val any) Condition {
	return Condition{FieldGte: val}
}

// Lt 生成小于条件 - 等价于 field < value
func Lt(val any) Condition {
	return Condition{FieldLt: val}
}

// Lte 生成小于等于条件 - 等价于 field <= value
func Lte(val any) Condition {
	return Condition{FieldLte: val}
}

// Contains 生成包含子字符串条件 - 等价于 field LIKE '%value%'
func Contains(val string) Condition {
	return Condition{FieldContains: val}
}

// StartsWith 生成以字符串开头条件 - 等价于 field LIKE 'value%'
func StartsWith(val string) Condition {
	return Condition{FieldStartsWith: val}
}

// EndsWith 生成以字符串结尾条件 - 等价于 field LIKE '%value'
func EndsWith(val string) Condition {
	return Condition{FieldEndsWith: val}
}

// ContainsInsensitive 生成不区分大小写的包含条件
func ContainsInsensitive(val string) Condition {
	return Condition{
		FieldContains: val,
		FieldMode:     QueryModeInsensitive,
	}
}

// StartsWithInsensitive 生成不区分大小写的开头匹配条件
func StartsWithInsensitive(val string) Condition {
	return Condition{
		FieldStartsWith: val,
		FieldMode:       QueryModeInsensitive,
	}
}

// EndsWithInsensitive 生成不区分大小写的结尾匹配条件
func EndsWithInsensitive(val string) Condition {
	return Condition{
		FieldEndsWith: val,
		FieldMode:     QueryModeInsensitive,
	}
}

// ============================================================================
// 逻辑操作符便捷函数
// ============================================================================

// And 生成AND逻辑条件 - 所有条件都必须满足
func And(conditions ...Condition) Condition {
	return Condition{LogicalOperatorAND: conditions}
}

// Or 生成OR逻辑条件 - 任一条件满足即可
func Or(conditions ...Condition) Condition {
	return Condition{LogicalOperatorOR: conditions}
}

// LogicalNot 生成NOT逻辑条件 - 否定整个条件块
// 参数:
//   - condition: 要否定的条件
//
// 返回:
//   - Condition: NOT条件，会匹配不满足指定条件的所有记录
//
// 示例:
//
//	LogicalNot(Field("status").Eq("draft"))  // 等价于 { NOT: { status: { equals: "draft" } } }
//	LogicalNot(Field("title").Contains("SQL"))  // 等价于 { NOT: { title: { contains: "SQL" } } }
func LogicalNot(condition Condition) Condition {
	return Condition{LogicalOperatorNOT: condition}
}

// ============================================================================
// 范围查询便捷函数
// ============================================================================

// Between 生成范围查询条件 - 等价于 field >= min AND field <= max
func Between(min, max any) Condition {
	return And(
		Gte(min),
		Lte(max),
	)
}

// NotBetween 生成排除范围条件 - 等价于 field < min OR field > max
func NotBetween(min, max any) Condition {
	return Or(
		Lt(min),
		Gt(max),
	)
}

// ============================================================================
// 空值检查便捷函数
// ============================================================================

// IsNull 生成空值检查条件
func IsNull() Condition {
	return Eq(nil)
}

// IsNotNull 生成非空值检查条件
func IsNotNull() Condition {
	return Not(nil)
}

// ============================================================================
// 保留关键字验证 - Reserved Keywords Validation
// ============================================================================

// reservedKeywords 所有查询操作符关键字列表，这些关键字不能用作字段名
// 保留这些关键字可以防止字段名与查询操作符冲突，确保查询解析的正确性
var reservedKeywords = []string{
	// 逻辑操作符 - Logical Operators
	"AND", "OR", "NOT",
	// 通用比较操作符 - Common Comparison Operators
	"equals", "not", "in",
	// 数值比较操作符 - Numeric Comparison Operators
	"lt", "lte", "gt", "gte",
	// 字符串操作符 - String Operators
	"contains", "startsWith", "endsWith",
	// 查询模式 - Query Mode
	"mode",
}

// GetReservedKeywords 返回所有保留关键字列表
// 这些关键字是Prisma风格的查询操作符，不能用作字段名
func GetReservedKeywords() []string {
	// 返回副本以防止外部修改
	keywords := make([]string, len(reservedKeywords))
	copy(keywords, reservedKeywords)
	return keywords
}

// IsReservedKeyword 检查给定的名称是否为保留关键字（不区分大小写）
// 参数:
//   - name: 要检查的字段名
//
// 返回:
//   - bool: 如果是保留关键字返回true，否则返回false
func IsReservedKeyword(name string) bool {
	if name == "" {
		return false
	}

	// 转换为小写进行比较（除了AND, OR, NOT需要特殊处理）
	nameLower := strings.ToLower(name)
	nameUpper := strings.ToUpper(name)

	for _, keyword := range reservedKeywords {
		keywordLower := strings.ToLower(keyword)
		// 对于大写关键字（AND, OR, NOT），也检查大写形式
		if keywordLower == nameLower || keyword == nameUpper {
			return true
		}
	}

	return false
}

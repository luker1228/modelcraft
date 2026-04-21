package dml

import (
	"modelcraft/internal/domain/query"
	"testing"
	"strings"

	"github.com/doug-martin/goqu/v9"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGoquWhereVisitor_VisitField(t *testing.T) {
	visitor := NewGoquWhereVisitor()

	tests := []struct {
		name        string
		field       string
		value       interface{}
		expectError bool
		expectedSQL string
	}{
		{
			name:        "simple field equality",
			field:       "name",
			value:       "John",
			expectError: false,
			expectedSQL: `"name" = ?`,
		},
		{
			name:        "numeric field equality",
			field:       "age",
			value:       30,
			expectError: false,
			expectedSQL: `"age" = ?`,
		},
		{
			name:        "empty field name",
			field:       "",
			value:       "value",
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			node := NewFieldNode(tt.field, tt.value)
			expr, err := visitor.VisitField(node)

			if tt.expectError {
				assert.Error(t, err)
				assert.Nil(t, expr)
			} else {
				require.NoError(t, err)
				require.NotNil(t, expr)

				// 验证生成的SQL
				sql, _, err := goqu.From("test").Where(expr).Prepared(true).ToSQL()
				require.NoError(t, err)
				assert.Contains(t, sql, tt.expectedSQL)
			}
		})
	}
}

func TestGoquWhereVisitor_VisitLogical(t *testing.T) {
	visitor := NewGoquWhereVisitor()

	tests := []struct {
		name        string
		operator    string
		children    []QueryNode
		expectError bool
		expectedSQL string
	}{
		{
			name:     "AND operator with two children",
			operator: query.LogicalOperatorAND,
			children: []QueryNode{
				NewFieldNode("name", "John"),
				NewFieldNode("age", 30),
			},
			expectError: false,
			expectedSQL: `(("name" = ?) AND ("age" = ?))`,
		},
		{
			name:     "OR operator with two children",
			operator: query.LogicalOperatorOR,
			children: []QueryNode{
				NewFieldNode("name", "John"),
				NewFieldNode("name", "Jane"),
			},
			expectError: false,
			expectedSQL: `(("name" = ?) OR ("name" = ?))`,
		},
		{
			name:        "NOT operator with one child",
			operator:    query.LogicalOperatorNOT,
			children:    []QueryNode{NewFieldNode("status", "active")},
			expectError: false,
			expectedSQL: `NOT ("status" = ?)`,
		},
		{
			name:        "NOT operator with multiple children should error",
			operator:    query.LogicalOperatorNOT,
			children:    []QueryNode{NewFieldNode("a", 1), NewFieldNode("b", 2)},
			expectError: true,
		},
		{
			name:        "empty children list",
			operator:    query.LogicalOperatorAND,
			children:    []QueryNode{},
			expectError: true,
		},
		{
			name:        "unsupported operator",
			operator:    "UNSUPPORTED",
			children:    []QueryNode{NewFieldNode("field", "value")},
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			node := NewLogicalNode(tt.operator, tt.children)
			expr, err := visitor.VisitLogical(node)

			if tt.expectError {
				assert.Error(t, err)
				assert.Nil(t, expr)
			} else {
				require.NoError(t, err)
				require.NotNil(t, expr)

				sql, _, err := goqu.From("test").Where(expr).Prepared(true).ToSQL()
				require.NoError(t, err)
				assert.Contains(t, sql, tt.expectedSQL)
			}
		})
	}
}

func TestGoquWhereVisitor_VisitComparison(t *testing.T) {
	visitor := NewGoquWhereVisitor()

	tests := []struct {
		name        string
		field       string
		opAndValue  map[string]interface{}
		expectError bool
		expectedSQL string
		allContains []string // all substrings must appear (order-independent)
	}{
		{
			name:  "equals operator",
			field: "name",
			opAndValue: map[string]interface{}{
				query.FieldEquals: "John",
			},
			expectError: false,
			expectedSQL: `"name" = ?`,
		},
		{
			name:  "not operator",
			field: "status",
			opAndValue: map[string]interface{}{
				query.FieldNot: "active",
			},
			expectError: false,
			expectedSQL: `"status" != ?`,
		},
		{
			name:  "greater than operator",
			field: "age",
			opAndValue: map[string]interface{}{
				query.FieldGt: 18,
			},
			expectError: false,
			expectedSQL: `"age" > ?`,
		},
		{
			name:  "less than operator",
			field: "age",
			opAndValue: map[string]interface{}{
				query.FieldLt: 65,
			},
			expectError: false,
			expectedSQL: `"age" < ?`,
		},
		{
			name:  "in operator with array",
			field: "status",
			opAndValue: map[string]interface{}{
				query.FieldIn: []interface{}{"active", "pending"},
			},
			expectError: false,
			expectedSQL: `"status" IN (?, ?)`,
		},
		{
			name:  "contains operator",
			field: "email",
			opAndValue: map[string]interface{}{
				query.FieldContains: "@example.com",
			},
			expectError: false,
			expectedSQL: `"email" LIKE ?`,
		},
		{
			name:  "startsWith operator",
			field: "name",
			opAndValue: map[string]interface{}{
				query.FieldStartsWith: "John",
			},
			expectError: false,
			expectedSQL: `"name" LIKE ?`,
		},
		{
			name:  "endsWith operator",
			field: "domain",
			opAndValue: map[string]interface{}{
				query.FieldEndsWith: ".com",
			},
			expectError: false,
			expectedSQL: `"domain" LIKE ?`,
		},
		{
			name:  "multiple operators on same field",
			field: "age",
			opAndValue: map[string]interface{}{
				query.FieldGt: 18,
				query.FieldLt: 65,
			},
			expectError: false,
			// map iteration order is non-deterministic, verify both conditions exist
			allContains: []string{`"age" > ?`, `"age" < ?`},
		},
		{
			name:  "empty field name",
			field: "",
			opAndValue: map[string]interface{}{
				query.FieldEquals: "value",
			},
			expectError: true,
		},
		{
			name:        "empty opAndValue",
			field:       "field",
			opAndValue:  map[string]interface{}{},
			expectError: true,
		},
		{
			name:  "unsupported operator",
			field: "field",
			opAndValue: map[string]interface{}{
				"unsupported": "value",
			},
			expectError: true,
		},
		{
			name:  "in operator with non-array value",
			field: "status",
			opAndValue: map[string]interface{}{
				query.FieldIn: "invalid",
			},
			expectError: true,
		},
		{
			name:  "string operator with non-string value",
			field: "email",
			opAndValue: map[string]interface{}{
				query.FieldContains: 123,
			},
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			node, err := NewComparisonNode(tt.field, tt.opAndValue)
			if err != nil {
				if tt.expectError {
					return
				}
				require.NoError(t, err)
			}

			expr, err := visitor.VisitComparison(node)

			if tt.expectError {
				assert.Error(t, err)
				assert.Nil(t, expr)
			} else {
				require.NoError(t, err)
				require.NotNil(t, expr)

				sql, _, err := goqu.From("test").Where(expr).Prepared(true).ToSQL()
				require.NoError(t, err)
				if len(tt.allContains) > 0 {
					for _, s := range tt.allContains {
						assert.Contains(t, sql, s)
					}
				} else {
					assert.Contains(t, sql, tt.expectedSQL)
				}
			}
		})
	}
}

func TestGoquWhereVisitor_OperatorHandlers(t *testing.T) {
	visitor := NewGoquWhereVisitor()
	column := goqu.C("test_field")

	tests := []struct {
		name        string
		operator    string
		value       interface{}
		expectError bool
		expectedSQL string
	}{
		{
			name:        "equals handler",
			operator:    query.FieldEquals,
			value:       "test",
			expectError: false,
			expectedSQL: `"test_field" = ?`,
		},
		{
			name:        "not handler",
			operator:    query.FieldNot,
			value:       "excluded",
			expectError: false,
			expectedSQL: `"test_field" != ?`,
		},
		{
			name:        "gt handler",
			operator:    query.FieldGt,
			value:       10,
			expectError: false,
			expectedSQL: `"test_field" > ?`,
		},
		{
			name:        "gte handler",
			operator:    query.FieldGte,
			value:       10,
			expectError: false,
			expectedSQL: `"test_field" >= ?`,
		},
		{
			name:        "lt handler",
			operator:    query.FieldLt,
			value:       100,
			expectError: false,
			expectedSQL: `"test_field" < ?`,
		},
		{
			name:        "lte handler",
			operator:    query.FieldLte,
			value:       100,
			expectError: false,
			expectedSQL: `"test_field" <= ?`,
		},
		{
			name:        "in handler with valid array",
			operator:    query.FieldIn,
			value:       []interface{}{1, 2, 3},
			expectError: false,
			expectedSQL: `"test_field" IN (?, ?, ?)`,
		},
		{
			name:        "contains handler",
			operator:    query.FieldContains,
			value:       "substring",
			expectError: false,
			expectedSQL: `"test_field" LIKE ?`,
		},
		{
			name:        "startsWith handler",
			operator:    query.FieldStartsWith,
			value:       "prefix",
			expectError: false,
			expectedSQL: `"test_field" LIKE ?`,
		},
		{
			name:        "endsWith handler",
			operator:    query.FieldEndsWith,
			value:       "suffix",
			expectError: false,
			expectedSQL: `"test_field" LIKE ?`,
		},
		{
			name:        "unsupported operator",
			operator:    "invalid_operator",
			value:       "value",
			expectError: true,
		},
		{
			name:        "in handler with non-array",
			operator:    query.FieldIn,
			value:       "not_an_array",
			expectError: true,
		},
		{
			name:        "string handler with non-string",
			operator:    query.FieldContains,
			value:       123,
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			expr, err := visitor.convertSingleExpr(tt.operator, tt.value, column)

			if tt.expectError {
				assert.Error(t, err)
				assert.Nil(t, expr)
			} else {
				require.NoError(t, err)
				require.NotNil(t, expr)

				sql, _, err := goqu.From("test").Where(expr).Prepared(true).ToSQL()
				require.NoError(t, err)
				assert.Contains(t, sql, tt.expectedSQL)
			}
		})
	}
}

func TestConvertToGoquExpression(t *testing.T) {
	tests := []struct {
		name        string
		node        QueryNode
		expectError bool
		expectedSQL string
	}{
		{
			name:        "field node",
			node:        NewFieldNode("name", "John"),
			expectError: false,
			expectedSQL: `"name" = ?`,
		},
		{
			name: "logical AND node",
			node: NewLogicalNode(query.LogicalOperatorAND, []QueryNode{
				NewFieldNode("name", "John"),
				NewFieldNode("age", 30),
			}),
			expectError: false,
			expectedSQL: `(("name" = ?) AND ("age" = ?))`,
		},
		{
			name: "comparison node",
			node: func() QueryNode {
				node, _ := NewComparisonNode("age", map[string]interface{}{
					query.FieldGt: 18,
					query.FieldLt: 65,
				})
				return node
			}(),
			expectError: false,
			expectedSQL: ``,
		},
		{
			name:        "nil node",
			node:        nil,
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			expr, err := ConvertToGoquExpression(tt.node)

			if tt.expectError {
				assert.Error(t, err)
				assert.Nil(t, expr)
			} else {
				require.NoError(t, err)
				require.NotNil(t, expr)

				sql, _, err := goqu.From("test").Where(expr).Prepared(true).ToSQL()
				require.NoError(t, err)
				if tt.name == "comparison node" {
					opt1 := `(("age" > ?) AND ("age" < ?))`
					opt2 := `(("age" < ?) AND ("age" > ?))`
					assert.True(t, strings.Contains(sql, opt1) || strings.Contains(sql, opt2))
					return
				}

				assert.Contains(t, sql, tt.expectedSQL)
			}
		})
	}
}

func TestHelperFunctions(t *testing.T) {
	tests := []struct {
		name        string
		value       interface{}
		expectError bool
	}{
		{
			name:        "convert array value with valid slice",
			value:       []interface{}{1, 2, 3},
			expectError: false,
		},
		{
			name:        "convert array value with array",
			value:       [3]interface{}{1, 2, 3},
			expectError: false,
		},
		{
			name:        "convert array value with non-array",
			value:       "not_an_array",
			expectError: true,
		},
		{
			name:        "assert string value with string",
			value:       "valid_string",
			expectError: false,
		},
		{
			name:        "assert string value with non-string",
			value:       123,
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			switch tt.name {
			case "convert array value with valid slice", "convert array value with array":
				result, err := convertArrayValue(tt.value)
				if tt.expectError {
					assert.Error(t, err)
					assert.Nil(t, result)
					return
				}
				require.NoError(t, err)
				assert.NotNil(t, result)
				assert.Len(t, result, 3)
			case "assert string value with string", "assert string value with non-string":
				result, err := assertStringValue(tt.value)
				if tt.expectError {
					assert.Error(t, err)
					assert.Equal(t, "", result)
					return
				}
				require.NoError(t, err)
				assert.Equal(t, "valid_string", result)
			}
		})
	}
}

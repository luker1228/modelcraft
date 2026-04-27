package dml

import (
	"context"
	"log"
	"modelcraft/internal/domain/modelruntime"
	"testing"

	"github.com/doug-martin/goqu/v9"
	_ "github.com/doug-martin/goqu/v9/dialect/mysql"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestQueryParser_SimpleConditions(t *testing.T) {
	tests := []struct {
		name     string
		where    map[string]any
		expected string // 预期的SQL片段
	}{
		{
			name: "simple equality",
			where: map[string]any{
				"name": "John",
			},
			expected: `SELECT * FROM "test" WHERE ("name" = ?)`,
		},
		{
			name: "multiple fields with AND",
			where: map[string]any{
				"name": "John",
				"age":  30,
			},
			expected: "", // map iteration order is not stable for multi-field AND
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			node, err := ParseQuery(tt.where)
			require.NoError(t, err)
			require.NotNil(t, node)

			expr, err := ConvertToGoquExpression(node)
			require.NoError(t, err)
			require.NotNil(t, expr)
			sql, params, err := goqu.From("test").Where(expr).Prepared(true).ToSQL()
			if err != nil {
				t.Logf("sql=%s params=%v, err=%+v", sql, params, err)
			}
			t.Logf("sql=%s params=%v", sql, params)

			if tt.name == "multiple fields with AND" {
				sql1 := `SELECT * FROM "test" WHERE (("name" = ?) AND ("age" = ?))`
				sql2 := `SELECT * FROM "test" WHERE (("age" = ?) AND ("name" = ?))`
				require.Contains(t, []string{sql1, sql2}, sql, "SQL should be logically equivalent")
				if sql == sql1 {
					assert.Equal(t, "John", params[0])
					assert.Equal(t, int64(30), params[1])
				} else {
					assert.Equal(t, int64(30), params[0])
					assert.Equal(t, "John", params[1])
				}
				return
			}

			assert.Equal(t, tt.expected, sql)
		})
	}
}

func TestQueryParser_LogicalOperators(t *testing.T) {
	tests := []struct {
		name  string
		where map[string]any
	}{
		{
			name: "OR operator",
			where: map[string]any{
				"$or": []interface{}{
					map[string]interface{}{"name": "John"},
					map[string]interface{}{"name": "Jane"},
				},
			},
		},
		{
			name: "AND operator",
			where: map[string]any{
				"$and": []interface{}{
					map[string]interface{}{"age": 30},
					map[string]interface{}{"status": "active"},
				},
			},
		},
		{
			name: "complex nested OR",
			where: map[string]any{
				"age": 30,
				"$or": []interface{}{
					map[string]interface{}{"name": "John1"},
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			node, err := ParseQuery(tt.where)
			require.NoError(t, err)
			require.NotNil(t, node)

			expr, err := ConvertToGoquExpression(node)
			require.NoError(t, err)
			require.NotNil(t, expr)
		})
	}
}

func TestQueryParser_ComparisonOperators(t *testing.T) {
	tests := []struct {
		name  string
		where map[string]any
	}{
		{
			name: "greater than",
			where: map[string]any{
				"age": map[string]interface{}{
					"gt": 18,
				},
			},
		},
		{
			name: "less than or equal",
			where: map[string]any{
				"age": map[string]interface{}{
					"lte": 65,
				},
			},
		},
		{
			name: "in operator",
			where: map[string]any{
				"status": map[string]interface{}{
					"in": []interface{}{"active", "pending"},
				},
			},
		},
		{
			name: "not equal",
			where: map[string]any{
				"name": map[string]interface{}{
					"not": "admin",
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			node, err := ParseQuery(tt.where)
			require.NoError(t, err)
			require.NotNil(t, node)

			expr, err := ConvertToGoquExpression(node)
			require.NoError(t, err)
			require.NotNil(t, expr)
		})
	}
}

func TestQueryParser_ErrorCases(t *testing.T) {
	tests := []struct {
		name        string
		where       map[string]any
		expectError bool
	}{
		{
			name:        "empty where",
			where:       map[string]any{},
			expectError: true,
		},
		{
			name: "invalid operator",
			where: map[string]any{
				"age": map[string]interface{}{
					"$invalid": 30,
				},
			},
			expectError: true,
		},
		{
			name: "OR with non-array value",
			where: map[string]any{
				"OR": "invalid",
			},
			expectError: true,
		},
		{
			name: "IN with non-array value",
			where: map[string]any{
				"status": map[string]interface{}{
					"$in": "invalid",
				},
			},
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := ParseQuery(tt.where)
			if tt.expectError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestSqlMapper_ComplexQuery(t *testing.T) {
	// 测试原始示例：{"age":30, "OR":[{"name":"John1"}]}
	input := &modelruntime.FindUniqueInput{
		TableName: "users",
		Where: map[string]any{
			"age": 30,
			"OR": []interface{}{
				map[string]interface{}{"name": "John1"},
			},
		},
	}

	sql, args, err := convertFindUniqueInputToSQL(context.Background(), input)
	require.NoError(t, err)
	require.NotEmpty(t, sql)
	require.NotEmpty(t, args)

	log.Printf("Complex query SQL: %s", sql)
	log.Printf("Complex query Args: %v", args)

	// 验证SQL包含预期的元素
	assert.Contains(t, sql, "SELECT")
	assert.Contains(t, sql, "users")
	assert.Contains(t, sql, "WHERE")
}

func TestSqlMapper_BackwardCompatibility(t *testing.T) {
	// 测试向后兼容性 - 简单的map[string]any应该仍然工作
	input := &modelruntime.FindUniqueInput{
		TableName: "users",
		Where: map[string]any{
			"name": "luke",
			"age":  11,
		},
	}

	sql, args, err := convertFindUniqueInputToSQL(context.Background(), input)
	require.NoError(t, err)
	require.NotEmpty(t, sql)
	require.NotEmpty(t, args)

	log.Printf("Simple query SQL: %s", sql)
	log.Printf("Simple query Args: %v", args)
}

func TestContainsComplexOperators(t *testing.T) {
	tests := []struct {
		name     string
		where    map[string]any
		expected bool
	}{
		{
			name: "simple conditions",
			where: map[string]any{
				"name": "John",
				"age":  30,
			},
			expected: false,
		},
		{
			name: "with OR operator",
			where: map[string]any{
				"OR": []interface{}{
					map[string]interface{}{"name": "John"},
				},
			},
			expected: true,
		},
		{
			name: "with comparison operator",
			where: map[string]any{
				"age": map[string]interface{}{
					"gt": 18,
				},
			},
			expected: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := containsComplexOperators(tt.where)
			assert.Equal(t, tt.expected, result)
		})
	}
}

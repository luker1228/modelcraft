package dml

import (
	"context"
	"fmt"
	"modelcraft/internal/domain/modelruntime"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestConvertWhereToExpression 测试where条件转换
func TestConvertWhereToExpression(t *testing.T) {
	tests := []struct {
		name      string
		where     map[string]any
		expectErr bool
		errMsg    string
	}{
		{
			name: "简单条件",
			where: map[string]any{
				"id":   "123",
				"name": "test",
			},
			expectErr: false,
		},
		{
			name: "复杂条件-AND操作符",
			where: map[string]any{
				"$and": []any{
					map[string]any{"id": "123"},
					map[string]any{"name": "test"},
				},
			},
			expectErr: false,
		},
		{
			name: "复杂条件-比较操作符",
			where: map[string]any{
				"age": map[string]interface{}{
					"$gt": 18,
				},
			},
			expectErr: false,
		},
		{
			name:      "空条件",
			where:     map[string]any{},
			expectErr: true,
			errMsg:    "where condition cannot be empty",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			expr, err := convertWhereToExpression(tt.where)
			if tt.expectErr {
				require.Error(t, err)
				if tt.errMsg != "" {
					assert.Contains(t, err.Error(), tt.errMsg)
				}
			} else {
				require.NoError(t, err)
				assert.NotNil(t, expr)
			}
		})
	}
}

// TestConvertFindUniqueInputToSQL 测试FindUnique转SQL
func TestConvertFindUniqueInputToSQL(t *testing.T) {
	ctx := context.Background()

	tests := []struct {
		name      string
		input     *modelruntime.FindUniqueInput
		expectErr bool
		errMsg    string
		checkSQL  func(t *testing.T, sql string, args []any)
	}{
		{
			name: "简单where条件",
			input: &modelruntime.FindUniqueInput{
				TableName: "users",
				Where: map[string]any{
					"id": "123",
				},
			},
			expectErr: false,
			checkSQL: func(t *testing.T, sql string, args []any) {
				assert.Contains(t, sql, "SELECT * FROM `users`")
				assert.Contains(t, sql, "WHERE")
				assert.Contains(t, sql, "`id`")
				assert.Equal(t, 1, len(args))
				assert.Equal(t, "123", args[0])
			},
		},
		{
			name: "指定Selection字段",
			input: &modelruntime.FindUniqueInput{
				TableName: "users",
				Where: map[string]any{
					"id": "123",
				},
				Selection: &modelruntime.Selection{
					FieldNames: map[string]bool{
						"id":    true,
						"name":  true,
						"email": true,
					},
				},
			},
			expectErr: false,
			checkSQL: func(t *testing.T, sql string, args []any) {
				assert.Contains(t, sql, "SELECT")
				assert.NotContains(t, sql, "SELECT *")
				assert.Contains(t, sql, "`id`")
				assert.Contains(t, sql, "`name`")
				assert.Contains(t, sql, "`email`")
				assert.Contains(t, sql, "FROM `users`")
				assert.Contains(t, sql, "WHERE")
				assert.Equal(t, 1, len(args))
			},
		},
		{
			name: "Selection只选择单个字段",
			input: &modelruntime.FindUniqueInput{
				TableName: "users",
				Where: map[string]any{
					"id": "123",
				},
				Selection: &modelruntime.Selection{
					FieldNames: map[string]bool{
						"name": true,
					},
				},
			},
			expectErr: false,
			checkSQL: func(t *testing.T, sql string, args []any) {
				assert.Contains(t, sql, "SELECT")
				assert.NotContains(t, sql, "SELECT *")
				assert.Contains(t, sql, "`name`")
				assert.Contains(t, sql, "FROM `users`")
			},
		},
		{
			name: "多个where条件",
			input: &modelruntime.FindUniqueInput{
				TableName: "users",
				Where: map[string]any{
					"id":    "123",
					"email": "test@example.com",
				},
			},
			expectErr: false,
			checkSQL: func(t *testing.T, sql string, args []any) {
				assert.Contains(t, sql, "SELECT * FROM `users`")
				assert.Contains(t, sql, "WHERE")
				assert.Equal(t, 2, len(args))
			},
		},
		{
			name: "复杂where条件-比较操作符",
			input: &modelruntime.FindUniqueInput{
				TableName: "users",
				Where: map[string]any{
					"age": map[string]interface{}{
						"gt": 18,
					},
				},
			},
			expectErr: false,
			checkSQL: func(t *testing.T, sql string, args []any) {
				assert.Contains(t, sql, "SELECT * FROM `users`")
				assert.Contains(t, sql, "WHERE")
			},
		},
		{
			name: "空where条件",
			input: &modelruntime.FindUniqueInput{
				TableName: "users",
				Where:     map[string]any{},
			},
			expectErr: true,
			errMsg:    "findUnique where cant be empty",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			sql, args, err := convertFindUniqueInputToSQL(ctx, tt.input)
			if tt.expectErr {
				require.Error(t, err)
				if tt.errMsg != "" {
					assert.Contains(t, err.Error(), tt.errMsg)
				}
			} else {
				require.NoError(t, err)
				assert.NotEmpty(t, sql)
				if tt.checkSQL != nil {
					tt.checkSQL(t, sql, args)
				}
			}
		})
	}
}

// TestConvertFindFirstInputToSQL 测试FindFirst转SQL
func TestConvertFindFirstInputToSQL(t *testing.T) {
	ctx := context.Background()

	tests := []struct {
		name      string
		input     *modelruntime.FindFirstInput
		expectErr bool
		checkSQL  func(t *testing.T, sql string, args []any)
	}{
		{
			name: "有where条件",
			input: &modelruntime.FindFirstInput{
				TableName: "users",
				Where: map[string]any{
					"status": "active",
				},
			},
			expectErr: false,
			checkSQL: func(t *testing.T, sql string, args []any) {
				assert.Contains(t, sql, "SELECT * FROM `users`")
				assert.Contains(t, sql, "WHERE")
				assert.Contains(t, sql, "LIMIT")
				// contains status 条件 和 limit
				assert.Equal(t, 2, len(args))
			},
		},
		{
			name: "有where条件和Selection",
			input: &modelruntime.FindFirstInput{
				TableName: "users",
				Where: map[string]any{
					"status": "active",
				},
				Selection: &modelruntime.Selection{
					FieldNames: map[string]bool{
						"id":     true,
						"name":   true,
						"status": true,
					},
				},
			},
			expectErr: false,
			checkSQL: func(t *testing.T, sql string, args []any) {
				assert.Contains(t, sql, "SELECT")
				assert.NotContains(t, sql, "SELECT *")
				assert.Contains(t, sql, "`id`")
				assert.Contains(t, sql, "`name`")
				assert.Contains(t, sql, "`status`")
				assert.Contains(t, sql, "WHERE")
				assert.Contains(t, sql, "LIMIT")
				assert.Equal(t, 2, len(args))
			},
		},
		{
			name: "无where条件",
			input: &modelruntime.FindFirstInput{
				TableName: "users",
				Where:     map[string]any{},
			},
			expectErr: false,
			checkSQL: func(t *testing.T, sql string, args []any) {
				assert.Contains(t, sql, "SELECT * FROM `users`")
				assert.Contains(t, sql, "LIMIT")
				assert.NotContains(t, sql, "WHERE")
			},
		},
		{
			name: "复杂where条件",
			input: &modelruntime.FindFirstInput{
				TableName: "users",
				Where: map[string]any{
					"OR": []any{
						map[string]any{"status": "active"},
						map[string]any{"status": "pending"},
					},
				},
			},
			expectErr: false,
			checkSQL: func(t *testing.T, sql string, args []any) {
				assert.Contains(t, sql, "SELECT * FROM `users`")
				assert.Contains(t, sql, "WHERE")
				assert.Contains(t, sql, "LIMIT")
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			sql, args, err := convertFindFirstInputToSQL(ctx, tt.input)
			if tt.expectErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
				assert.NotEmpty(t, sql)
				if tt.checkSQL != nil {
					tt.checkSQL(t, sql, args)
				}
			}
		})
	}
}

// TestConvertFindManyInputToSQL 测试FindMany转SQL
func TestConvertFindManyInputToSQL(t *testing.T) {
	ctx := context.Background()

	tests := []struct {
		name      string
		input     *modelruntime.FindManyInput
		expectErr bool
		checkSQL  func(t *testing.T, sql string, args []any)
	}{
		{
			name: "有where条件",
			input: &modelruntime.FindManyInput{
				TableName: "users",
				Where: map[string]any{
					"status": "active",
				},
			},
			expectErr: false,
			checkSQL: func(t *testing.T, sql string, args []any) {
				fmt.Printf("sql=%s", sql)
				assert.Contains(t, sql, "SELECT * FROM `users`")
				assert.Contains(t, sql, "WHERE")
				assert.Equal(t, 1, len(args))
			},
		},
		{
			name: "有where条件和Selection",
			input: &modelruntime.FindManyInput{
				TableName: "users",
				Where: map[string]any{
					"status": "active",
				},
				Selection: &modelruntime.Selection{
					FieldNames: map[string]bool{
						"id":         true,
						"name":       true,
						"email":      true,
						"created_at": true,
					},
				},
			},
			expectErr: false,
			checkSQL: func(t *testing.T, sql string, args []any) {
				assert.Contains(t, sql, "SELECT")
				assert.NotContains(t, sql, "SELECT *")
				assert.Contains(t, sql, "`id`")
				assert.Contains(t, sql, "`name`")
				assert.Contains(t, sql, "`email`")
				assert.Contains(t, sql, "`created_at`")
				assert.Contains(t, sql, "WHERE")
				assert.Equal(t, 1, len(args))
			},
		},
		{
			name: "有where条件 limit, offset",
			input: &modelruntime.FindManyInput{
				TableName: "users",
				Where: map[string]any{
					"status": "active",
				},
				Limit:  10,
				Offset: 1,
			},
			expectErr: false,
			checkSQL: func(t *testing.T, sql string, args []any) {
				fmt.Printf("sql=%s", sql)
				assert.Contains(t, sql, "SELECT * FROM `users`")
				assert.Contains(t, sql, "WHERE")
				assert.Contains(t, sql, "LIMIT")
				assert.Contains(t, sql, "OFFSET")
				assert.Equal(t, 3, len(args))
			},
		},
		{
			name: "无where条件",
			input: &modelruntime.FindManyInput{
				TableName: "users",
				Where:     map[string]any{},
			},
			expectErr: false,
			checkSQL: func(t *testing.T, sql string, args []any) {
				assert.Contains(t, sql, "SELECT * FROM `users`")
				assert.NotContains(t, sql, "WHERE")
			},
		},
		{
			name: "复杂where条件-IN操作符",
			input: &modelruntime.FindManyInput{
				TableName: "users",
				Where: map[string]any{
					"status": map[string]interface{}{
						"in": []string{"active", "pending"},
					},
				},
			},
			expectErr: false,
			checkSQL: func(t *testing.T, sql string, args []any) {
				assert.Contains(t, sql, "SELECT * FROM `users`")
				assert.Contains(t, sql, "WHERE")
			},
		},
		{
			name: "无where条件但有Selection",
			input: &modelruntime.FindManyInput{
				TableName: "users",
				Where:     map[string]any{},
				Selection: &modelruntime.Selection{
					FieldNames: map[string]bool{
						"id":   true,
						"name": true,
					},
				},
			},
			expectErr: false,
			checkSQL: func(t *testing.T, sql string, args []any) {
				assert.Contains(t, sql, "SELECT")
				assert.NotContains(t, sql, "SELECT *")
				assert.Contains(t, sql, "`id`")
				assert.Contains(t, sql, "`name`")
				assert.NotContains(t, sql, "WHERE")
			},
		},
		{
			name: "Selection和limit/offset组合",
			input: &modelruntime.FindManyInput{
				TableName: "users",
				Where: map[string]any{
					"status": "active",
				},
				Selection: &modelruntime.Selection{
					FieldNames: map[string]bool{
						"id":   true,
						"name": true,
					},
				},
				Limit:  10,
				Offset: 5,
			},
			expectErr: false,
			checkSQL: func(t *testing.T, sql string, args []any) {
				assert.Contains(t, sql, "SELECT")
				assert.NotContains(t, sql, "SELECT *")
				assert.Contains(t, sql, "`id`")
				assert.Contains(t, sql, "`name`")
				assert.Contains(t, sql, "WHERE")
				assert.Contains(t, sql, "LIMIT")
				assert.Contains(t, sql, "OFFSET")
				assert.Equal(t, 3, len(args)) // status + limit + offset
			},
		},
		{
			name: "带 orderBy 条件",
			input: &modelruntime.FindManyInput{
				TableName: "users",
				Where: map[string]any{
					"status": "active",
				},
				Limit:  10,
				Offset: 5,
				OrderBy: []modelruntime.OrderBy{
					{Field: "created_at", Direction: modelruntime.OrderByDesc},
					{Field: "id", Direction: modelruntime.OrderByAsc},
				},
			},
			expectErr: false,
			checkSQL: func(t *testing.T, sql string, args []any) {
				assert.Contains(t, sql, "ORDER BY")
				assert.Contains(t, sql, "`created_at` DESC")
				assert.Contains(t, sql, "`id` ASC")
				assert.Contains(t, sql, "LIMIT")
				assert.Contains(t, sql, "OFFSET")
				assert.Equal(t, 3, len(args))
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			sql, args, err := convertFindManyInputToSQL(ctx, tt.input)
			if tt.expectErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
				assert.NotEmpty(t, sql)
				if tt.checkSQL != nil {
					tt.checkSQL(t, sql, args)
				}
			}
		})
	}
}

// TestConvertCreateOneInputToSQL 测试CreateOne转SQL
func TestConvertCreateOneInputToSQL(t *testing.T) {
	ctx := context.Background()

	tests := []struct {
		name      string
		input     *modelruntime.CreateOneInput
		expectErr bool
		errMsg    string
		checkSQL  func(t *testing.T, sql string, args []any)
	}{
		{
			name: "正常创建",
			input: &modelruntime.CreateOneInput{
				TableName: "users",
				Data: map[string]any{
					"id":    "123",
					"name":  "test",
					"email": "test@example.com",
				},
			},
			expectErr: false,
			checkSQL: func(t *testing.T, sql string, args []any) {
				assert.Contains(t, sql, "INSERT INTO `users`")
				assert.Equal(t, 3, len(args))
			},
		},
		{
			name: "空数据",
			input: &modelruntime.CreateOneInput{
				TableName: "users",
				Data:      map[string]any{},
			},
			expectErr: true,
			errMsg:    "createOne data cannot be empty",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			sql, args, err := convertCreateOneInputToSQL(ctx, tt.input)
			if tt.expectErr {
				require.Error(t, err)
				if tt.errMsg != "" {
					assert.Contains(t, err.Error(), tt.errMsg)
				}
			} else {
				require.NoError(t, err)
				assert.NotEmpty(t, sql)
				if tt.checkSQL != nil {
					tt.checkSQL(t, sql, args)
				}
			}
		})
	}
}

// TestConvertUpdateOneInputToSQL 测试UpdateOne转SQL
func TestConvertUpdateOneInputToSQL(t *testing.T) {
	ctx := context.Background()

	tests := []struct {
		name      string
		input     *modelruntime.UpdateOneInput
		expectErr bool
		errMsg    string
		checkSQL  func(t *testing.T, sql string, args []any)
	}{
		{
			name: "正常更新",
			input: &modelruntime.UpdateOneInput{
				TableName: "users",
				Where: map[string]any{
					"id": "123",
				},
				Data: map[string]any{
					"name": "updated",
				},
			},
			expectErr: false,
			checkSQL: func(t *testing.T, sql string, args []any) {
				assert.Contains(t, sql, "UPDATE `users`")
				assert.Contains(t, sql, "SET")
				assert.Contains(t, sql, "WHERE")
				assert.Equal(t, 2, len(args))
			},
		},
		{
			name: "空where条件",
			input: &modelruntime.UpdateOneInput{
				TableName: "users",
				Where:     map[string]any{},
				Data: map[string]any{
					"name": "updated",
				},
			},
			expectErr: true,
			errMsg:    "updateOne where cannot be empty",
		},
		{
			name: "空data",
			input: &modelruntime.UpdateOneInput{
				TableName: "users",
				Where: map[string]any{
					"id": "123",
				},
				Data: map[string]any{},
			},
			expectErr: true,
			errMsg:    "updateOne data cannot be empty",
		},
		{
			name: "复杂where条件",
			input: &modelruntime.UpdateOneInput{
				TableName: "users",
				Where: map[string]any{
					"age": map[string]interface{}{
						"gte": 18,
					},
				},
				Data: map[string]any{
					"status": "adult",
				},
			},
			expectErr: false,
			checkSQL: func(t *testing.T, sql string, args []any) {
				assert.Contains(t, sql, "UPDATE `users`")
				assert.Contains(t, sql, "SET")
				assert.Contains(t, sql, "WHERE")
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			sql, args, err := convertUpdateOneInputToSQL(ctx, tt.input)
			if tt.expectErr {
				require.Error(t, err)
				if tt.errMsg != "" {
					assert.Contains(t, err.Error(), tt.errMsg)
				}
			} else {
				require.NoError(t, err)
				assert.NotEmpty(t, sql)
				if tt.checkSQL != nil {
					tt.checkSQL(t, sql, args)
				}
			}
		})
	}
}

// TestConvertUpdateManyInputToSQL 测试UpdateMany转SQL
func TestConvertUpdateManyInputToSQL(t *testing.T) {
	ctx := context.Background()

	tests := []struct {
		name      string
		input     *modelruntime.UpdateManyInput
		expectErr bool
		checkSQL  func(t *testing.T, sql string, args []any)
	}{
		{
			name: "有where条件",
			input: &modelruntime.UpdateManyInput{
				TableName: "users",
				Where: map[string]any{
					"status": "pending",
				},
				Data: map[string]any{
					"status": "active",
				},
				Take: 10,
			},
			expectErr: false,
			checkSQL: func(t *testing.T, sql string, args []any) {
				assert.Contains(t, sql, "UPDATE `users`")
				assert.Contains(t, sql, "SET")
				assert.Contains(t, sql, "WHERE")
				assert.Equal(t, 3, len(args))
			},
		},
		{
			name: "无where条件",
			input: &modelruntime.UpdateManyInput{
				TableName: "users",
				Where:     map[string]any{},
				Data: map[string]any{
					"updated_at": "2024-01-01",
				},
				Take: 10,
			},
			expectErr: false,
			checkSQL: func(t *testing.T, sql string, args []any) {
				assert.Contains(t, sql, "UPDATE `users`")
				assert.Contains(t, sql, "SET")
				assert.NotContains(t, sql, "WHERE")
			},
		},
		{
			name: "take 0",
			input: &modelruntime.UpdateManyInput{
				TableName: "users",
				Where:     map[string]any{},
				Data: map[string]any{
					"updated_at": "2024-01-01",
				},
				Take: 0,
			},
			expectErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			sql, args, err := convertUpdateManyInputToSQL(ctx, tt.input)
			if tt.expectErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
				assert.NotEmpty(t, sql)
				if tt.checkSQL != nil {
					tt.checkSQL(t, sql, args)
				}
			}
		})
	}
}

// TestConvertDeleteOneInputToSQL 测试DeleteOne转SQL
func TestConvertDeleteOneInputToSQL(t *testing.T) {
	ctx := context.Background()

	tests := []struct {
		name      string
		input     *modelruntime.DeleteOneInput
		expectErr bool
		errMsg    string
		checkSQL  func(t *testing.T, sql string, args []any)
	}{
		{
			name: "正常删除",
			input: &modelruntime.DeleteOneInput{
				TableName: "users",
				Where: map[string]any{
					"id": "123",
				},
			},
			expectErr: false,
			checkSQL: func(t *testing.T, sql string, args []any) {
				assert.Contains(t, sql, "DELETE FROM `users`")
				assert.Contains(t, sql, "WHERE")
				assert.Equal(t, 2, len(args))
			},
		},
		{
			name: "空where条件",
			input: &modelruntime.DeleteOneInput{
				TableName: "users",
				Where:     map[string]any{},
			},
			expectErr: true,
			errMsg:    "deleteOne where cannot be empty",
		},
		{
			name: "复杂where条件",
			input: &modelruntime.DeleteOneInput{
				TableName: "users",
				Where: map[string]any{
					"AND": []any{
						map[string]any{"status": "inactive"},
						map[string]any{
							"created_at": map[string]interface{}{
								"lt": "2020-01-01",
							},
						},
					},
				},
			},
			expectErr: false,
			checkSQL: func(t *testing.T, sql string, args []any) {
				assert.Contains(t, sql, "DELETE FROM `users`")
				assert.Contains(t, sql, "WHERE")
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			sql, args, err := convertDeleteOneInputToSQL(ctx, tt.input)
			if tt.expectErr {
				require.Error(t, err)
				if tt.errMsg != "" {
					assert.Contains(t, err.Error(), tt.errMsg)
				}
			} else {
				require.NoError(t, err)
				assert.NotEmpty(t, sql)
				if tt.checkSQL != nil {
					tt.checkSQL(t, sql, args)
				}
			}
		})
	}
}

// TestConvertCreateManyInputToSQL 测试CreateMany转SQL
func TestConvertCreateManyInputToSQL(t *testing.T) {
	ctx := context.Background()

	tests := []struct {
		name      string
		input     *modelruntime.CreateManyInput
		expectErr bool
		errMsg    string
		checkSQL  func(t *testing.T, sql string, args []any)
	}{
		{
			name: "正常批量创建",
			input: &modelruntime.CreateManyInput{
				TableName: "users",
				Data: []map[string]any{
					{"id": "1", "name": "user1"},
					{"id": "2", "name": "user2"},
					{"id": "3", "name": "user3"},
				},
			},
			expectErr: false,
			checkSQL: func(t *testing.T, sql string, args []any) {
				assert.Contains(t, sql, "INSERT INTO `users`")
				// 3条记录，每条2个字段，共6个参数
				assert.Equal(t, 6, len(args))
			},
		},
		{
			name: "单条记录",
			input: &modelruntime.CreateManyInput{
				TableName: "users",
				Data: []map[string]any{
					{"id": "1", "name": "user1"},
				},
			},
			expectErr: false,
			checkSQL: func(t *testing.T, sql string, args []any) {
				assert.Contains(t, sql, "INSERT INTO `users`")
				assert.Equal(t, 2, len(args))
			},
		},
		{
			name: "空数据",
			input: &modelruntime.CreateManyInput{
				TableName: "users",
				Data:      []map[string]any{},
			},
			expectErr: true,
			errMsg:    "createMany data cannot be empty",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			sql, args, err := convertCreateManyInputToSQL(ctx, tt.input)
			if tt.expectErr {
				require.Error(t, err)
				if tt.errMsg != "" {
					assert.Contains(t, err.Error(), tt.errMsg)
				}
			} else {
				require.NoError(t, err)
				assert.NotEmpty(t, sql)
				if tt.checkSQL != nil {
					tt.checkSQL(t, sql, args)
				}
			}
		})
	}
}

// TestConvertDeleteManyInputToSQL 测试DeleteMany转SQL
func TestConvertDeleteManyInputToSQL(t *testing.T) {
	ctx := context.Background()

	tests := []struct {
		name      string
		input     *modelruntime.DeleteManyInput
		expectErr bool
		checkSQL  func(t *testing.T, sql string, args []any)
	}{
		{
			name: "有where条件",
			input: &modelruntime.DeleteManyInput{
				TableName: "users",
				Where: map[string]any{
					"status": "inactive",
				},
				Take: 10,
			},
			expectErr: false,
			checkSQL: func(t *testing.T, sql string, args []any) {
				assert.Contains(t, sql, "DELETE FROM `users`")
				assert.Contains(t, sql, "WHERE")
				assert.Contains(t, sql, "LIMIT")
				// 1个where参数 + 1个limit参数
				assert.Equal(t, 2, len(args))
				assert.Equal(t, "inactive", args[0])
				assert.Equal(t, int64(10), args[1])
			},
		},
		{
			name: "无where条件",
			input: &modelruntime.DeleteManyInput{
				TableName: "users",
				Where:     map[string]any{},
				Take:      5,
			},
			expectErr: false,
			checkSQL: func(t *testing.T, sql string, args []any) {
				assert.Contains(t, sql, "DELETE FROM `users`")
				assert.NotContains(t, sql, "WHERE")
				assert.Contains(t, sql, "LIMIT")
				// 只有limit参数
				assert.Equal(t, 1, len(args))
				assert.Equal(t, int64(5), args[0])
			},
		},
		{
			name: "复杂where条件-比较操作符",
			input: &modelruntime.DeleteManyInput{
				TableName: "users",
				Where: map[string]any{
					"created_at": map[string]interface{}{
						"lt": "2020-01-01",
					},
				},
				Take: 100,
			},
			expectErr: false,
			checkSQL: func(t *testing.T, sql string, args []any) {
				assert.Contains(t, sql, "DELETE FROM `users`")
				assert.Contains(t, sql, "WHERE")
				assert.Contains(t, sql, "LIMIT")
			},
		},
		{
			name: "复杂where条件-AND操作符",
			input: &modelruntime.DeleteManyInput{
				TableName: "users",
				Where: map[string]any{
					"AND": []any{
						map[string]any{"status": "inactive"},
						map[string]any{
							"last_login": map[string]interface{}{
								"lt": "2023-01-01",
							},
						},
					},
				},
				Take: 50,
			},
			expectErr: false,
			checkSQL: func(t *testing.T, sql string, args []any) {
				assert.Contains(t, sql, "DELETE FROM `users`")
				assert.Contains(t, sql, "WHERE")
				assert.Contains(t, sql, "LIMIT")
			},
		},
		{
			name: "复杂where条件-OR操作符",
			input: &modelruntime.DeleteManyInput{
				TableName: "users",
				Where: map[string]any{
					"OR": []any{
						map[string]any{"status": "deleted"},
						map[string]any{"status": "banned"},
					},
				},
				Take: 20,
			},
			expectErr: false,
			checkSQL: func(t *testing.T, sql string, args []any) {
				assert.Contains(t, sql, "DELETE FROM `users`")
				assert.Contains(t, sql, "WHERE")
				assert.Contains(t, sql, "LIMIT")
			},
		},
		{
			name: "复杂where条件-IN操作符",
			input: &modelruntime.DeleteManyInput{
				TableName: "users",
				Where: map[string]any{
					"status": map[string]interface{}{
						"in": []string{"deleted", "banned", "suspended"},
					},
				},
				Take: 30,
			},
			expectErr: false,
			checkSQL: func(t *testing.T, sql string, args []any) {
				assert.Contains(t, sql, "DELETE FROM `users`")
				assert.Contains(t, sql, "WHERE")
				assert.Contains(t, sql, "LIMIT")
			},
		},
		{
			name: "Take为0",
			input: &modelruntime.DeleteManyInput{
				TableName: "users",
				Where: map[string]any{
					"status": "inactive",
				},
				Take: 0,
			},
			expectErr: true,
		},
		{
			name: "多个简单where条件",
			input: &modelruntime.DeleteManyInput{
				TableName: "users",
				Where: map[string]any{
					"status": "inactive",
					"role":   "guest",
				},
				Take: 15,
			},
			expectErr: false,
			checkSQL: func(t *testing.T, sql string, args []any) {
				assert.Contains(t, sql, "DELETE FROM `users`")
				assert.Contains(t, sql, "WHERE")
				assert.Contains(t, sql, "LIMIT")
				// 2个where参数 + 1个limit参数
				assert.Equal(t, 3, len(args))
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			sql, args, err := convertDeleteManyInputToSQL(ctx, tt.input)
			if tt.expectErr {
				require.Error(t, err)
			} else {
				require.NoError(t, err)
				assert.NotEmpty(t, sql)
				if tt.checkSQL != nil {
					tt.checkSQL(t, sql, args)
				}
			}
		})
	}
}

// TestSQLInjectionPrevention 测试SQL注入防护
func TestSQLInjectionPrevention(t *testing.T) {
	ctx := context.Background()

	// 测试恶意输入
	maliciousInput := &modelruntime.FindUniqueInput{
		TableName: "users",
		Where: map[string]any{
			"id": "123' OR '1'='1",
		},
	}

	sql, args, err := convertFindUniqueInputToSQL(ctx, maliciousInput)
	require.NoError(t, err)

	// 验证使用了参数化查询
	assert.Contains(t, sql, "?")
	assert.Equal(t, 1, len(args))
	// 恶意字符串应该作为参数值，而不是直接拼接到SQL中
	assert.Equal(t, "123' OR '1'='1", args[0])
}

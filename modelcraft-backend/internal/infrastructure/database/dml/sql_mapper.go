package dml

import (
	"context"
	"modelcraft/internal/domain/modelruntime"
	"modelcraft/internal/domain/query"
	"modelcraft/pkg/bizerrors"

	"github.com/doug-martin/goqu/v9"
)

// convertWhereToExpression 将where条件转换为goqu表达式
// 支持简单条件和复杂查询语法树
func convertWhereToExpression(where map[string]any) (goqu.Expression, error) {
	if len(where) == 0 {
		return nil, bizerrors.Errorf("where condition cannot be empty")
	}

	// 检查是否包含复杂操作符
	if containsComplexOperators(where) {
		// 使用新的查询解析器处理复杂条件
		return ParseAndConvert(where)
	}

	// 使用原有的简单处理方式，保持向后兼容
	return goqu.Ex(where), nil
}

// containsComplexOperators 检查是否包含复杂操作符
func containsComplexOperators(where map[string]any) bool {
	for key, value := range where {
		// 检查逻辑操作符
		if query.IsLogicalOperator(key) {
			return true
		}

		// 检查嵌套的比较操作符
		if valueMap, ok := value.(map[string]interface{}); ok {
			for operator := range valueMap {
				if query.IsComparisonOperator(operator) {
					return true
				}
			}
		}
	}
	return false
}

func convertFindUniqueInputToSQL(
	ctx context.Context,
	input *modelruntime.FindUniqueInput,
) (sql string, args []any, err error) {
	if len(input.Where) == 0 {
		err = bizerrors.Errorf("findUnique where cant be empty")
		return
	}

	// 使用新的表达式转换函数
	whereExpr, err := convertWhereToExpression(input.Where)
	if err != nil {
		return "", nil, bizerrors.Errorf("failed to convert where condition: %w", err)
	}
	dialectWrapper := goqu.Dialect("mysql")
	selectStep := dialectWrapper.Select("*")
	if input.Selection != nil {
		fieldNames := make([]any, 0, len(input.Selection.FieldNames))
		for fieldName := range input.Selection.FieldNames {
			fieldNames = append(fieldNames, fieldName)
		}
		selectStep = dialectWrapper.Select(fieldNames...)
	}

	ds := selectStep.From(input.TableName).Where(whereExpr)
	sql, args, err = ds.Prepared(true).ToSQL()
	return
}

func convertFindFirstInputToSQL(
	ctx context.Context,
	input *modelruntime.FindFirstInput,
) (sql string, args []any, err error) {
	dialectWrapper := goqu.Dialect("mysql")
	selectStep := dialectWrapper.Select("*")
	if input.Selection != nil {
		fieldNames := make([]any, 0, len(input.Selection.FieldNames))
		for fieldName := range input.Selection.FieldNames {
			fieldNames = append(fieldNames, fieldName)
		}
		selectStep = dialectWrapper.Select(fieldNames...)
	}
	if len(input.Where) == 0 {
		// 如果没有where条件，返回不带where的查询
		ds := selectStep.From(input.TableName).Limit(1)
		sql, args, err = ds.Prepared(true).ToSQL()
		return
	}
	// 使用新的表达式转换函数
	whereExpr, err := convertWhereToExpression(input.Where)
	if err != nil {
		return "", nil, bizerrors.Errorf("failed to convert where condition: %w", err)
	}

	ds := selectStep.From(input.TableName).Where(whereExpr).Limit(1)
	sql, args, err = ds.Prepared(true).ToSQL()
	return
}

func convertFindManyInputToSQL(
	ctx context.Context,
	input *modelruntime.FindManyInput,
) (sql string, args []any, err error) {
	dialectWrapper := goqu.Dialect("mysql")
	selectStep := dialectWrapper.Select("*")
	if input.Selection != nil {
		fieldNames := make([]any, 0, len(input.Selection.FieldNames))
		for fieldName := range input.Selection.FieldNames {
			fieldNames = append(fieldNames, fieldName)
		}
		selectStep = dialectWrapper.Select(fieldNames...)
	}
	if len(input.Where) == 0 {
		// 如果没有where条件，返回不带where的查询
		ds := selectStep.From(input.TableName).Limit(input.Limit).Offset(input.Offset)
		sql, args, err = ds.Prepared(true).ToSQL()
		return
	}

	// 使用新的表达式转换函数
	whereExpr, err := convertWhereToExpression(input.Where)
	if err != nil {
		return "", nil, bizerrors.Errorf("failed to convert where condition: %w", err)
	}

	ds := selectStep.From(input.TableName).Where(whereExpr).Limit(input.Limit).Offset(input.Offset)
	sql, args, err = ds.Prepared(true).ToSQL()
	return
}

func convertUpdateManyInputToSQL(
	ctx context.Context,
	input *modelruntime.UpdateManyInput,
) (sql string, args []any, err error) {
	if input.Take <= 0 {
		return "", nil, bizerrors.Errorf("take cant less equal than 0")
	}
	dialect := goqu.Dialect("mysql")
	goquSetRecord := input.Data
	if len(input.Where) == 0 {
		// 如果没有where条件，返回不带where的查询
		ds := dialect.Update(input.TableName).Set(goquSetRecord).Limit(input.Take)
		sql, args, err = ds.Prepared(true).ToSQL()
		return
	}

	// 使用新的表达式转换函数
	whereExpr, err := convertWhereToExpression(input.Where)
	if err != nil {
		return "", nil, bizerrors.Errorf("failed to convert where condition: %w", err)
	}

	ds := dialect.Update(input.TableName).Set(goquSetRecord).Where(whereExpr).Limit(input.Take)
	sql, args, err = ds.Prepared(true).ToSQL()
	return
}

func convertDeleteManyInputToSQL(
	ctx context.Context,
	input *modelruntime.DeleteManyInput,
) (sql string, args []any, err error) {
	if input.Take <= 0 {
		return "", nil, bizerrors.Errorf("take cant less equal than 0")
	}
	dialect := goqu.Dialect("mysql")
	if len(input.Where) == 0 {
		// 如果没有where条件，返回不带where的查询
		ds := dialect.Delete(input.TableName).Limit(input.Take)
		sql, args, err = ds.Prepared(true).ToSQL()
		return
	}

	// 使用新的表达式转换函数
	whereExpr, err := convertWhereToExpression(input.Where)
	if err != nil {
		return "", nil, bizerrors.Errorf("failed to convert where condition: %w", err)
	}

	ds := dialect.Delete(input.TableName).Where(whereExpr).Limit(input.Take)
	sql, args, err = ds.Prepared(true).ToSQL()
	return
}

// 变更操作的SQL转换函数

func convertCreateOneInputToSQL(
	ctx context.Context,
	input *modelruntime.CreateOneInput,
) (sql string, args []any, err error) {
	if len(input.Data) == 0 {
		err = bizerrors.Errorf("createOne data cannot be empty")
		return
	}

	ds := goqu.Dialect("mysql").Insert(input.TableName).Rows(input.Data)
	sql, args, err = ds.Prepared(true).ToSQL()
	return
}

func convertUpdateOneInputToSQL(
	ctx context.Context,
	input *modelruntime.UpdateOneInput,
) (sql string, args []any, err error) {
	if len(input.Where) == 0 {
		err = bizerrors.Errorf("updateOne where cannot be empty")
		return
	}
	if len(input.Data) == 0 {
		err = bizerrors.Errorf("updateOne data cannot be empty")
		return
	}

	// 使用新的表达式转换函数
	whereExpr, err := convertWhereToExpression(input.Where)
	if err != nil {
		return "", nil, bizerrors.Errorf("failed to convert where condition: %w", err)
	}

	ds := goqu.Dialect("mysql").Update(input.TableName).Set(input.Data).Where(whereExpr)
	sql, args, err = ds.Prepared(true).ToSQL()
	return
}

func convertDeleteOneInputToSQL(
	ctx context.Context,
	input *modelruntime.DeleteOneInput,
) (sql string, args []any, err error) {
	if len(input.Where) == 0 {
		err = bizerrors.Errorf("deleteOne where cannot be empty")
		return
	}

	// 使用新的表达式转换函数
	whereExpr, err := convertWhereToExpression(input.Where)
	if err != nil {
		return "", nil, bizerrors.Errorf("failed to convert where condition: %w", err)
	}

	ds := goqu.Dialect("mysql").Delete(input.TableName).Where(whereExpr).Limit(1)
	sql, args, err = ds.Prepared(true).ToSQL()
	return
}

func convertCreateManyInputToSQL(
	ctx context.Context,
	input *modelruntime.CreateManyInput,
) (sql string, args []any, err error) {
	if len(input.Data) == 0 {
		err = bizerrors.Errorf("createMany data cannot be empty")
		return
	}

	// 将 []map[string]any 转换为 []interface{}
	rows := make([]interface{}, 0, len(input.Data))
	for _, row := range input.Data {
		rows = append(rows, row)
	}

	ds := goqu.Dialect("mysql").Insert(input.TableName).Rows(rows...)
	sql, args, err = ds.Prepared(true).ToSQL()
	return
}

func convertAggregateInputToSQL(
	ctx context.Context,
	input *modelruntime.AggregateInput,
) (sql string, args []any, err error) {
	dialect := goqu.Dialect("mysql")

	// 构建 SELECT 子句，包含所有聚合函数
	selectionCount := len(input.Count) + len(input.Avg) + len(input.Sum) + len(input.Min) + len(input.Max)
	selections := make([]interface{}, 0, selectionCount)

	// 处理 _count
	for field := range input.Count {
		if field == modelruntime.Field_All {
			// COUNT(*) as _count__all
			selections = append(selections, goqu.COUNT("*").As("_count__all"))
		} else {
			// COUNT(field) as _count_field
			selections = append(selections, goqu.COUNT(field).As("_count_"+field))
		}
	}

	// 处理 _avg
	for field := range input.Avg {
		// AVG(field) as _avg_field
		selections = append(selections, goqu.AVG(field).As("_avg_"+field))
	}

	// 处理 _sum
	for field := range input.Sum {
		// SUM(field) as _sum_field
		selections = append(selections, goqu.SUM(field).As("_sum_"+field))
	}

	// 处理 _min
	for field := range input.Min {
		// MIN(field) as _min_field
		selections = append(selections, goqu.MIN(field).As("_min_"+field))
	}

	// 处理 _max
	for field := range input.Max {
		// MAX(field) as _max_field
		selections = append(selections, goqu.MAX(field).As("_max_"+field))
	}

	// 构建查询
	ds := dialect.Select(selections...).From(input.TableName)

	// 添加 WHERE 条件
	if len(input.Where) > 0 {
		whereExpr, err := convertWhereToExpression(input.Where)
		if err != nil {
			return "", nil, bizerrors.Errorf("failed to convert where condition: %w", err)
		}
		ds = ds.Where(whereExpr)
	}

	sql, args, err = ds.Prepared(true).ToSQL()
	return
}

// convertCountInputToSQL 将 CountInput 转换为 SQL 查询
// 参数:
//   - ctx: 上下文
//   - input: Count查询输入参数
//
// 返回:
//   - sql: 生成的SQL语句
//   - args: SQL参数
//   - err: 错误信息
//
// SQL生成规则:
//  1. 如果 Select 为空：SELECT COUNT(*) as count FROM table
//  2. 如果 Select 不为空：SELECT COUNT(*) as _count__all, COUNT(field1) as _count_field1, ... FROM table
//  3. 支持 WHERE 条件过滤
func convertCountInputToSQL(ctx context.Context, input *modelruntime.CountInput) (sql string, args []any, err error) {
	dialect := goqu.Dialect("mysql")

	// 构建 SELECT 子句
	var selections []interface{}

	if len(input.Select) == 0 {
		// 简单计数：SELECT COUNT(*) as count
		selections = append(selections, goqu.COUNT("*").As("count"))
	} else {
		// 字段级计数：SELECT COUNT(*) as _count__all, COUNT(field1) as _count_field1, ...
		for field := range input.Select {
			if field == modelruntime.Field_All {
				// COUNT(*) as _count__all（双下划线表示 _all 特殊字段）
				selections = append(selections, goqu.COUNT("*").As("_count__all"))
			} else {
				// COUNT(field) as _count_field
				selections = append(selections, goqu.COUNT(field).As("_count_"+field))
			}
		}
	}

	// 构建查询
	ds := dialect.Select(selections...).From(input.TableName)

	// 添加 WHERE 条件
	if len(input.Where) > 0 {
		whereExpr, err := convertWhereToExpression(input.Where)
		if err != nil {
			return "", nil, bizerrors.Errorf("failed to convert where condition: %w", err)
		}
		ds = ds.Where(whereExpr)
	}

	sql, args, err = ds.Prepared(true).ToSQL()
	return
}

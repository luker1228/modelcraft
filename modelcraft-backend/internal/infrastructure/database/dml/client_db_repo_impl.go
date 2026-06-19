package dml

import (
	"context"
	sqldb "database/sql"
	"fmt"
	"modelcraft/internal/domain/modelruntime"
	"modelcraft/pkg/bizerrors"
	"modelcraft/pkg/logfacade"
	"runtime/debug"
	"strconv"
	"strings"

	common "modelcraft/internal/infrastructure/database/common"

	"github.com/jmoiron/sqlx"
	"github.com/spf13/cast"
)

// ClientDBRepoImpl 客户端DB仓库实现
type ClientDBRepoImpl struct {
	stdDB *sqlx.DB
}

// NewClientDB 创建客户端数据库仓库实现
// 参数:
//   - clientDB: 标准数据库连接
//
// 返回:
//   - *ClientDBRepoImpl: 客户端数据库仓库实现实例
func NewClientDB(clientDB *sqldb.DB) *ClientDBRepoImpl {
	stdDB := sqlx.NewDb(clientDB, "mysql")
	return &ClientDBRepoImpl{
		stdDB: stdDB,
	}
}

func execute[T, R any](ctx context.Context, logger logfacade.Logger, input T, fn func() (R, error)) (R, error) {
	defer func() {
		if r := recover(); r != nil {
			logger.Infof(ctx, "execute_client_db_panic=%+v, stask=%s", r, debug.Stack())
			return
		}
	}()
	logger.Infof(ctx, "execute_client_db_input=%+v", input)
	result, err := fn()
	if err != nil {
		var r R
		logger.Infof(ctx, "execute_client_db_error=%+v", err)
		return r, common.WrapDatabaseError(err)
	}
	logger.Infof(ctx, "execute_client_db_result=%+v", result)
	return result, err
}

// FindUnique 查找唯一记录
// 参数:
//   - ctx: 上下文
//   - input: 查找唯一记录的输入参数
//
// 返回:
//   - map[string]any: 查找到的记录数据
//   - error: 错误信息
func (c *ClientDBRepoImpl) FindUnique(
	ctx context.Context,
	input *modelruntime.FindUniqueInput,
) (map[string]any, error) {
	logger := logfacade.GetLogger(ctx)
	return execute(ctx,
		logger, input,
		func() (map[string]any, error) {
			sql, args, err := convertFindUniqueInputToSQL(ctx, input)
			if err != nil {
				return nil, err
			}
			// 添加 LIMIT 2 以检查是否有多条记录
			sql = sql + " LIMIT 2"
			logger.Infof(ctx, "sql=%v args=%v", sql, args)

			rows, err := c.stdDB.Queryx(sql, args...)
			if err != nil {
				return nil, err
			}
			defer rows.Close()

			var results []map[string]any
			for rows.Next() {
				result := make(map[string]any)
				if err := rows.MapScan(result); err != nil {
					return nil, err
				}
				results = append(results, convertBytesToString(result))
			}

			if len(results) == 0 {
				return nil, sqldb.ErrNoRows
			}
			if len(results) > 1 {
				return nil, fmt.Errorf("multiple records found: expected 1, got %d", len(results))
			}
			return results[0], nil
		},
	)
}

// FindFirst 查找第一条记录
// 参数:
//   - ctx: 上下文
//   - input: 查找第一条记录的输入参数
//
// 返回:
//   - map[string]any: 查找到的记录数据
//   - error: 错误信息
func (c *ClientDBRepoImpl) FindFirst(ctx context.Context, input *modelruntime.FindFirstInput) (map[string]any, error) {
	logger := logfacade.GetLogger(ctx)
	return execute(ctx,
		logger, input,
		func() (map[string]any, error) {
			result := make(map[string]any)
			sql, args, err := convertFindFirstInputToSQL(ctx, input)
			if err != nil {
				return nil, err
			}
			logger.Infof(ctx, "sql=%v args=%v", sql, args)
			row := c.stdDB.QueryRowx(sql, args...)
			err = row.MapScan(result)
			if err != nil {
				return nil, err
			}
			result = convertBytesToString(result)
			return result, nil
		},
	)
}

// FindManyIn 通过 IN 条件批量查找关联记录，用于解决 N+1 问题。
// 等价于：SELECT * FROM tableName WHERE referenceKey IN (values...)
func (c *ClientDBRepoImpl) FindManyIn(
	ctx context.Context, input *modelruntime.FindManyInInput,
) ([]map[string]any, error) {
	logger := logfacade.GetLogger(ctx)
	return execute[*modelruntime.FindManyInInput, []map[string]any](
		ctx, logger, input,
		func() ([]map[string]any, error) {
			sql, args, err := convertFindManyInInputToSQL(ctx, input)
			if err != nil {
				return nil, err
			}
			logger.Infof(ctx, "sql=%v args=%v", sql, args)

			rows, err := c.stdDB.Queryx(sql, args...)
			if err != nil {
				logger.Error(ctx, "FindManyIn query fail", logfacade.Err(err))
				return nil, err
			}
			defer rows.Close()

			results := make([]map[string]any, 0, len(input.Values))
			for rows.Next() {
				record := make(map[string]any)
				if err := rows.MapScan(record); err != nil {
					logger.Error(ctx, "FindManyIn map scan fail", logfacade.Err(err))
					return nil, err
				}
				results = append(results, convertBytesToString(record))
			}
			if err := rows.Err(); err != nil {
				logger.Error(ctx, "FindManyIn rows iteration fail", logfacade.Err(err))
				return nil, err
			}

			logger.Infof(ctx, "FindManyIn found %d records", len(results))
			return results, nil
		},
	)
}

// 参数:
//   - ctx: 上下文
//   - input: 查找多条记录的输入参数
//
// 返回:
//   - []map[string]any: 查找到的记录列表
//   - error: 错误信息
func (c *ClientDBRepoImpl) FindMany(ctx context.Context, input *modelruntime.FindManyInput) ([]map[string]any, error) {
	// Short-circuit: take=0 explicitly requested — return empty result without hitting the DB.
	if input.ExplicitLimit && input.Limit == 0 {
		return []map[string]any{}, nil
	}
	logger := logfacade.GetLogger(ctx)
	return execute[*modelruntime.FindManyInput, []map[string]any](
		ctx, logger, input,
		func() ([]map[string]any, error) {
			sql, args, err := convertFindManyInputToSQL(ctx, input)
			if err != nil {
				return nil, err
			}
			logger.Infof(ctx, "sql=%v args=%v", sql, args)

			// 使用 Queryx 和 MapScan 来处理动态结果
			rows, err := c.stdDB.Queryx(sql, args...)
			if err != nil {
				logger.Error(ctx, "query fail", logfacade.Err(err))
				return nil, err
			}
			defer rows.Close()
			var results_cap uint
			if input.Limit < 10 {
				results_cap = 10
			} else {
				results_cap = input.Limit
			}
			results := make([]map[string]any, 0, results_cap)
			for rows.Next() {
				record := make(map[string]any)
				err := rows.MapScan(record)
				if err != nil {
					logger.Error(ctx, "map scan fail", logfacade.Err(err))
					return nil, err
				}
				results = append(results, convertBytesToString(record))
			}

			if err := rows.Err(); err != nil {
				logger.Error(ctx, "rows iteration fail", logfacade.Err(err))
				return nil, err
			}

			results = convertBytesSliceToString(results)
			logger.Infof(ctx, "found %d records", len(results))
			return results, nil
		},
	)
}

// ListByCursor executes a keyset cursor pagination query.
// Returns at most limit+1 rows — the caller checks len(result) > limit to determine hasNextPage.
func (c *ClientDBRepoImpl) ListByCursor(
	ctx context.Context,
	input *modelruntime.ListByCursorInput,
) ([]map[string]any, error) {
	logger := logfacade.GetLogger(ctx)
	return execute(ctx, logger, input, func() ([]map[string]any, error) {
		sql, args, err := convertListByCursorInputToSQL(ctx, input)
		if err != nil {
			return nil, err
		}
		logger.Infof(ctx, "sql=%v args=%v", sql, args)

		rows, err := c.stdDB.Queryx(sql, args...)
		if err != nil {
			logger.Error(ctx, "listByCursor query fail", logfacade.Err(err))
			return nil, err
		}
		defer rows.Close()

		results := make([]map[string]any, 0, input.Limit+1)
		for rows.Next() {
			record := make(map[string]any)
			if err := rows.MapScan(record); err != nil {
				logger.Error(ctx, "listByCursor map scan fail", logfacade.Err(err))
				return nil, err
			}
			results = append(results, convertBytesToString(record))
		}
		if err := rows.Err(); err != nil {
			logger.Error(ctx, "listByCursor rows iteration fail", logfacade.Err(err))
			return nil, err
		}
		return results, nil
	})
}

// convertBytesToString 将 map 中的 []byte 值转换为 string
func convertBytesToString(data map[string]any) map[string]any {
	for key, value := range data {
		if bytes, ok := value.([]byte); ok {
			data[key] = string(bytes)
		}
	}
	return data
}

// convertBytesSliceToString 将 []map[string]any 中所有的 []byte 值转换为 string
func convertBytesSliceToString(dataSlice []map[string]any) []map[string]any {
	for i := range dataSlice {
		dataSlice[i] = convertBytesToString(dataSlice[i])
	}
	return dataSlice
}

// isAggregateField 检查字段是否是聚合字段
func isAggregateField(key string) bool {
	return strings.HasPrefix(key, "_count_") ||
		strings.HasPrefix(key, "_count__") ||
		strings.HasPrefix(key, "_avg_") ||
		strings.HasPrefix(key, "_sum_") ||
		strings.HasPrefix(key, "_min_") ||
		strings.HasPrefix(key, "_max_")
}

// convertAggregateBytes 将聚合查询结果中的 []byte 值转换为适当的类型
// 参数:
//   - data: 聚合查询的原始结果 map
//
// 返回:
//   - map[string]any: 转换后的结果，聚合字段的 []byte 会被转换为 int64/float64/string
//
// 转换规则:
//  1. 先将所有 []byte 转换为 string（处理 MySQL 的字节数组输出）
//  2. 对于聚合字段（_count_*, _avg_*, _sum_*, _min_*, _max_*）:
//     - 尝试解析为 int64（用于 COUNT 和整数类型的 SUM/AVG/MIN/MAX）
//     - 如果失败，尝试解析为 float64（用于小数类型的 SUM/AVG/MIN/MAX）
//     - 如果都失败，保持为 string（用于日期时间类型的 MIN/MAX）
func convertAggregateBytes(data map[string]any) map[string]any {
	// 步骤 1: 将所有 []byte 转换为 string
	data = convertBytesToString(data)

	// 步骤 2: 将聚合字段的 string 值转换为数值类型
	for key, value := range data {
		if !isAggregateField(key) {
			continue
		}
		if parsed, ok := parseAggregateValue(value); ok {
			data[key] = parsed
		}
	}
	return data
}

func parseAggregateValue(value any) (any, bool) {
	str, ok := value.(string)
	if !ok {
		return nil, false
	}

	if intVal, err := strconv.ParseInt(str, 10, 64); err == nil {
		return intVal, true
	}
	if floatVal, err := strconv.ParseFloat(str, 64); err == nil {
		return floatVal, true
	}
	return str, true
}

// Aggregate 执行聚合查询
// 参数:
//   - ctx: 上下文
//   - input: 聚合查询的输入参数
//
// 返回:
//   - map[string]any: 聚合查询结果（嵌套结构：_count._all, _avg.field等）
//   - error: 错误信息
func (c *ClientDBRepoImpl) Aggregate(ctx context.Context, input *modelruntime.AggregateInput) (map[string]any, error) {
	logger := logfacade.GetLogger(ctx)
	return execute(ctx,
		logger, input,
		func() (map[string]any, error) {
			sql, args, err := convertAggregateInputToSQL(ctx, input)
			if err != nil {
				return nil, err
			}
			logger.Infof(ctx, "sql=%v args=%v", sql, args)

			// 执行查询并获取单行结果
			flatResult := make(map[string]any)
			row := c.stdDB.QueryRowx(sql, args...)
			err = row.MapScan(flatResult)
			if err != nil {
				return nil, err
			}

			// 将 []byte 转换为适当的数值类型
			flatResult = convertAggregateBytes(flatResult)

			// 将扁平结果转换为嵌套结构
			// 例如：{ "_count__all": 10, "_avg_amount": 50.5 }
			// 转换为：{ "_count": { "_all": 10 }, "_avg": { "amount": 50.5 } }
			nestedResult := convertFlatToNestedAggregate(flatResult)

			logger.Infof(ctx, "aggregate result=%+v", nestedResult)
			return nestedResult, nil
		},
	)
}

// convertFlatToNestedAggregate 将扁平的聚合结果转换为嵌套结构
func convertFlatToNestedAggregate(flat map[string]any) map[string]any {
	result := make(map[string]any)

	// 初始化嵌套结构
	countMap := make(map[string]any)
	avgMap := make(map[string]any)
	sumMap := make(map[string]any)
	minMap := make(map[string]any)
	maxMap := make(map[string]any)

	for key, value := range flat {
		// 解析列名并放入相应的嵌套map
		switch {
		case strings.HasPrefix(key, "_count_"):
			// _count__all -> _count._all
			// _count_field -> _count.field
			fieldName := strings.TrimPrefix(key, "_count_")
			countMap[fieldName] = value
		case strings.HasPrefix(key, "_avg_"):
			// _avg_field -> _avg.field
			fieldName := strings.TrimPrefix(key, "_avg_")
			avgMap[fieldName] = value
		case strings.HasPrefix(key, "_sum_"):
			// _sum_field -> _sum.field
			fieldName := strings.TrimPrefix(key, "_sum_")
			sumMap[fieldName] = value
		case strings.HasPrefix(key, "_min_"):
			// _min_field -> _min.field
			fieldName := strings.TrimPrefix(key, "_min_")
			minMap[fieldName] = value
		case strings.HasPrefix(key, "_max_"):
			// _max_field -> _max.field
			fieldName := strings.TrimPrefix(key, "_max_")
			maxMap[fieldName] = value
		}
	}

	// 只添加非空的聚合结果
	if len(countMap) > 0 {
		result[modelruntime.Field_Count] = countMap
	}
	if len(avgMap) > 0 {
		result[modelruntime.Field_Avg] = avgMap
	}
	if len(sumMap) > 0 {
		result[modelruntime.Field_Sum] = sumMap
	}
	if len(minMap) > 0 {
		result[modelruntime.Field_Min] = minMap
	}
	if len(maxMap) > 0 {
		result[modelruntime.Field_Max] = maxMap
	}

	return result
}

// Count 执行计数查询
// 参数:
//   - ctx: 上下文
//   - input: 计数查询的输入参数
//
// 返回:
//   - map[string]any: 计数查询结果，格式为 {count: N} 或 {fieldsCount: {_all: N, field1: N, ...}}
//   - error: 错误信息
func (c *ClientDBRepoImpl) Count(ctx context.Context, input *modelruntime.CountInput) (map[string]any, error) {
	logger := logfacade.GetLogger(ctx)
	return execute(ctx,
		logger, input,
		func() (map[string]any, error) {
			sql, args, err := convertCountInputToSQL(ctx, input)
			if err != nil {
				return nil, err
			}
			logger.Infof(ctx, "sql=%v args=%v", sql, args)

			// 执行查询并获取单行结果
			flatResult := make(map[string]any)
			row := c.stdDB.QueryRowx(sql, args...)
			err = row.MapScan(flatResult)
			if err != nil {
				return nil, bizerrors.Errorf("failed to execute count query: %w", err)
			}

			logger.Infof(ctx, "count flatResult=%+v", flatResult)

			// 转换字节数组和数值类型
			flatResult = convertAggregateBytes(flatResult)

			// 根据是否使用 select 参数决定返回格式
			if len(input.Select) == 0 {
				// 简单计数：{ count: N }
				return flatResult, nil
			}

			// 字段级计数：{ fieldsCount: { _all: N, field1: N, ... } }
			fieldsCountMap := make(map[string]any)
			for key, value := range flatResult {
				// 移除 _count_ 或 _count__ 前缀
				fieldName := strings.TrimPrefix(key, "_count_")
				fieldName = strings.TrimPrefix(fieldName, "_count__")
				fieldsCountMap[fieldName] = value
			}

			result := map[string]any{
				modelruntime.FieldFieldsCount: fieldsCountMap,
			}

			logger.Infof(ctx, "count result=%+v", result)
			return result, nil
		},
	)
}

// 变更操作实现

// CreateOne 创建单条记录
// 参数:
//   - ctx: 上下文
//   - input: 创建单条记录的输入参数
//
// 返回:
//   - string: 创建的记录ID
//   - error: 错误信息
func (c *ClientDBRepoImpl) CreateOne(ctx context.Context, input *modelruntime.CreateOneInput) (string, error) {
	logger := logfacade.GetLogger(ctx)
	return execute[*modelruntime.CreateOneInput, string](
		ctx, logger, input,
		func() (string, error) {
			input.Data["id"] = input.Id
			sql, args, err := convertCreateOneInputToSQL(ctx, input)
			if err != nil {
				return "", err
			}
			logger.Infof(ctx, "sql=%v args=%v", sql, args)

			// 执行插入操作
			_, err = c.stdDB.Exec(sql, args...)
			if err != nil {
				logger.Error(ctx, "createOne fail", logfacade.Err(err))
				return "", err
			}
			return input.Id, nil
		},
	)
}

// UpdateOne 更新单条记录
// 参数:
//   - ctx: 上下文
//   - input: 更新单条记录的输入参数
//
// 返回:
//   - map[string]any: 更新后的记录数据（如果需要返回）
//   - error: 错误信息
func (c *ClientDBRepoImpl) UpdateOne(ctx context.Context, input *modelruntime.UpdateOneInput) (map[string]any, error) {
	logger := logfacade.GetLogger(ctx)
	return execute[*modelruntime.UpdateOneInput, map[string]any](
		ctx, logger, input,
		func() (map[string]any, error) {
			sql, args, err := convertUpdateOneInputToSQL(ctx, input)
			if err != nil {
				return nil, err
			}
			logger.Infof(ctx, "sql=%v args=%v", sql, args)

			result, err := c.stdDB.Exec(sql, args...)
			if err != nil {
				logger.Error(ctx, "updateOne fail", logfacade.Err(err))
				return nil, err
			}

			rowsAffected, err := result.RowsAffected()
			if err != nil {
				logger.Error(ctx, "get rows affected fail", logfacade.Err(err))
				return nil, err
			}

			if rowsAffected == 0 {
				return nil, sqldb.ErrNoRows
			}

			if input.UpdatedObj {
				findInput := &modelruntime.FindUniqueInput{
					TableName:  input.TableName,
					Where:      input.Where,
					RawFilters: input.RawFilters,
				}
				return c.FindUnique(ctx, findInput)
			}
			return nil, nil
		},
	)
}

// DeleteOne 删除单条记录
// 参数:
//   - ctx: 上下文
//   - input: 删除单条记录的输入参数
//
// 返回:
//   - map[string]any: 被删除的记录数据（如果需要返回）
//   - error: 错误信息
func (c *ClientDBRepoImpl) DeleteOne(ctx context.Context, input *modelruntime.DeleteOneInput) (map[string]any, error) {
	logger := logfacade.GetLogger(ctx)
	return execute[*modelruntime.DeleteOneInput, map[string]any](
		ctx, logger, input,
		func() (map[string]any, error) {
			findInput := &modelruntime.FindUniqueInput{
				TableName:  input.TableName,
				Where:      input.Where,
				RawFilters: input.RawFilters,
			}
			toDeleteOne, err := c.FindUnique(ctx, findInput)
			if err != nil {
				return nil, err
			}

			var deletedRecord map[string]any
			if input.DeletedObj {
				deletedRecord = toDeleteOne
			}

			sql, args, err := convertDeleteOneInputToSQL(ctx, input)
			if err != nil {
				return nil, err
			}
			logger.Infof(ctx, "sql=%v args=%v", sql, args)

			result, err := c.stdDB.Exec(sql, args...)
			if err != nil {
				logger.Error(ctx, "deleteOne fail", logfacade.Err(err))
				return nil, err
			}

			rowsAffected, err := result.RowsAffected()
			if err != nil {
				logger.Error(ctx, "get rows affected fail", logfacade.Err(err))
				return nil, err
			}
			_ = rowsAffected

			return deletedRecord, nil
		},
	)
}

// CreateMany 批量创建记录
// 参数:
//   - ctx: 上下文
//   - input: 批量创建记录的输入参数
//
// 返回:
//   - interface{}: 创建结果（包含数量和ID列表）
//   - error: 错误信息
func (c *ClientDBRepoImpl) CreateMany(ctx context.Context, input *modelruntime.CreateManyInput) (interface{}, error) {
	logger := logfacade.GetLogger(ctx)
	logger.Infof(ctx, "CreateMany_start: tableName=%s, dataCount=%d, skipDuplicates=%v, returnIdList=%v",
		input.TableName, len(input.Data), input.SkipDuplicates, input.ReturnIdList)

	// 验证批量大小（防御性检查）
	if len(input.Data) > modelruntime.MaxCreateManyBatchSize {
		return nil, fmt.Errorf(
			"createMany batch size exceeds limit: %d > %d",
			len(input.Data),
			modelruntime.MaxCreateManyBatchSize,
		)
	}

	if len(input.Data) == 0 {
		return map[string]any{
			modelruntime.FieldCount:  0,
			modelruntime.FieldIdList: []string{},
		}, nil
	}

	if input.SkipDuplicates {
		// 策略A：逐条插入，跳过唯一索引冲突
		return c.createManyWithSkipDuplicates(ctx, input)
	}

	// 批量插入，不跳过冲突
	return c.createManyBatch(ctx, input)
}

// createManyBatch 批量插入（不跳过唯一索引冲突）
func (c *ClientDBRepoImpl) createManyBatch(
	ctx context.Context,
	input *modelruntime.CreateManyInput,
) (interface{}, error) {
	logger := logfacade.GetLogger(ctx)

	sql, args, err := convertCreateManyInputToSQL(ctx, input)
	if err != nil {
		logger.Error(ctx, "convert sql fail", logfacade.Err(err))
		return nil, common.WrapDatabaseError(err)
	}
	logger.Infof(ctx, "sql=%v args=%v", sql, args)

	// 执行批量插入操作
	result, err := c.stdDB.Exec(sql, args...)
	if err != nil {
		logger.Error(ctx, "batch insert fail", logfacade.Err(err))
		return nil, common.WrapDatabaseError(err)
	}

	// 获取受影响的行数
	rowsAffected, err := result.RowsAffected()
	if err != nil {
		logger.Error(ctx, "get rows affected fail", logfacade.Err(err))
		return nil, common.WrapDatabaseError(err)
	}

	// 构建返回结果
	response := map[string]any{
		modelruntime.FieldCount: int(rowsAffected),
	}

	// 如果需要返回ID列表，查询插入的记录
	if input.ReturnIdList {
		idList := make([]string, 0, len(input.Data))
		for _, item := range input.Data {
			idList = append(idList, cast.ToString(item["id"]))
		}
		response[modelruntime.FieldIdList] = idList
	}

	return response, nil
}

// createManyWithSkipDuplicates 逐条插入，跳过唯一索引冲突
func (c *ClientDBRepoImpl) createManyWithSkipDuplicates(
	ctx context.Context,
	input *modelruntime.CreateManyInput,
) (interface{}, error) {
	logger := logfacade.GetLogger(ctx)

	idList := make([]string, 0, len(input.Data))
	for i, row := range input.Data {
		// 构建单条插入的SQL
		singleInput := &modelruntime.CreateManyInput{
			TableName: input.TableName,
			Data:      []map[string]any{row},
		}

		sql, args, err := convertCreateManyInputToSQL(ctx, singleInput)
		if err != nil {
			logger.Warn(ctx, fmt.Sprintf("convert sql fail at index %d: %v", i, err))
			continue
		}

		logger.Infof(ctx, "insert row %d: sql=%v args=%v", i, sql, args)

		_, err = c.stdDB.Exec(sql, args...)
		if err != nil {
			// 检查是否是唯一索引冲突错误
			if isUniqueConstraintError(err) {
				logger.Warn(ctx, fmt.Sprintf("unique constraint violation at index %d, skipping", i))
				continue
			}
			// 其他错误直接返回
			logger.Error(ctx, "insert fail", logfacade.Err(err))
			return nil, common.WrapDatabaseError(err)
		}

		// 获取插入的ID
		idList = append(idList, cast.ToString(row[modelruntime.FieldID]))
	}

	logger.Infof(ctx, "createMany_success: idListLen=%d", len(idList))

	response := map[string]any{
		modelruntime.FieldCount: len(idList),
	}

	if input.ReturnIdList {
		response[modelruntime.FieldIdList] = idList
	}

	return response, nil
}

// isUniqueConstraintError 检查是否是唯一索引冲突错误
func isUniqueConstraintError(err error) bool {
	if err == nil {
		return false
	}

	errMsg := err.Error()
	// MySQL唯一索引冲突错误包含这些关键字
	return strings.Contains(errMsg, "Duplicate entry") ||
		strings.Contains(errMsg, "duplicate key") ||
		strings.Contains(errMsg, "UNIQUE constraint failed")
}

// UpdateMany 批量更新记录
// 参数:
//   - ctx: 上下文
//   - input: 批量更新记录的输入参数
//
// 返回:
//   - interface{}: 更新结果（包含受影响的记录数量）
//   - error: 错误信息
func (c *ClientDBRepoImpl) UpdateMany(ctx context.Context, input *modelruntime.UpdateManyInput) (interface{}, error) {
	logger := logfacade.GetLogger(ctx)
	logger.Infof(ctx, "UpdateMany_start: tableName=%s, take=%d, whereIsNil=%v",
		input.TableName, input.Take, input.Where == nil)

	if input.Take < 1 || input.Take > modelruntime.MaxCreateManyBatchSize {
		return nil, fmt.Errorf("take must be between 1 and %d, got %d", modelruntime.MaxCreateManyBatchSize, input.Take)
	}

	// 构建更新SQL
	sql, args, err := convertUpdateManyInputToSQL(ctx, input)
	if err != nil {
		return nil, err
	}
	logger.Infof(ctx, "sql=%s args=%v", sql, args)

	result, err := c.stdDB.Exec(sql, args...)
	if err != nil {
		logger.Error(ctx, "updateMany fail", logfacade.Err(err))
		return nil, common.WrapDatabaseError(err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		logger.Error(ctx, "get rows affected fail", logfacade.Err(err))
		return nil, common.WrapDatabaseError(err)
	}

	logger.Infof(ctx, "updateMany_success: count=%d", rowsAffected)

	return map[string]any{
		modelruntime.FieldCount: int(rowsAffected),
	}, nil
}

// DeleteMany 批量删除记录
// 参数:
//   - ctx: 上下文
//   - input: 批量删除记录的输入参数
//
// 返回:
//   - interface{}: 删除结果（包含受影响的记录数量）
//   - error: 错误信息
func (c *ClientDBRepoImpl) DeleteMany(ctx context.Context, input *modelruntime.DeleteManyInput) (interface{}, error) {
	logger := logfacade.GetLogger(ctx)
	logger.Infof(ctx, "DeleteMany_start: tableName=%s, take=%d, whereIsNil=%v",
		input.TableName, input.Take, input.Where == nil)

	if input.Take < 1 || input.Take > modelruntime.MaxCreateManyBatchSize {
		return nil, fmt.Errorf("take must be between 1 and %d, got %d", modelruntime.MaxCreateManyBatchSize, input.Take)
	}

	// 构建删除SQL
	sql, args, err := convertDeleteManyInputToSQL(ctx, input)
	if err != nil {
		return nil, err
	}

	logger.Infof(ctx, "sql=%s args=%v", sql, args)

	result, err := c.stdDB.Exec(sql, args...)
	if err != nil {
		logger.Error(ctx, "deleteMany fail", logfacade.Err(err))
		return nil, common.WrapDatabaseError(err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		logger.Error(ctx, "get rows affected fail", logfacade.Err(err))
		return nil, common.WrapDatabaseError(err)
	}

	logger.Infof(ctx, "deleteMany_success: count=%d", rowsAffected)

	return map[string]any{
		modelruntime.FieldCount: int(rowsAffected),
	}, nil
}

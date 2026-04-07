package modelruntime

import (
	"context"

	"github.com/graph-gophers/dataloader/v7"
)

// newRelationBatchLoader 为单个 (tableName, referenceKey) 组合创建一个
// dataloader.Loader[string, map[string]any]。
//
// 批量函数在 graphql-go 广度优先执行完当前层所有字段 resolver 的 Load() 调用后，
// 由 dataloader 聚合触发，发出一条 WHERE referenceKey IN (...) SQL。
//
// 键类型 K = string（外键值的字符串形式）
// 值类型 V = map[string]any（目标表记录，nil 表示对应记录不存在）
func newRelationBatchLoader(
	clientRepo ClientDatabaseRepository,
	tableName string,
	referenceKey string,
) *dataloader.Loader[string, map[string]any] {
	batchFn := func(ctx context.Context, keys []string) []*dataloader.Result[map[string]any] {
		results := make([]*dataloader.Result[map[string]any], len(keys))

		// 将 []string 转换为 []any 供 FindManyIn 使用
		values := make([]any, len(keys))
		for i, k := range keys {
			values[i] = k
		}

		records, err := clientRepo.FindManyIn(ctx, &FindManyInInput{
			TableName:    tableName,
			ReferenceKey: referenceKey,
			Values:       values,
		})
		if err != nil {
			// 批量查询失败：所有 key 均返回错误
			for i := range results {
				results[i] = &dataloader.Result[map[string]any]{Error: err}
			}
			return results
		}

		// 将结果按 referenceKey 建立索引
		index := make(map[string]map[string]any, len(records))
		for _, record := range records {
			if val, ok := record[referenceKey]; ok {
				if key, ok := toString(val); ok {
					index[key] = record
				}
			}
		}

		// 按照 keys 顺序填充结果，不存在的 key 返回 nil（悬空外键）
		for i, key := range keys {
			results[i] = &dataloader.Result[map[string]any]{
				Data: index[key], // nil if not found
			}
		}

		return results
	}

	return dataloader.NewBatchedLoader(batchFn)
}

// toString 将 any 值转为 string，用于 dataloader 的 key。
func toString(v any) (string, bool) {
	switch s := v.(type) {
	case string:
		return s, true
	case []byte:
		return string(s), true
	default:
		return "", false
	}
}

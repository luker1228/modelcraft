# Runtime 反向关联查询实现计划

> **For agentic workers:** REQUIRED: Use superpowers:subagent-driven-development (if subagents available) or superpowers:executing-plans to implement this plan. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** 支持一对多（reverse）关联查询，使 User 模型可以嵌套查询所有关联的 Order 记录

**Architecture:** 在 `model_resolver.go` 的 `createRelationField` 中添加方向判断，分支处理 normal（多对一）和 reverse（一对多）两种情况。reverse 方向使用 FindMany 查询关联表，构建的 WHERE 条件支持复合外键。

**Tech Stack:** Go, graphql-go, LogicalForeignKey (Direction enum), graphql.NewList

---

## 文件修改范围

**修改**：`internal/domain/modelruntime/model_resolver.go`
- Line 800-833：`createRelationField` 方法，添加方向判断
- Line 836-893：`createManyToOneFieldFromFK` / `createManyToOneResolverFromFK`（现有，无改动）
- 新增：`createOneToManyFieldFromFK` 方法（~10 行）
- 新增：`createOneToManyResolverFromFK` 方法（~50 行）

**新增测试**：`internal/domain/modelruntime/model_resolver_test.go`
- 新增测试 `TestCreateRelationFieldReverse`

---

## Chunk 1: 修改 createRelationField 添加方向判断

### Task 1: 修改 createRelationField 添加方向判断

**文件**：`internal/domain/modelruntime/model_resolver.go:800-834`

- [ ] **Step 1: 阅读现有 createRelationField 实现**

文件位置：`internal/domain/modelruntime/model_resolver.go` 行 800-834

现有代码结构：
```go
func (r *graphqlModelResolver) createRelationField(...) (*graphql.Field, error) {
	if field.RelateFKID == nil {
		return nil, bizerrors.Errorf("RELATION field %s has no relate_fk_id", field.Name)
	}

	// 通过 relate_fk_id 查询 LogicalForeignKey
	lf, err := r.lfkRepo.GetByID(r.ctx, *field.RelateFKID)
	if err != nil {
		return nil, bizerrors.Errorf("failed to get logical foreign key for field %s: %w", field.Name, err)
	}

	// 获取被引用模型
	refModel, err := r.modelRepo.GetByID(r.ctx, lf.RefModelID)
	if err != nil {
		return nil, bizerrors.Errorf("failed to get reference model %s: %w", lf.RefModelID, err)
	}

	// 获取或创建被引用模型的 GraphQL 对象类型
	referenceObj, exists := relateObjMaps[refModel.Name]
	if !exists {
		referenceObj, err = r.generateModelType(maxDepth-1, refModel, relateObjMaps)
		if err != nil {
			return nil, bizerrors.Errorf("failed to generate model type for %s: %w", refModel.Name, err)
		}
		relateObjMaps[refModel.Name] = referenceObj
	}

	// RELATION 字段表示多对一关系
	return r.createManyToOneFieldFromFK(lf, referenceObj, graphqlField), nil
}
```

**需要改动的部分**：最后一行替换为方向判断

- [ ] **Step 2: 实现方向判断分支**

将现有代码的最后一行：
```go
return r.createManyToOneFieldFromFK(lf, referenceObj, graphqlField), nil
```

替换为：
```go
if lf.IsReverse() {
	return r.createOneToManyFieldFromFK(lf, referenceObj, graphqlField), nil
}

return r.createManyToOneFieldFromFK(lf, referenceObj, graphqlField), nil
```

完整改动在 `internal/domain/modelruntime/model_resolver.go` 的 `createRelationField` 方法末尾。

- [ ] **Step 3: Commit**

```bash
cd /root/modelcraft_project/modelcraft-go
git add internal/domain/modelruntime/model_resolver.go
git commit -m "feat: add direction check in createRelationField for normal/reverse FK handling"
```

---

## Chunk 2: 新增 createOneToManyFieldFromFK 方法

### Task 2: 新增 createOneToManyFieldFromFK 方法

**文件**：`internal/domain/modelruntime/model_resolver.go`

在 `createManyToOneFieldFromFK` 方法后面添加新方法（行号大约在 835 后）。

- [ ] **Step 1: 添加 createOneToManyFieldFromFK 方法**

在 `createManyToOneFieldFromFK` 之后添加：

```go
// createOneToManyFieldFromFK 创建一对多关系字段（基于 LogicalForeignKey 的 reverse 方向）
func (r *graphqlModelResolver) createOneToManyFieldFromFK(lf *modeldesign.LogicalForeignKey,
	referenceObj *graphql.Object, graphqlField *graphql.Field,
) *graphql.Field {
	graphqlField.Type = graphql.NewList(graphql.NewNonNull(referenceObj))
	graphqlField.Resolve = r.createOneToManyResolverFromFK(lf, referenceObj.Name())
	return graphqlField
}
```

- [ ] **Step 2: Commit**

```bash
cd /root/modelcraft_project/modelcraft-go
git add internal/domain/modelruntime/model_resolver.go
git commit -m "feat: add createOneToManyFieldFromFK method for reverse relation GraphQL type"
```

---

## Chunk 3: 新增 createOneToManyResolverFromFK 方法（核心逻辑）

### Task 3: 新增 createOneToManyResolverFromFK 方法核心逻辑

**文件**：`internal/domain/modelruntime/model_resolver.go`

在 `createOneToManyFieldFromFK` 方法后面添加新方法。

- [ ] **Step 1: 实现 createOneToManyResolverFromFK 方法**

```go
// createOneToManyResolverFromFK 创建一对多关系解析器（基于 LogicalForeignKey 的 reverse 方向）
// 对于 reverse 方向的 LFK：
//   - lf.SourceFields = 当前模型（User）用于查询的列，比如 ["id"]
//   - lf.TargetFields = 关联模型（Order）中要匹配的列，比如 ["user_id"]
// 查询逻辑：将当前记录的 SourceFields 值作为 WHERE 条件中 TargetFields 的值
func (r *graphqlModelResolver) createOneToManyResolverFromFK(
	lf *modeldesign.LogicalForeignKey,
	refModelName string,
) graphql.FieldResolveFn {
	return func(p graphql.ResolveParams) (interface{}, error) {
		logger := logfacade.GetLogger(r.ctx)

		// 获取父对象记录
		record, ok := p.Source.(map[string]any)
		if !ok {
			logger.Warn(r.ctx, "invalid source type for one-to-many relation")
			return []map[string]any{}, nil
		}

		logger.Infof(r.ctx, "resolving one-to-many relation: record=%+v", record)

		// Step 1: 从当前记录中提取 SourceFields 的所有值
		sourceValues := make([]any, 0, len(lf.SourceFields))
		for _, sourceField := range lf.SourceFields {
			value, exists := record[sourceField]
			if !exists || value == nil {
				// 如果任意一个 SourceField 值为 nil，返回空数组
				logger.Infof(r.ctx, "one-to-many relation: source field %s is nil or missing, returning empty array", sourceField)
				return []map[string]any{}, nil
			}
			sourceValues = append(sourceValues, value)
		}

		logger.Infof(r.ctx, "extracted source values: %v", sourceValues)

		// Step 2: 构建 WHERE 条件 - zip(TargetFields, sourceValues)
		whereMap := make(map[string]any)
		for i, targetField := range lf.TargetFields {
			if i < len(sourceValues) {
				whereMap[targetField] = sourceValues[i]
			}
		}

		logger.Infof(r.ctx, "querying one-to-many relation: TableName=%s, WHERE=%+v",
			lf.RefModelName, whereMap)

		// Step 3: 调用 FindMany 查询关联记录
		results, err := r.clientRepo.FindMany(r.ctx, &FindManyInput{
			TableName: lf.RefModelName,
			Where:     whereMap,
		})
		if err != nil {
			logger.Errorf(r.ctx, "failed to query one-to-many relation: %v", err)
			return nil, err
		}

		if results == nil {
			return []map[string]any{}, nil
		}

		return results, nil
	}
}
```

- [ ] **Step 2: Verify imports at top of file**

确保文件已有这些 imports（通常都有）：
- `github.com/graphql-go/graphql`
- `modelcraft/internal/domain/modeldesign`
- `modelcraft/pkg/logfacade`

- [ ] **Step 3: Run go fmt to format code**

```bash
cd /root/modelcraft_project/modelcraft-go
go fmt ./internal/domain/modelruntime/model_resolver.go
```

- [ ] **Step 4: Compile check**

```bash
cd /root/modelcraft_project/modelcraft-go
go build ./internal/domain/modelruntime/
```

Expected output: No errors

- [ ] **Step 5: Commit**

```bash
cd /root/modelcraft_project/modelcraft-go
git add internal/domain/modelruntime/model_resolver.go
git commit -m "feat: add createOneToManyResolverFromFK for reverse relation query execution"
```

---

## Chunk 4: 新增单元测试

### Task 4: 添加 TestCreateRelationFieldReverse 测试

**文件**：`internal/domain/modelruntime/model_resolver_test.go`

在现有测试之后添加新的测试函数。

- [ ] **Step 1: 在测试文件末尾添加测试函数**

在 `model_resolver_test.go` 最后添加：

```go
// TestCreateRelationFieldReverse 测试 reverse 方向的关联字段创建
func TestCreateRelationFieldReverse(t *testing.T) {
	tests := []struct {
		name      string
		lf        *modeldesign.LogicalForeignKey
		sourceObj *graphql.Object
		wantErr   bool
		checkFunc func(t *testing.T, field *graphql.Field)
	}{
		{
			name: "创建一对多关联字段 - reverse 方向",
			lf: &modeldesign.LogicalForeignKey{
				ID:           "lf-reverse-1",
				Direction:    modeldesign.DirectionReverse,
				ModelName:    "User",
				RefModelName: "Order",
				SourceFields: []string{"id"},
				TargetFields: []string{"user_id"},
			},
			sourceObj: graphql.NewObject(graphql.ObjectConfig{
				Name: "Order",
				Fields: graphql.Fields{
					"id": &graphql.Field{Type: graphql.ID},
					"amount": &graphql.Field{Type: graphql.Float},
				},
			}),
			wantErr: false,
			checkFunc: func(t *testing.T, field *graphql.Field) {
				assert.NotNil(t, field)
				assert.NotNil(t, field.Type)
				// 验证类型是 [Order!]!
				listType, ok := field.Type.(*graphql.List)
				assert.True(t, ok, "Field type should be a List")
				assert.NotNil(t, listType.OfType)
				nonNullType, ok := listType.OfType.(*graphql.NonNull)
				assert.True(t, ok, "List item type should be NonNull")
				objectType, ok := nonNullType.OfType.(*graphql.Object)
				assert.True(t, ok, "NonNull item type should be Object")
				assert.Equal(t, "Order", objectType.Name())
				// 验证 Resolve 函数存在
				assert.NotNil(t, field.Resolve)
			},
		},
		{
			name: "创建复合外键的一对多关联字段",
			lf: &modeldesign.LogicalForeignKey{
				ID:           "lf-reverse-2",
				Direction:    modeldesign.DirectionReverse,
				ModelName:    "User",
				RefModelName: "Order",
				SourceFields: []string{"org_id", "id"},
				TargetFields: []string{"org_id", "user_id"},
			},
			sourceObj: graphql.NewObject(graphql.ObjectConfig{
				Name: "Order",
				Fields: graphql.Fields{
					"id": &graphql.Field{Type: graphql.ID},
				},
			}),
			wantErr: false,
			checkFunc: func(t *testing.T, field *graphql.Field) {
				assert.NotNil(t, field)
				assert.NotNil(t, field.Type)
				// 验证类型结构
				listType, ok := field.Type.(*graphql.List)
				assert.True(t, ok)
				assert.NotNil(t, listType.OfType)
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			resolver := &graphqlModelResolver{
				ctx:     context.Background(),
				lfkRepo: nil, // Not needed for this test
			}

			graphqlField := &graphql.Field{
				Name:        "orders",
				Description: "关联的订单列表",
			}

			field := resolver.createOneToManyFieldFromFK(tt.lf, tt.sourceObj, graphqlField)

			if tt.wantErr {
				assert.Nil(t, field)
			} else {
				tt.checkFunc(t, field)
			}
		})
	}
}

// TestCreateOneToManyResolverFromFK 测试 reverse 方向关系解析器
func TestCreateOneToManyResolverFromFK(t *testing.T) {
	tests := []struct {
		name           string
		lf             *modeldesign.LogicalForeignKey
		sourceRecord   map[string]any
		expectedWhere  map[string]any
		expectedResult []map[string]any
		wantErr        bool
	}{
		{
			name: "简单一对多关系 - 查询成功",
			lf: &modeldesign.LogicalForeignKey{
				ID:           "lf-1",
				Direction:    modeldesign.DirectionReverse,
				ModelName:    "User",
				RefModelName: "Order",
				SourceFields: []string{"id"},
				TargetFields: []string{"user_id"},
			},
			sourceRecord: map[string]any{
				"id":   "user-1",
				"name": "Alice",
			},
			expectedWhere: map[string]any{
				"user_id": "user-1",
			},
			expectedResult: []map[string]any{
				{"id": "order-1", "amount": 100},
				{"id": "order-2", "amount": 200},
			},
			wantErr: false,
		},
		{
			name: "复合外键 - 两个字段",
			lf: &modeldesign.LogicalForeignKey{
				ID:           "lf-2",
				Direction:    modeldesign.DirectionReverse,
				ModelName:    "User",
				RefModelName: "Order",
				SourceFields: []string{"org_id", "id"},
				TargetFields: []string{"org_id", "user_id"},
			},
			sourceRecord: map[string]any{
				"org_id": "org-1",
				"id":     "user-1",
				"name":   "Bob",
			},
			expectedWhere: map[string]any{
				"org_id":  "org-1",
				"user_id": "user-1",
			},
			expectedResult: []map[string]any{
				{"id": "order-1", "org_id": "org-1"},
			},
			wantErr: false,
		},
		{
			name: "源字段为 nil - 返回空数组",
			lf: &modeldesign.LogicalForeignKey{
				ID:           "lf-3",
				Direction:    modeldesign.DirectionReverse,
				ModelName:    "User",
				RefModelName: "Order",
				SourceFields: []string{"id"},
				TargetFields: []string{"user_id"},
			},
			sourceRecord: map[string]any{
				"id":   nil,
				"name": "Charlie",
			},
			expectedWhere:  nil,
			expectedResult: []map[string]any{},
			wantErr:        false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Mock ClientDatabaseRepository
			mockClientRepo := &mockClientDatabaseRepository{
				expectedWhere: tt.expectedWhere,
				result:        tt.expectedResult,
			}

			resolver := &graphqlModelResolver{
				ctx:        context.Background(),
				clientRepo: mockClientRepo,
			}

			resolverFn := resolver.createOneToManyResolverFromFK(tt.lf, "Order")

			// Create mock ResolveParams
			params := graphql.ResolveParams{
				Source:  tt.sourceRecord,
				Context: context.Background(),
			}

			result, err := resolverFn(params)

			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, result)
				resultArray, ok := result.([]map[string]any)
				assert.True(t, ok)
				assert.Equal(t, len(tt.expectedResult), len(resultArray))
			}
		})
	}
}

// mockClientDatabaseRepository 模拟的数据库仓储
type mockClientDatabaseRepository struct {
	expectedWhere map[string]any
	result        []map[string]any
}

func (m *mockClientDatabaseRepository) FindUnique(ctx context.Context, input *FindUniqueInput) (map[string]any, error) {
	return nil, nil
}

func (m *mockClientDatabaseRepository) FindFirst(ctx context.Context, input *FindFirstInput) (map[string]any, error) {
	return nil, nil
}

func (m *mockClientDatabaseRepository) FindMany(ctx context.Context, input *FindManyInput) ([]map[string]any, error) {
	// 简单验证 WHERE 条件（实际测试可以更严格）
	if m.expectedWhere != nil {
		for k, v := range m.expectedWhere {
			if whereVal, ok := input.Where[k]; !ok || whereVal != v {
				return []map[string]any{}, nil
			}
		}
	}
	return m.result, nil
}

func (m *mockClientDatabaseRepository) Aggregate(ctx context.Context, input *AggregateInput) (map[string]any, error) {
	return nil, nil
}

func (m *mockClientDatabaseRepository) Count(ctx context.Context, input *CountInput) (map[string]any, error) {
	return nil, nil
}

func (m *mockClientDatabaseRepository) CreateOne(ctx context.Context, input *CreateOneInput) (any, error) {
	return nil, nil
}

func (m *mockClientDatabaseRepository) UpdateOne(ctx context.Context, input *UpdateOneInput) (map[string]any, error) {
	return nil, nil
}

func (m *mockClientDatabaseRepository) DeleteOne(ctx context.Context, input *DeleteOneInput) (map[string]any, error) {
	return nil, nil
}

func (m *mockClientDatabaseRepository) CreateMany(ctx context.Context, input *CreateManyInput) ([]any, error) {
	return nil, nil
}

func (m *mockClientDatabaseRepository) UpdateMany(ctx context.Context, input *UpdateManyInput) (map[string]any, error) {
	return nil, nil
}

func (m *mockClientDatabaseRepository) DeleteMany(ctx context.Context, input *DeleteManyInput) (map[string]any, error) {
	return nil, nil
}
```

- [ ] **Step 2: Run tests to verify they pass**

```bash
cd /root/modelcraft_project/modelcraft-go
go test -v ./internal/domain/modelruntime/... -run TestCreateRelationFieldReverse
```

Expected: Tests pass

- [ ] **Step 3: Run tests to verify one-to-many resolver**

```bash
cd /root/modelcraft_project/modelcraft-go
go test -v ./internal/domain/modelruntime/... -run TestCreateOneToManyResolverFromFK
```

Expected: Tests pass

- [ ] **Step 4: Run all modelruntime tests to ensure no regression**

```bash
cd /root/modelcraft_project/modelcraft-go
go test -v ./internal/domain/modelruntime/...
```

Expected: All tests pass

- [ ] **Step 5: Commit**

```bash
cd /root/modelcraft_project/modelcraft-go
git add internal/domain/modelruntime/model_resolver_test.go
git commit -m "test: add TestCreateRelationFieldReverse and TestCreateOneToManyResolverFromFK"
```

---

## Chunk 5: 集成验证

### Task 5: 集成测试与手动验证

- [ ] **Step 1: 运行全套测试**

```bash
cd /root/modelcraft_project/modelcraft-go
go test -v ./internal/domain/modelruntime/...
```

Expected: All tests pass

- [ ] **Step 2: 运行 golangci-lint 代码检查**

```bash
cd /root/modelcraft_project/modelcraft-go
golangci-lint run ./internal/domain/modelruntime/model_resolver.go
```

Expected: No errors

- [ ] **Step 3: 构建整个项目**

```bash
cd /root/modelcraft_project/modelcraft-go
go build ./...
```

Expected: Build succeeds

- [ ] **Step 4: 创建集成测试文档注释**

在 `model_resolver.go` 的新方法上方添加详细的文档注释，说明：
- 方法作用
- 参数含义
- 返回值
- 复合外键的处理方式

- [ ] **Step 5: Final commit for integration verification**

```bash
cd /root/modelcraft_project/modelcraft-go
git log --oneline -5
```

Expected output showing 4-5 commits for this feature

---

## 验收标准

✅ `createRelationField` 正确分支处理 normal 和 reverse 方向  
✅ `createOneToManyFieldFromFK` 返回 GraphQL List 类型  
✅ `createOneToManyResolverFromFK` 支持复合外键  
✅ 任意 SourceField 为 nil 时返回空数组  
✅ 单元测试覆盖基本场景和复合外键场景  
✅ 所有现有测试仍通过（无回归）  
✅ 代码通过 golangci-lint 检查  
✅ 项目成功构建

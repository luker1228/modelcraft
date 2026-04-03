# Runtime 反向关联查询设计文档

**日期**: 2026-03-26  
**状态**: 待实现  

---

## 1. 背景

ModelCraft runtime 查询系统基于 GraphQL，支持用户对其数据库进行动态查询。

当前系统已支持 **多对一（normal）** 方向的关联查询：
- 例如：Order 模型有 `user` 字段（`RelateFKID` 指向 `direction=normal` 的 LFK）
- 查询 Order 时，可以嵌套查出关联的 User 记录（单个对象）

**本次需求**：支持 **一对多（reverse）** 方向的关联查询：
- 例如：User 模型有 `orders` 字段（`RelateFKID` 指向 `direction=reverse` 的 LFK）
- 查询 User 时，可以嵌套查出所有关联的 Order 记录（对象数组）

---

## 2. 现有数据结构

### LogicalForeignKey

每个 FK 关系在数据库中存有**两条**记录，共享同一个 `PairID`：

```
direction=normal（Order 侧，拥有 FK 列）:
  ModelName:    "Order"
  RefModelName: "User"
  SourceFields: ["user_id"]   // Order 表的 FK 列
  TargetFields: ["id"]        // User 表的主键列

direction=reverse（User 侧，被引用）:
  ModelName:    "User"
  RefModelName: "Order"
  SourceFields: ["id"]        // 当前模型（User）用于查询的列
  TargetFields: ["user_id"]   // 关联模型（Order）中要匹配的列
```

### RuntimeField（RELATION 类型）

用户在 User 模型上显式添加了一个 `orders` 字段：
- `Type.Format = FormatRelation`
- `RelateFKID` 指向 `direction=reverse` 的 LFK 记录

---

## 3. 需求范围

- **支持**：一对多（reverse）查询，返回全部关联记录（`[]Object`）
- **不支持（本次）**：关联查询内的 where / orderBy / limit 过滤
- **不支持（本次）**：depth > 1（仍保持 maxDepth=1）
- **不变**：多对一（normal）现有逻辑不受影响

---

## 4. 设计方案

### 4.1 架构层次

本次改动只涉及 **Domain 层** (`internal/domain/modelruntime/model_resolver.go`)，不涉及其他层。

### 4.2 核心判断点

在 `createRelationField` 中，拿到 `lf` 之后，根据 `lf.Direction` 分支：

```
lf.Direction == "normal"  →  现有逻辑：createManyToOneFieldFromFK（返回单个对象）
lf.Direction == "reverse" →  新逻辑：createOneToManyFieldFromFK（返回对象数组）
```

### 4.3 新增：createOneToManyFieldFromFK

GraphQL 字段类型：`graphql.NewList(graphql.NewNonNull(referenceObj))`

### 4.4 新增：createOneToManyResolverFromFK

**Resolver 逻辑**（支持复合外键）：

```
1. 从 p.Source（当前记录 map）中，按 lf.SourceFields 按顺序取出所有字段值
2. 检查：如果任意 SourceField 的值为 nil，返回空数组 []
3. 构建 WHERE 条件：zip(lf.TargetFields, values) → map[string]any
   - 使用 AND 连接多个条件（示例见下）
4. 调用 clientRepo.FindMany(ctx, &FindManyInput{
     TableName: lf.RefModelName,
     Where: whereMap
   })
5. 返回结果：
   - 如果 FindMany 返回空 slice → 返回 []
   - 如果出错 → 返回错误（同多对一逻辑）
```

**复合外键示例**：
```
假设 Order 表有复合 FK: (org_id, user_id)

lf.SourceFields = ["org_id", "id"]          // User 模型的列
lf.TargetFields = ["org_id", "user_id"]     // Order 模型的列

当前 User 记录: { org_id: "org-1", id: "user-1" }

提取值: values = ["org-1", "user-1"]

WHERE 条件: { org_id: "org-1", user_id: "user-1" }
  (两个条件用 AND 连接)
```

**错误处理**：
- 如果 findMany 返回错误 → 返回错误（遵循多对一错误处理模式）
- 如果 findMany 返回空 slice → 返回空数组（GraphQL 中显示为 `[]`）
- 如果 findMany 返回数据 → 返回 []map[string]any

**GraphQL 返回类型约定**：
- Resolver 返回类型：`[]map[string]any`（可能为空 slice）
- GraphQL 类型定义：`graphql.NewList(graphql.NewNonNull(referenceObj))`
- 即：非空对象的非空列表（`[RefModel!]!`），空列表有效

---

## 5. 变更文件

| 文件 | 改动内容 |
|------|---------|
| `internal/domain/modelruntime/model_resolver.go` | 在 `createRelationField` 中添加方向判断；新增 `createOneToManyFieldFromFK` 和 `createOneToManyResolverFromFK` 两个方法 |

**不需要改动**：
- DML 层（`FindMany` 已支持 where 条件查询）
- Application 层
- Interface 层
- 数据库 Schema

---

## 6. GraphQL Schema 变化示例

### 查询示例

```graphql
query {
  findMany(where: { name: { eq: "Alice" } }) {
    items {
      id
      name
      orders {        # 新增：一对多，返回数组
        id
        amount
        status
      }
    }
  }
}
```

### 返回示例

```json
{
  "items": [
    {
      "id": "user-1",
      "name": "Alice",
      "orders": [
        { "id": "order-1", "amount": 100, "status": "paid" },
        { "id": "order-2", "amount": 200, "status": "pending" }
      ]
    }
  ]
}
```

---

## 7. 边界情况

| 情况 | 处理方式 |
|------|---------|
| 任意 SourceField 值为 nil | 返回空数组 `[]` |
| 关联表无匹配记录 | 返回空数组 `[]` |
| 复合外键（多字段） | 全部字段都加入 WHERE 条件，用 AND 连接 |
| maxDepth=0 时的 reverse 字段 | 跳过（与 normal 一致，由现有深度控制逻辑处理） |
| FindMany 调用失败 | 返回错误（遵循多对一错误处理模式） |
| SourceField 在当前记录中不存在 | 按不存在处理，返回 nil/undefined，导致返回空数组 |

---

## 8. 实现对标的现有逻辑

本次实现应参考多对一（normal）的实现模式：
- 错误处理：使用相同的 `logger.Error` 和 `bizerrors` 模式
- 类型断言：使用 `cast.ToXxxE` 进行安全类型转换
- 字段查询：确保 SourceFields 存在再访问（防止 panic）

---

## 9. 不在范围内

- 关联查询内的 where / orderBy / take / skip 过滤参数
- N+1 查询优化（DataLoader / 批量查询）
- depth > 1 嵌套
- Mutation 中的关联写入

# CLI 数据命令

---

## 1. 概述

数据命令对应 Runtime GraphQL 的查询和变更操作。CLI 将用户参数转换为 GraphQL 请求，发送到：

```
POST /graphql/org/{orgName}/project/{projectSlug}/db/{db}/model/{model}
Authorization: Bearer <enduser-jwt>
```

所有命令共享相同的资源路径解析和输出格式。

---

## 2. `mc query` — 查询多条记录

对应 Runtime GraphQL 的 `findMany`。

### 用法

```bash
mc query <resource-path> [flags]
```

### 标志

| 标志 | 类型 | 必填 | 默认值 | 说明 |
|------|------|------|--------|------|
| `--where` | JSON | 否 | 无 | 过滤条件 |
| `--select` | JSON Array | 否 | 全部字段 | 返回字段列表 |
| `--orderBy` | JSON | 否 | 无 | 排序 `{"field":"asc\|desc"}` |
| `--take` | int | 否 | 20 | 返回条数（服务端有上限） |
| `--skip` | int | 否 | 0 | 跳过条数（分页） |

### 示例

```bash
mc query sales.maindb.users \
  --where '{"username":{"contains":"alice"}}' \
  --select '["id","username","email","createdAt"]' \
  --orderBy '{"createdAt":"desc"}' \
  --take 20 \
  --skip 0
```

### 输出

```json
{
  "ok": true,
  "data": [
    {
      "id": "01944...",
      "username": "alice",
      "email": "alice@example.com",
      "createdAt": "2026-01-15T08:30:00Z"
    }
  ],
  "meta": {
    "count": 1,
    "take": 20,
    "skip": 0,
    "hasMore": false
  }
}
```

---

## 3. `mc get` — 查询单条记录

对应 Runtime GraphQL 的 `findUnique`。

### 用法

```bash
mc get <resource-path> --where '{"id":{"equals":"01944..."}}' [flags]
```

### 标志

| 标志 | 类型 | 必填 | 默认值 | 说明 |
|------|------|------|--------|------|
| `--where` | JSON | 是 | — | 唯一标识条件 |
| `--select` | JSON Array | 否 | 全部字段 | 返回字段列表 |

### 输出

```json
{
  "ok": true,
  "data": {
    "id": "01944...",
    "username": "alice",
    "email": "alice@example.com"
  }
}
```

找不到时：

```json
{
  "ok": false,
  "error": {
    "code": "NOT_FOUND",
    "message": "No record found matching the given criteria",
    "retryable": false
  }
}
```

---

## 4. `mc create` — 创建记录

对应 Runtime GraphQL 的 `createOne`。

### 用法

```bash
mc create <resource-path> --data '<json>' [flags]
```

### 标志

| 标志 | 类型 | 必填 | 默认值 | 说明 |
|------|------|------|--------|------|
| `--data` | JSON | 是 | — | 待创建的数据 |
| `--select` | JSON Array | 否 | 全部字段 | 返回字段列表 |

### 示例

```bash
mc create sales.maindb.users \
  --data '{"username":"bob","email":"bob@example.com","age":25}'
```

### 输出

```json
{
  "ok": true,
  "data": {
    "id": "01944...",
    "username": "bob",
    "email": "bob@example.com",
    "age": 25,
    "createdAt": "2026-04-25T10:00:00Z"
  }
}
```

---

## 5. `mc update` — 更新记录

对应 Runtime GraphQL 的 `updateOne`。

### 用法

```bash
mc update <resource-path> --where '<json>' --data '<json>' [flags]
```

### 标志

| 标志 | 类型 | 必填 | 默认值 | 说明 |
|------|------|------|--------|------|
| `--where` | JSON | 是 | — | 定位目标记录 |
| `--data` | JSON | 是 | — | 更新数据 |
| `--select` | JSON Array | 否 | 全部字段 | 返回字段列表 |

### 示例

```bash
mc update sales.maindb.users \
  --where '{"id":{"equals":"01944..."}}' \
  --data '{"email":"bob-new@example.com"}'
```

### 输出

```json
{
  "ok": true,
  "data": {
    "id": "01944...",
    "username": "bob",
    "email": "bob-new@example.com",
    "updatedAt": "2026-04-25T10:05:00Z"
  }
}
```

---

## 6. `mc delete` — 删除记录

对应 Runtime GraphQL 的 `deleteOne`。

### 用法

```bash
mc delete <resource-path> --where '<json>' [flags]
```

### 标志

| 标志 | 类型 | 必填 | 默认值 | 说明 |
|------|------|------|--------|------|
| `--where` | JSON | 是 | — | 定位目标记录 |

### 示例

```bash
mc delete sales.maindb.users \
  --where '{"id":{"equals":"01944..."}}'
```

### 输出

```json
{
  "ok": true,
  "data": {
    "id": "01944...",
    "username": "bob",
    "deletedAt": "2026-04-25T10:10:00Z"
  }
}
```

---

## 7. `mc count` — 计数

对应 Runtime GraphQL 的 `count`。

### 用法

```bash
mc count <resource-path> [--where '<json>']
```

### 示例

```bash
mc count sales.maindb.users \
  --where '{"status":{"equals":"active"}}'
```

### 输出

```json
{
  "ok": true,
  "count": 42
}
```

---

## 8. `mc aggregate` — 聚合查询

对应 Runtime GraphQL 的 `aggregate`。

### 用法

```bash
mc aggregate <resource-path> [flags]
```

### 标志

| 标志 | 类型 | 必填 | 默认值 | 说明 |
|------|------|------|--------|------|
| `--where` | JSON | 否 | 无 | 过滤条件 |
| `--avg` | JSON Array | 否 | — | 求平均的字段列表 |
| `--sum` | JSON Array | 否 | — | 求和的字段列表 |
| `--min` | JSON Array | 否 | — | 求最小值的字段列表 |
| `--max` | JSON Array | 否 | — | 求最大值的字段列表 |

至少需要提供 `--avg`、`--sum`、`--min`、`--max` 中的一个。

### 示例

```bash
mc aggregate sales.maindb.orders \
  --avg '["amount"]' \
  --sum '["amount"]' \
  --min '["amount"]' \
  --max '["amount"]' \
  --where '{"status":{"equals":"completed"}}'
```

### 输出

```json
{
  "ok": true,
  "aggregate": {
    "_avg": { "amount": 156.78 },
    "_sum": { "amount": 6271.20 },
    "_min": { "amount": 12.00 },
    "_max": { "amount": 999.99 },
    "_count": 40
  }
}
```

---

## 9. `--where` 过滤语法

JSON 格式，与 Runtime GraphQL Where Input 完全对齐。

### 9.1 按字段类型支持的操作符

| 字段类型 | 操作符 |
|----------|--------|
| String | `equals`, `not`, `in`, `contains`, `startsWith`, `endsWith`, `mode` |
| Int / Float | `equals`, `not`, `in`, `lt`, `lte`, `gt`, `gte` |
| Boolean | `equals`, `not` |
| DateTime | `equals`, `not`, `in`, `lt`, `lte`, `gt`, `gte` |

### 9.2 逻辑组合

```json
{
  "AND": [
    { "status": { "equals": "active" } },
    { "age": { "gte": 18 } }
  ]
}
```

```json
{
  "OR": [
    { "role": { "equals": "admin" } },
    { "role": { "equals": "manager" } }
  ]
}
```

```json
{
  "NOT": { "status": { "equals": "deleted" } }
}
```

### 9.3 组合嵌套

```json
{
  "AND": [
    { "status": { "equals": "active" } },
    {
      "OR": [
        { "role": { "equals": "admin" } },
        { "department": { "equals": "engineering" } }
      ]
    }
  ]
}
```

---

## 10. GraphQL 映射

CLI 命令到 Runtime GraphQL 的转换示例：

### mc query → findMany

```bash
mc query sales.maindb.users --where '{"age":{"gte":18}}' --take 10 --skip 0 --orderBy '{"createdAt":"desc"}'
```

转换为：

```graphql
query {
  findManyUsers(
    where: { age: { gte: 18 } }
    take: 10
    skip: 0
    orderBy: { createdAt: desc }
  ) {
    id
    username
    email
    age
    createdAt
  }
}
```

发送到：

```
POST /graphql/org/acme/project/sales/db/maindb/model/users
Authorization: Bearer eyJ...
Content-Type: application/json

{"query": "..."}
```

### mc create → createOne

```bash
mc create sales.maindb.users --data '{"username":"bob","email":"bob@example.com"}'
```

转换为：

```graphql
mutation {
  createOneUsers(data: { username: "bob", email: "bob@example.com" }) {
    id
    username
    email
    createdAt
  }
}
```

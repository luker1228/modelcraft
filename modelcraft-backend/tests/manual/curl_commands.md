# ModelCraft API 手动测试 - cURL 命令集合

本文档包含所有 ModelCraft API 接口的 cURL 测试命令。

## 🔧 测试环境设置

**服务地址**: `http://localhost:8080`  
**Content-Type**: `application/json`

## 📋 接口测试命令

### 1. 基础模型操作

#### 1.1 创建模型
```bash
curl -X POST http://localhost:8080/api/design/models/createModel \
  -H "Content-Type: application/json" \
  -d '{
    "title": "用户模型",
    "name": "user",
    "description": "用户信息管理",
    "storageType": "mysql",
    "editable": true,
    "deletable": false,
    "tags": ["用户", "核心"],
    "createdBy": "admin",
    "fields": [
      {
        "key": "username",
        "title": "用户名",
        "format": "id_like_text",
        "required": true,
        "maxLength": 50
      }
    ]
  }'
```

#### 1.2 查询模型列表
```bash
curl -X GET http://localhost:8080/api/design/models
```

#### 1.3 获取单个模型
```bash
# 替换 {modelId} 为实际的模型ID
curl -X GET http://localhost:8080/api/models/{modelId}
```

#### 1.4 更新模型
```bash
# 替换 {modelId} 为实际的模型ID
curl -X PUT http://localhost:8080/api/models/{modelId} \
  -H "Content-Type: application/json" \
  -d '{
    "name": "更新后的用户模型",
    "description": "更新后的描述",
    "tags": ["用户", "核心", "更新"]
  }'
```

#### 1.5 删除模型
```bash
# 替换 {modelId} 为实际的模型ID
curl -X DELETE http://localhost:8080/api/models/{modelId}

curl -X DELETE http://localhost:8080/api/design/models/d1156cbc-fc70-4031-910c-b7b28d2e322f

```

#### 1.6 数据流验证
curl -X POST "http://localhost:8080/graphql/base_client/modelcraft_client/master" -d'{"query": "query { findFirst(where: {id:{equals:\"1\"}}) { id } }"}'


curl -X POST "http://localhost:8080/graphql/base_client/modelcraft_client/master" -d'{"query": "query { findFirst { id } }"}'


curl -X POST http://localhost:8080/graphql/base_client/modelcraft_client/master -d'{
"query": "mutation { create(data:{desc:\"ken\"}) { id \n createdObj{desc} } }"
}'



curl -X POST http://localhost:8080/graphql/user -d'{
"query": "query { findFirst(where: {id:{contains:\"1\"}}) { id username} }"
}'

curl -X POST http://localhost:8080/graphql/user -d'{
"query": "query { findFirst(where: {id:{not:\"1\"}}) { id username} }"
}'

curl -X POST http://localhost:8080/graphql/user -d'{
"query": "query { findUnique(where: {id:\"1\"}) { id username} }"
}'

curl -X POST http://localhost:8080/graphql/user -d'{
"query": "query { findUnique(where: {id:\"3\"}) { id username} }"
}'


curl -X POST http://localhost:8080/graphql/user -d'{
"query": "query { findMany(where: {id:{equals:\"1\"}}) { id username} }"
}'

curl -X POST http://localhost:8080/graphql/user -d'{
"query": "query { findMany(where: {id:{in:[\"1\", \"2\", \"3\"]}}) { id username} }"
}'

curl -X POST http://localhost:8080/graphql/user -d'{
"query": "query { findMany(where: {id:{not:\"1\", equals:\"3\"}}) { id username} }"
}'

curl -X POST http://localhost:8080/graphql/user -d'{
"query": "query { findMany(where: {_or:[{id:{not:\"1\"}}, {id:{not:\"2\"}}]}) { id username} }"
}'


curl -X POST http://localhost:8080/graphql/user -d'{
"query": "mutation { createOne(data:{username:\"ken\"}) { id } }"
}'
### 2. 模型数据验证

#### 2.1 验证模型数据
```bash
# 替换 {modelId} 为实际的模型ID
curl -X POST http://localhost:8080/api/models/{modelId}/validate \
  -H "Content-Type: application/json" \
  -d '{
    "username": "testuser",
    "email": "test@example.com"
  }'
```

### 3. Schema 导入导出

#### 3.1 导出模型 Schema
```bash
# 替换 {modelId} 为实际的模型ID
curl -X GET http://localhost:8080/api/models/{modelId}/modeldesign
```

#### 3.2 导入模型 Schema
```bash
# 替换 {modelId} 为实际的模型ID
curl -X PUT http://localhost:8080/api/models/{modelId}/modeldesign \
  -H "Content-Type: application/json" \
  -d '{
    "version": "1.0",
    "fields": [
      {
        "key": "newField",
        "name": "新字段",
        "type": "string",
        "required": false
      }
    ]
  }'
```

### 4. 字段管理

#### 4.1 添加字段
```bash
# 替换 {modelId} 为实际的模型ID
curl -X POST http://localhost:8080/api/models/{modelId}/fields \
  -H "Content-Type: application/json" \
  -d '{
    "key": "age",
    "name": "年龄",
    "description": "用户年龄",
    "type": "number",
    "format": "integer",
    "required": false,
    "minimum": 0,
    "maximum": 150
  }'
```

#### 4.2 更新字段
```bash
# 替换 {modelId} 和 {fieldId} 为实际的ID
curl -X PUT http://localhost:8080/api/models/{modelId}/fields/{fieldId} \
  -H "Content-Type: application/json" \
  -d '{
    "name": "更新后的年龄",
    "description": "更新后的用户年龄描述",
    "required": true,
    "minimum": 18
  }'
```

#### 4.3 删除字段
```bash
# 替换 {modelId} 和 {fieldId} 为实际的ID
curl -X DELETE http://localhost:8080/api/models/{modelId}/fields/{fieldId}
```

### 5. 关联管理

#### 5.1 添加关联
```bash
# 替换 {modelId} 为实际的模型ID
curl -X POST http://localhost:8080/api/models/{modelId}/relations \
  -H "Content-Type: application/json" \
  -d '{
    "name": "用户订单",
    "targetModelId": "order_model_id",
    "relationType": "one_to_many",
    "foreignKey": "user_id",
    "description": "用户与订单的关联关系"
  }'
```

#### 5.2 删除关联
```bash
# 替换 {modelId} 和 {relationId} 为实际的ID
curl -X DELETE http://localhost:8080/api/models/{modelId}/relations/{relationId}
```

### 6. 字段类型和格式查询

#### 6.1 验证字段定义
```bash
curl -X POST http://localhost:8080/api/fields/validate \
  -H "Content-Type: application/json" \
  -d '{
    "fields": [
      {
        "key": "testField",
        "name": "测试字段",
        "type": "string",
        "required": true
      }
    ]
  }'
```

#### 6.2 获取字段类型
```bash
curl -X GET http://localhost:8080/api/fields/types
```

#### 6.3 获取字段格式
```bash
curl -X GET http://localhost:8080/api/fields/formats
```

## 🚀 测试流程建议

### 完整测试流程：

1. **创建模型** (1.1) → 获取返回的 `modelId`
2. **查询模型列表** (1.2) → 验证模型已创建
3. **获取单个模型** (1.3) → 验证模型详情
4. **添加字段** (4.1) → 获取返回的 `fieldId`
5. **更新字段** (4.2) → 验证字段更新
6. **验证模型数据** (2.1) → 测试数据验证
7. **导出 Schema** (3.1) → 获取模型结构
8. **更新模型** (1.4) → 验证模型更新
9. **删除字段** (4.3) → 清理测试字段
10. **删除模型** (1.5) → 清理测试数据

### 快速验证流程：

1. **获取字段类型** (6.2)
2. **获取字段格式** (6.3)
3. **验证字段定义** (6.1)
4. **创建简单模型** (1.1)
5. **查询模型列表** (1.2)

## 📝 注意事项

1. **替换占位符**: 将 `{modelId}`, `{fieldId}`, `{relationId}` 替换为实际返回的 ID
2. **检查响应**: 注意查看 HTTP 状态码和响应内容
3. **数据格式**: 确保 JSON 格式正确
4. **服务状态**: 确保 ModelCraft 服务正在运行

## 🔍 响应示例

### 成功响应 (200/201):
```json
{
  "resourceId": "model_123456",
  "name": "用户模型",
  "identifier": "user",
  "status": "success"
}
```

### 错误响应 (400/404/500):
```json
{
  "error": "错误信息",
  "code": "ERROR_CODE",
  "details": "详细错误描述"
}
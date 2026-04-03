# ModelCraft API 手动测试

本目录包含 ModelCraft API 的手动测试工具和文档。

## 📁 文件说明

### 📋 测试文档
- **`curl_commands.md`** - 完整的 cURL 命令集合，包含所有 API 接口的测试命令
- **`README.md`** - 本文档，手动测试指南

### 🔧 测试工具
- **`test_scenarios.sh`** - 自动化测试场景脚本，包含多个测试流程
- **`postman_collection.json`** - Postman 测试集合，可直接导入 Postman 使用

## 🚀 快速开始

### 1. 使用 cURL 命令

查看 `curl_commands.md` 文件，复制相应的命令进行测试：

```bash
# 示例：创建模型
curl -X POST http://localhost:8080/api/models \
  -H "Content-Type: application/json" \
  -d '{
    "name": "测试模型",
    "identifier": "test_model",
    "storageType": "mysql"
  }'
```

### 2. 使用测试场景脚本

```bash
# 给脚本执行权限
chmod +x test_scenarios.sh

# 运行特定场景
./test_scenarios.sh basic_crud

# 运行所有测试场景
./test_scenarios.sh all

# 查看帮助
./test_scenarios.sh help
```

### 3. 使用 Postman

1. 打开 Postman
2. 导入 `postman_collection.json` 文件
3. 设置环境变量 `baseUrl` 为 `http://localhost:8080`
4. 按顺序执行请求

## 📋 测试场景

### 场景1: 基础模型 CRUD
测试模型的创建、查询、更新、删除操作

### 场景2: 字段管理
测试字段的添加、更新、删除，以及字段类型查询

### 场景3: 数据验证
测试模型数据的验证功能

### 场景4: Schema 导入导出
测试模型 Schema 的导出和导入功能

## 🔍 测试前准备

1. **确保服务运行**
   ```bash
   # 检查服务状态
   curl http://localhost:8080/health
   ```

2. **准备测试环境**
   - 确保 ModelCraft 服务在 `localhost:8080` 运行
   - 确保数据库连接正常
   - 清理之前的测试数据（如需要）

## 📊 测试结果验证

### 成功响应示例
```json
{
  "resourceId": "model_123456",
  "name": "测试模型",
  "status": "success"
}
```

### 错误响应示例
```json
{
  "error": "Invalid request",
  "code": "VALIDATION_ERROR",
  "details": "字段验证失败"
}
```

## 🔧 常见问题

### 1. 连接被拒绝
- 检查 ModelCraft 服务是否启动
- 确认端口号是否正确 (默认 8080)

### 2. 404 错误
- 检查 API 路径是否正确
- 确认路由是否已正确配置

### 3. 400 错误
- 检查请求体格式是否正确
- 验证必填字段是否提供

### 4. 500 错误
- 检查服务器日志
- 确认数据库连接是否正常

## 📝 测试记录

建议在测试过程中记录：
- 测试时间
- 测试场景
- 请求参数
- 响应结果
- 发现的问题

## 🔄 测试流程建议

1. **环境检查** - 确认服务状态
2. **基础功能** - 测试 CRUD 操作
3. **高级功能** - 测试字段管理、数据验证等
4. **边界测试** - 测试异常情况和边界条件
5. **清理数据** - 删除测试产生的数据

## 📞 支持

如果在测试过程中遇到问题，请：
1. 检查服务器日志
2. 查看 API 文档
3. 联系开发团队
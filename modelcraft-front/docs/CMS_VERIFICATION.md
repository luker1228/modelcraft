# CMS 功能验证指南

## 修复的问题

### 1. Google Fonts ORB 错误

**问题**: `ERR_BLOCKED_BY_ORB` - Opaque Response Blocking

**原因**: 浏览器安全策略阻止了跨域字体资源加载

**解决方案**: 使用 Next.js 的 `next/font/google` 本地化字体

**修改的文件**:
- `src/app/layout.tsx` - 导入并配置 Open Sans 和 Poppins 字体
- `src/app/globals.css` - 移除 `@import` CDN 链接，使用 CSS 变量
- `tailwind.config.ts` - 更新字体配置使用 CSS 变量

**验证**:
```bash
# 重启开发服务器
cd modelcraft-react
npm run dev
```

打开浏览器控制台，应该不再看到 `ERR_BLOCKED_BY_ORB` 错误。字体会自动从 Next.js 优化的本地缓存加载。

## CMS 功能完整性验证

### 架构流程

```
┌─────────────────────────────────────────────────────────┐
│          Frontend: /app/cms                             │
│          (modelcraft-react:3000)                        │
└───────┬─────────────────────────────────┬───────────────┘
        │                                 │
        │ Design-Time API                 │ Runtime API
        │ (Models, Schema)                │ (Content CRUD)
        ▼                                 ▼
┌──────────────────────┐        ┌──────────────────────┐
│  GraphQL Backend     │        │   Agent Backend      │
│  modelcraft-go:8080  │        │ modelcraft-agent:8000│
└──────────────────────┘        └──────────────────────┘
```

### 数据流

1. **列表页** (`/app/cms`)
   - 查询: `models(input: ModelQueryInput!)`
   - Client: `designTimeClient` (端口 8080)
   - 显示所有可用的 Model

2. **内容列表页** (`/app/cms/[modelId]`)
   - 查询 1: `model(projectName, id)` - 获取 Model 详情
   - 查询 2: `modelJsonSchema(projectName, id)` - 获取 JSON Schema
   - 查询 3: `findMany<ModelName>(take, skip)` - 获取内容列表
   - Client: 前两个用 `designTimeClient`，第三个用 `runtimeClient` (端口 8000)

3. **创建/编辑页** (`/app/cms/[modelId]/[contentId]`)
   - 查询: Schema + 单条内容
   - 变更: `createOne<ModelName>` 或 `updateOne<ModelName>`
   - Client: `runtimeClient`

### 自动化测试脚本

运行验证脚本:

```bash
cd modelcraft-react
./scripts/test-cms.sh
```

脚本会检查:
- ✅ GraphQL 服务是否运行 (8080)
- ✅ 能否查询 Models
- ✅ 能否获取 Model Schema
- ✅ Runtime API 是否运行 (8000)
- ✅ 能否查询内容

### 手动验证步骤

#### 前置条件

确保所有服务都在运行:

```bash
# 终端 1: 后端 API
cd modelcraft-go
task run

# 终端 2: AI Agent
cd modelcraft-agent
source .venv/bin/activate
make dev

# 终端 3: 前端
cd modelcraft-react
npm run dev
```

#### 测试步骤

**步骤 1: 访问 CMS 入口**

```
URL: http://localhost:3000/app/cms
```

预期结果:
- 显示 "Content Management" 标题
- 如果有 Model，显示 Model 卡片网格
- 每个卡片显示: 标题、描述、Cluster、Database
- 如果没有 Model，显示提示信息

检查点:
- [ ] 页面正常加载
- [ ] 能看到 Model 列表
- [ ] 控制台没有错误

**步骤 2: 进入内容管理**

点击任意 Model 卡片

```
URL: http://localhost:3000/app/cms/<model-id>
```

预期结果:
- 显示 Model 标题和描述
- 显示 "Create New" 按钮
- 显示内容列表表格
- 表头包含字段名称 (ID + 前 5 个字段)

检查点:
- [ ] Model 信息正确显示
- [ ] 如果有内容，表格正常显示
- [ ] "Edit" 和 "Delete" 按钮可见

**步骤 3: 创建内容**

点击 "Create New" 按钮

```
URL: http://localhost:3000/app/cms/<model-id>/new
```

预期结果:
- 根据 Schema 动态生成表单
- 字段类型对应表单控件:
  - String → Input
  - Number → Input (type=number)
  - Boolean → Checkbox
  - Enum → Select
  - DateTime → DatePicker

检查点:
- [ ] 表单字段完整
- [ ] 必填字段有标识
- [ ] 表单验证工作
- [ ] 提交后跳转回列表页

**步骤 4: 编辑内容**

点击列表中的 "Edit" 按钮

```
URL: http://localhost:3000/app/cms/<model-id>/<content-id>
```

预期结果:
- 表单预填充现有数据
- 可以修改字段
- "Save" 按钮可用

检查点:
- [ ] 数据正确加载
- [ ] 修改可以保存
- [ ] 保存后返回列表页

**步骤 5: 删除内容**

点击 "Delete" 按钮

预期结果:
- 弹出确认对话框
- 确认后删除
- 列表自动刷新

检查点:
- [ ] 确认对话框出现
- [ ] 删除成功
- [ ] 列表更新

### 常见问题排查

#### 问题 1: 看不到 Model

**症状**: `/app/cms` 显示 "No models found"

**检查**:
```bash
# 查询 Models
curl -X POST http://localhost:8080/graphql \
  -H "Content-Type: application/json" \
  -d '{
    "query": "query { models(input: {projectName: \"default\", databaseName: \"\", limit: 10, offset: 0}) { totalCount edges { node { id name } } } }"
  }' | jq .
```

**可能原因**:
- GraphQL 服务未运行
- `projectName` 不匹配 (代码中硬编码为 `"default"`)
- 数据库中没有 Model 数据

**解决**:
- 先访问 `/app/models` 创建 Model

#### 问题 2: 内容列表加载失败

**症状**: 进入 `/app/cms/[modelId]` 后显示错误

**检查**:
```bash
# 检查 Schema
curl -X POST http://localhost:8080/graphql \
  -H "Content-Type: application/json" \
  -d '{
    "query": "query { modelJsonSchema(projectName: \"default\", id: \"<model-id>\") { modelName schema } }"
  }' | jq .

# 检查 Runtime API
curl -s http://localhost:8000/graphql
```

**可能原因**:
- Model Schema 为空
- Runtime API (modelcraft-agent) 未运行
- GraphQL 表尚未生成

**解决**:
1. 确保 `modelcraft-agent` 运行在 8000 端口
2. 检查 Model 是否已经 "发布" (生成了表)

#### 问题 3: 创建/编辑表单不工作

**症状**: 点击 "Create New" 或 "Edit" 无响应

**检查浏览器控制台**:
- 查看是否有 GraphQL 错误
- 查看 Network 标签的请求

**可能原因**:
- FormRenderer 组件错误
- Schema 解析失败
- Mutation 失败

**解决**:
```bash
# 手动测试创建
curl -X POST http://localhost:8000/graphql \
  -H "Content-Type: application/json" \
  -d '{
    "query": "mutation { createOneUser(data: {name: \"Test\", email: \"test@example.com\"}) { id name } }"
  }' | jq .
```

#### 问题 4: projectName 上下文丢失

**症状**: 所有查询返回空或错误

**检查代码**:
```typescript
// 当前硬编码为 'default'
const projectName = 'default'
```

**长期解决方案**:
参考 `PROJECT_CONTEXT_FLOW.md`，应该从 Zustand store 获取:

```typescript
import { useProjectStore } from '@/stores/project-store'

const projectName = useProjectStore(state => state.selectedProject?.name)
```

### 性能检查

**页面加载时间**:
- `/app/cms` 首次加载: < 1s
- `/app/cms/[modelId]` 加载: < 2s (含 Schema 解析)

**数据量测试**:
- 100 条内容: 表格应流畅滚动
- 1000 条内容: 需要分页 (目前 `take: 50`)

### 安全检查

- [ ] projectName 验证 - 确保用户只能访问自己的项目
- [ ] Input 清理 - 防止 XSS
- [ ] GraphQL 权限 - 确保 Mutation 有权限控制

## 总结

完成以上验证步骤后，CMS 功能应该可以:

✅ 列出所有 Model
✅ 查看 Model 的内容列表
✅ 创建新内容
✅ 编辑现有内容
✅ 删除内容
✅ 表单根据 Schema 动态生成
✅ 支持各种字段类型

如果有任何步骤失败，参考上面的排查指南定位问题。

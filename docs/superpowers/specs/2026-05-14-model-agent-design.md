# model-agent 服务设计

> **日期**: 2026-05-14  
> **状态**: 已批准  
> **目标**: 独立 Python Agent 服务，通过 CopilotKit 协议让前端 AI 助手能执行数据查询和 NL2Filter 能力，并直接操控前端 FilterPanel UI 状态。

---

## 1. 问题陈述

前端 CopilotKit 已集成（`CopilotProvider`、`CopilotSidebar`、`runtimeUrl="/api/copilotkit"`），但后端 Agent 服务不存在，`/api/copilotkit` 路由也未实现。

用户希望通过 AI 助手：
1. **执行数据查询**：向 AI 说出查询意图，AI 调用 Backend GraphQL 返回结果
2. **NL2Filter**：自然语言 → ModelCraft filter JSON，Agent 直接操控 FilterPanel UI 状态

---

## 2. 整体架构

```
浏览器
  Authorization: Bearer <enduser-jwt>
       │
       ▼
Next.js BFF: /api/copilotkit/route.ts   ← 新建
  透传 Authorization header（与现有 BFF route 模式一致）
  透传 Cookie（refreshToken）
  注入 orgName + projectSlug 作为 CopilotKit properties
       │
       │  HTTP → modelcraft-agent:8000
       │  Authorization: Bearer <enduser-jwt>  ← header 原样透传
       ▼
modelcraft-agent (Python / FastAPI)   ← 新服务
  CopilotKit Python SDK
  Agent: "modelcraft_agent"
  LLM: DeepSeek / OpenAI-compatible

  从 FastAPI Request headers 读取 Authorization
  每次 Tool 调用时，将此 token 带入下游请求

  Python Tools:               Frontend Actions（前端注册）:
  ┌──────────────────────────────┐     ┌──────────────────────────┐
  │ query_model                  │     │ set_filter(filter_json)  │
  │ headers["Authorization"] =   │     │ clear_filter()           │
  │   Bearer <enduser-jwt>       │     └──────────────────────────┘
  └──────────────────────────────┘              │
       │                                        ▼
       │  Authorization: Bearer <enduser-jwt>  FilterPanel (React State)
       ▼                                  whereJsonCommitted 更新 → 触发查询
  Gateway (:8090)
  - 验签 JWT
  - 注入 X-User-ID（从 JWT claims 提取，供后端日志等使用）
  - Authorization: Bearer <jwt> 保留继续往后传
       │
       ▼
  Backend GraphQL (:8080)
  (收到 X-User-ID + Authorization header)
```

**关键设计决策：**
- **强制过 Gateway（:8090）**：Python Agent 的所有 GraphQL 调用必须发往 `http://gateway:8090`，禁止直连 `backend:8080`。Gateway 验签 JWT 后提取 `X-User-ID` 注入（供后端日志等），同时保留 `Authorization` header 继续透传给 Backend。
- **JWT 全程 Header 透传**：浏览器 → BFF（原样透传 Authorization header）→ Python Agent（从 FastAPI request.headers 读取）→ Gateway（每次 Tool 调用都带上 `Authorization: Bearer <token>`）。不走 CopilotKit properties，properties 只传业务上下文（orgName/projectSlug）。
- **LLM Provider**：通过环境变量切换（DeepSeek / OpenAI / Claude），当前使用 DeepSeek（OpenAI-compatible API）
- **UI 操控**：通过 CopilotKit Frontend Actions，Agent 可直接设置 FilterPanel 的筛选条件

---

## 3. Python 服务结构（`modelcraft-agent/`）

```
modelcraft-agent/
├── main.py                     # FastAPI 入口，注册 CopilotKit handler
├── agent.py                    # CopilotKit Agent 定义 + Tool 注册
├── tools/
│   ├── query_model.py          # Tool: 调用 Backend GraphQL 查询数据
│   └── nl2filter.py            # Tool: 自然语言 → filter JSON（通过 LLM 推理）
├── llm/
│   └── client.py               # 模型接入层：统一 LLM 客户端，支持 provider 切换
├── client/
│   └── graphql_client.py       # Gateway GraphQL HTTP client（httpx，每次调用透传 Authorization header）
├── config.py                   # 环境变量配置
├── requirements.txt
└── Dockerfile
```

### 3.1 模型接入层（`llm/client.py`）

通过环境变量选择 provider，agent.py 无需感知底层实现：

```
LLM_PROVIDER=deepseek        # deepseek | openai | anthropic
LLM_MODEL=deepseek-chat      # 模型名称
LLM_API_KEY=sk-xxx           # API Key
LLM_BASE_URL=https://api.deepseek.com  # OpenAI-compatible provider 时设置
```

- `deepseek` / `openai`：使用 `openai` Python SDK + `base_url` 切换
- `anthropic`：使用 `anthropic` Python SDK

### 3.2 Tool：`query_model`

```python
@copilotkit_tool
async def query_model(
    model_name: str,       # 模型名，如 "users"
    database: str,         # 数据库名，如 "maindb"
    filters: dict,         # ModelCraft filter JSON
    fields: list[str],     # 要返回的字段列表
    limit: int = 20,
) -> dict:
    """查询指定模型的数据，返回记录列表。"""
    # graphql_client 从当前请求 context 中读取 Authorization header
    # 每次调用 Gateway 时原样带上：Authorization: Bearer <enduser-jwt>
    # Gateway 验签后转发给 Backend
```

### 3.3 Tool：`nl2filter`

```python
@copilotkit_tool
async def nl2filter(
    natural_language: str,   # 自然语言描述，如"名字包含张"
    model_schema: dict,      # 模型字段定义（名称+类型）
) -> dict:
    """将自然语言转换为 ModelCraft filter JSON。
    
    返回示例：{"AND": [{"name": {"contains": "张"}}]}
    """
    # 通过 LLM prompt 直接推理，无需额外 API 调用
    # model_schema 作为上下文注入 prompt
```

**注**：`nl2filter` 本身就是一次 LLM 推理，agent 内部调用即可，不需要额外 HTTP 请求。

---

## 4. 前端接入层

### 4.1 BFF Route（必须）

```
新建：modelcraft-front/src/app/api/copilotkit/route.ts
```

与现有 BFF route（如 `/api/bff/graphql/org/[orgName]/project/[projectSlug]/route.ts`）保持**相同透传模式**：

```typescript
// 原样透传 Authorization header 和 Cookie
const authHeader = req.headers.get('Authorization')
if (authHeader) headers.set('Authorization', authHeader)
const cookieHeader = req.headers.get('cookie')
if (cookieHeader) headers.set('Cookie', cookieHeader)
```

Python Agent 收到请求后，从 `request.headers["authorization"]` 读取 token，后续每次调用 Gateway 的 Tool 都带上这个 header。

前端 `CopilotProvider` 已配置：
- `runtimeUrl="/api/copilotkit"` — 指向此 route
- `agent="modelcraft_agent"` — 对应 Python 侧的 agent 名称

### 4.2 Frontend Actions（UI 操控）

在 `EndUserRecordWorkspace` 组件内注册（需在 CopilotKit context 内，即在 `CopilotProvider` 包裹下的子组件中调用）：

```typescript
// Agent 调用此 action 设置筛选条件，FilterPanel 自动更新
useCopilotAction({
  name: "set_filter",
  description: "设置 FilterPanel 的 where 筛选条件，接受 ModelCraft filter JSON 字符串",
  parameters: [{ name: "filter_json", type: "string", required: true }],
  handler: async ({ filter_json }) => {
    setWhereJsonDraft(filter_json)
    setWhereJsonCommitted(filter_json)
  }
})

// Agent 调用此 action 清空筛选
useCopilotAction({
  name: "clear_filter",
  description: "清空 FilterPanel 的筛选条件",
  handler: async () => { handleClearFilter() }
})
```

---

## 5. 数据流示例

**场景：用户说"筛选名字包含张的用户"**

```
1. 用户输入 → CopilotSidebar
2. SDK 调用 /api/copilotkit (BFF)
3. BFF 注入 JWT，代理到 Python Agent
4. Agent 决策：调用 nl2filter("名字包含张", schema)
5. LLM 返回 filter JSON：{"AND": [{"name": {"contains": "张"}}]}
6. Agent 调用 Frontend Action: set_filter(filter_json)
7. FilterPanel whereJsonCommitted 更新
8. useQuery 触发 → GraphQL 查询 → 表格刷新
9. Agent 回复确认消息
```

**场景：用户说"查一下 users 里最新的 5 条数据"**

```
1. 用户输入 → CopilotSidebar
2. Agent 决策：调用 query_model("users", "maindb", {}, ["id","name","createdAt"], 5)
3. Python 构建 GraphQL 请求，带 Authorization: Bearer <JWT>
4. 请求发往 Gateway(:8090) → 验签 → 注入 X-User-ID → Authorization 保留 → 转发 Backend(:8080)
5. 返回数据列表
6. Agent 格式化后回复给用户
```

---

## 6. 环境配置

```bash
# modelcraft-agent/.env（新建）
LLM_PROVIDER=deepseek
LLM_MODEL=deepseek-chat
LLM_API_KEY=sk-xxx
LLM_BASE_URL=https://api.deepseek.com
GATEWAY_URL=http://gateway:8090       # 所有 GraphQL 调用必须过网关，禁止直连 backend:8080
# orgName / projectSlug 不在此配置，由每次请求的 CopilotKit properties 运行时注入

# modelcraft-front/.env（追加）
AGENT_SERVICE_URL=http://modelcraft-agent:8000
```

**URL 构建逻辑：** Python Agent 从 CopilotKit 消息的 `properties` 字段读取 `orgName` 和 `projectSlug`（前端 `CopilotProvider` 已注入），动态构建 GraphQL endpoint。JWT 从 HTTP request header 读取，不从 properties 传递：

```python
# agent.py 启动时从 FastAPI request 中提取 token，注入到 graphql_client
authorization = request.headers.get("authorization", "")

# graphql_client.py 每次调用 Gateway
url = f"{config.GATEWAY_URL}/graphql/org/{org_name}/project/{project_slug}/"
headers = {"Authorization": authorization}   # 原样透传，Gateway 负责验签
```

---

## 7. 部署

`docker-compose.yml` 新增 `modelcraft-agent` service，同网络内可访问 `backend`。

---

## 8. 范围（v1）

**包含：**
- Python Agent 服务（FastAPI + CopilotKit SDK）
- `query_model` Tool
- `nl2filter` Tool
- 模型接入层（DeepSeek，兼容 OpenAI SDK）
- Next.js BFF `/api/copilotkit` route
- Frontend Actions（`set_filter` + `clear_filter`）
- Docker 部署

**不包含（v2 考虑）：**
- 对话历史持久化
- 多轮对话上下文记忆
- 流式 SSE（先用同步响应验证流程）
- 权限校验（由 Gateway 处理）

---

## 9. 依赖关系

本服务依赖 Filter Panel 已完成实现（`EndUserRecordWorkspace` 含 `whereJsonCommitted` 状态和 `handleClearFilter`）。  
参见：`docs/superpowers/plans/2026-05-14-filter-panel.md`

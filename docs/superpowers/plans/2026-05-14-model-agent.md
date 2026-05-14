# model-agent 服务实现计划

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** 新建独立 Python Agent 服务（`modelcraft-agent/`），通过 CopilotKit + LangGraph 协议为前端提供两个能力：`query_model`（调 Gateway GraphQL 查数据）和 `nl2filter`（自然语言 → filter JSON），同时注册前端 Actions 让 Agent 直接操控 FilterPanel 状态。

**Architecture:** FastAPI + CopilotKit Python SDK + LangGraph，Agent 名称 `modelcraft_agent`。前端通过 `/api/copilotkit` BFF route 代理请求，JWT 全程 header 透传，经 Gateway(:8090) 验签后注入 `X-User-ID` 再转发 Backend(:8080)。LLM 通过环境变量切换（当前用 DeepSeek，兼容 OpenAI SDK）。

**Tech Stack:** Python 3.11, FastAPI, copilotkit==0.1.89, langgraph, openai SDK（兼容 DeepSeek base_url），httpx（异步 HTTP 客户端），Next.js BFF route（TypeScript）

---

## 文件结构

| 操作 | 路径 | 职责 |
|------|------|------|
| **新建目录** | `modelcraft-agent/` | Python Agent 服务根目录 |
| **新建** | `modelcraft-agent/config.py` | 环境变量加载（LLM、Gateway 配置） |
| **新建** | `modelcraft-agent/llm/client.py` | 模型接入层，通过 env 切换 provider |
| **新建** | `modelcraft-agent/client/graphql_client.py` | Gateway GraphQL HTTP client，透传 Authorization |
| **新建** | `modelcraft-agent/agent.py` | LangGraph Agent 定义、AgentState、工具节点 |
| **新建** | `modelcraft-agent/main.py` | FastAPI 入口，注册 CopilotKit endpoint |
| **新建** | `modelcraft-agent/requirements.txt` | Python 依赖 |
| **新建** | `modelcraft-agent/Dockerfile` | 多阶段构建镜像 |
| **新建** | `modelcraft-agent/.env.example` | 环境变量示例 |
| **新建** | `modelcraft-front/src/app/api/copilotkit/route.ts` | BFF 代理 route，透传 Authorization + Cookie |
| **修改** | `modelcraft-front/src/web/components/features/end-user-data/EndUserRecordWorkspace.tsx` | 注册 useCopilotAction (set_filter, clear_filter) |
| **修改** | `docker-compose.yml` | 新增 modelcraft-agent service |
| **修改** | `.env.example`（根目录） | 新增 AGENT_SERVICE_URL、LLM_* 环境变量 |

---

## Task 1: Python 服务骨架（config + requirements + 目录）

**Files:**
- Create: `modelcraft-agent/config.py`
- Create: `modelcraft-agent/requirements.txt`
- Create: `modelcraft-agent/.env.example`
- Create: `modelcraft-agent/llm/__init__.py`
- Create: `modelcraft-agent/client/__init__.py`

- [ ] **Step 1: 新建目录结构**

```bash
mkdir -p modelcraft-agent/llm modelcraft-agent/client
touch modelcraft-agent/llm/__init__.py modelcraft-agent/client/__init__.py
```

- [ ] **Step 2: 创建 requirements.txt**

新建 `modelcraft-agent/requirements.txt`：

```
fastapi==0.115.12
uvicorn[standard]==0.34.2
copilotkit==0.1.89
langgraph==0.4.3
openai==1.77.0
httpx==0.28.1
python-dotenv==1.1.0
```

- [ ] **Step 3: 创建 config.py**

新建 `modelcraft-agent/config.py`：

```python
import os
from dotenv import load_dotenv

load_dotenv()

# LLM provider settings
LLM_PROVIDER: str = os.environ.get("LLM_PROVIDER", "deepseek")
LLM_MODEL: str = os.environ.get("LLM_MODEL", "deepseek-chat")
LLM_API_KEY: str = os.environ.get("LLM_API_KEY", "")
LLM_BASE_URL: str = os.environ.get("LLM_BASE_URL", "https://api.deepseek.com")

# Gateway settings — all GraphQL calls MUST go through gateway, never direct to backend
GATEWAY_URL: str = os.environ.get("GATEWAY_URL", "http://localhost:8090")

# Server settings
PORT: int = int(os.environ.get("PORT", "8000"))
```

- [ ] **Step 4: 创建 .env.example**

新建 `modelcraft-agent/.env.example`：

```bash
# LLM Provider (deepseek | openai | anthropic)
LLM_PROVIDER=deepseek
LLM_MODEL=deepseek-chat
LLM_API_KEY=sk-your-key-here
LLM_BASE_URL=https://api.deepseek.com

# Gateway URL — all GraphQL calls must go through gateway
GATEWAY_URL=http://gateway:8090

# Server
PORT=8000
```

- [ ] **Step 5: Commit**

```bash
git add modelcraft-agent/
git commit -m "feat(agent): bootstrap Python service skeleton with config"
```

---

## Task 2: 模型接入层（llm/client.py）

**Files:**
- Create: `modelcraft-agent/llm/client.py`

- [ ] **Step 1: 创建 llm/client.py**

新建 `modelcraft-agent/llm/client.py`：

```python
"""
LLM client factory.

Supports deepseek/openai (OpenAI-compatible SDK) and anthropic.
Switch provider via LLM_PROVIDER env var — agent code never imports provider SDK directly.
"""
from openai import AsyncOpenAI
import config


def get_llm_client() -> AsyncOpenAI:
    """
    Return an async OpenAI-compatible client.

    DeepSeek and OpenAI both use the OpenAI SDK; only base_url differs.
    For Anthropic, swap this to the anthropic SDK in a future iteration.
    """
    if config.LLM_PROVIDER in ("deepseek", "openai"):
        kwargs = {"api_key": config.LLM_API_KEY}
        if config.LLM_BASE_URL:
            kwargs["base_url"] = config.LLM_BASE_URL
        return AsyncOpenAI(**kwargs)

    raise ValueError(
        f"Unsupported LLM_PROVIDER='{config.LLM_PROVIDER}'. "
        "Supported values: deepseek, openai"
    )


def get_model_name() -> str:
    return config.LLM_MODEL
```

- [ ] **Step 2: 手动验证（无自动测试，逻辑太薄）**

```bash
cd modelcraft-agent
python3 -c "
import os; os.environ['LLM_PROVIDER']='deepseek'; os.environ['LLM_API_KEY']='test'; os.environ['LLM_BASE_URL']='https://api.deepseek.com'
from llm.client import get_llm_client, get_model_name
c = get_llm_client()
print('client type:', type(c).__name__)  # Should print: AsyncOpenAI
print('model:', get_model_name())
"
```

期望输出：
```
client type: AsyncOpenAI
model: deepseek-chat
```

- [ ] **Step 3: Commit**

```bash
git add modelcraft-agent/llm/client.py
git commit -m "feat(agent): add LLM client factory supporting deepseek/openai"
```

---

## Task 3: Gateway GraphQL client（client/graphql_client.py）

**Files:**
- Create: `modelcraft-agent/client/graphql_client.py`

Runtime GraphQL 路由：`/graphql/org/{orgName}/project/{projectSlug}/db/{dbName}/model/{modelName}`

findMany 查询格式：
```graphql
query { findMany(where: {...}, take: N, skip: 0) { items { field1 field2 } totalCount } }
```

- [ ] **Step 1: 创建 graphql_client.py**

新建 `modelcraft-agent/client/graphql_client.py`：

```python
"""
Gateway GraphQL client.

All requests go through Gateway(:8090), never directly to backend(:8080).
Gateway validates JWT, injects X-User-ID, preserves Authorization header,
then forwards to Backend.
"""
import json
from typing import Any

import httpx

import config


class GraphQLClient:
    """Async GraphQL client that forwards Authorization header to Gateway."""

    def __init__(self, authorization: str):
        """
        Args:
            authorization: The full 'Authorization: Bearer <token>' value
                           received from the incoming request. Forwarded as-is.
        """
        self._authorization = authorization

    def _build_url(self, org_name: str, project_slug: str, db_name: str, model_name: str) -> str:
        return (
            f"{config.GATEWAY_URL}/graphql/org/{org_name}"
            f"/project/{project_slug}/db/{db_name}/model/{model_name}"
        )

    async def find_many(
        self,
        org_name: str,
        project_slug: str,
        db_name: str,
        model_name: str,
        fields: list[str],
        where: dict | None = None,
        take: int = 20,
        skip: int = 0,
    ) -> dict[str, Any]:
        """
        Execute a findMany query on the runtime GraphQL endpoint.

        Returns the full GraphQL response dict, e.g.:
        {"data": {"findMany": {"items": [...], "totalCount": N}}, "errors": [...]}
        """
        fields_str = " ".join(fields) if fields else "id"
        where_arg = f", where: {json.dumps(where)}" if where else ""
        query = (
            f"{{ findMany(take: {take}, skip: {skip}{where_arg}) "
            f"{{ items {{ {fields_str} }} totalCount }} }}"
        )

        url = self._build_url(org_name, project_slug, db_name, model_name)
        headers = {
            "Content-Type": "application/json",
            "Authorization": self._authorization,
        }
        payload = {"query": query}

        async with httpx.AsyncClient(timeout=30.0) as client:
            response = await client.post(url, headers=headers, json=payload)
            response.raise_for_status()
            return response.json()
```

- [ ] **Step 2: Commit**

```bash
git add modelcraft-agent/client/graphql_client.py
git commit -m "feat(agent): add Gateway GraphQL client with Authorization header forwarding"
```

---

## Task 4: LangGraph Agent（agent.py）

**Files:**
- Create: `modelcraft-agent/agent.py`

Agent 包含两个工具节点：
1. `query_model_tool` — 调 Gateway GraphQL 查询数据
2. `nl2filter_tool` — 调 LLM 把自然语言翻译成 ModelCraft filter JSON

AgentState 携带 `authorization` 字段（从 HTTP header 提取，注入给工具节点）。

- [ ] **Step 1: 创建 agent.py**

新建 `modelcraft-agent/agent.py`：

```python
"""
ModelCraft LangGraph Agent.

Agent state carries `authorization` (the Bearer token from the incoming HTTP request).
Both tools forward this token to Gateway when making downstream calls.
"""
import json
from typing import Annotated, Any

from langchain_openai import ChatOpenAI
from langgraph.graph import StateGraph, END
from langgraph.graph.message import add_messages
from langgraph.prebuilt import ToolNode
from langchain_core.messages import HumanMessage, AIMessage
from langchain_core.tools import tool
from typing_extensions import TypedDict

import config
from client.graphql_client import GraphQLClient

# ---------------------------------------------------------------------------
# Agent State
# ---------------------------------------------------------------------------

class AgentState(TypedDict):
    messages: Annotated[list, add_messages]
    # JWT from the incoming HTTP request — forwarded to Gateway on every tool call.
    authorization: str
    # Runtime context injected by CopilotKit from CopilotProvider properties.
    org_name: str
    project_slug: str


# ---------------------------------------------------------------------------
# Tools
# ---------------------------------------------------------------------------

def make_tools(state_getter):
    """
    Build tool functions bound to the current agent state.
    state_getter() returns the current AgentState dict.
    """

    @tool
    async def query_model(
        db_name: str,
        model_name: str,
        fields: list[str],
        where: dict | None = None,
        take: int = 20,
    ) -> str:
        """
        Query records from a ModelCraft data model.

        Args:
            db_name: Database name, e.g. "maindb"
            model_name: Model name, e.g. "users"
            fields: List of field names to return, e.g. ["id", "name", "createdAt"]
            where: Optional ModelCraft filter JSON, e.g. {"name": {"contains": "张"}}
            take: Max records to return (default 20)

        Returns:
            JSON string with items array and totalCount.
        """
        state = state_getter()
        client = GraphQLClient(authorization=state["authorization"])
        result = await client.find_many(
            org_name=state["org_name"],
            project_slug=state["project_slug"],
            db_name=db_name,
            model_name=model_name,
            fields=fields,
            where=where,
            take=take,
        )
        if "errors" in result and result["errors"]:
            return f"GraphQL error: {result['errors']}"
        data = result.get("data", {}).get("findMany", {})
        return json.dumps(data, ensure_ascii=False)

    @tool
    async def nl2filter(
        natural_language: str,
        field_names: list[str],
    ) -> str:
        """
        Convert a natural language filter description into a ModelCraft where JSON.

        Args:
            natural_language: User's filter intent, e.g. "名字包含张的"
            field_names: Available field names in the model, e.g. ["id", "name", "age"]

        Returns:
            A JSON string representing the ModelCraft where clause,
            e.g. {"AND": [{"name": {"contains": "张"}}]}
        """
        llm = ChatOpenAI(
            model=config.LLM_MODEL,
            api_key=config.LLM_API_KEY,
            base_url=config.LLM_BASE_URL if config.LLM_BASE_URL else None,
            temperature=0,
        )
        system_prompt = f"""You are a filter JSON generator for ModelCraft.
Convert the user's natural language filter description into a valid ModelCraft where JSON.

Available fields: {field_names}

ModelCraft where JSON rules:
- Top level: {{"AND": [...]}}, {{"OR": [...]}}, or a single field condition
- String operators: contains, startsWith, endsWith, equals, not
- Number operators: equals, not, gt, gte, lt, lte
- Combined: {{"AND": [{{"name": {{"contains": "张"}}}}, {{"age": {{"gte": 18}}}}]}}

Return ONLY the raw JSON object, no explanation, no markdown code fences."""

        response = await llm.ainvoke([
            {"role": "system", "content": system_prompt},
            {"role": "user", "content": natural_language},
        ])
        raw = response.content.strip()
        # Validate it parses as JSON
        json.loads(raw)
        return raw

    return [query_model, nl2filter]


# ---------------------------------------------------------------------------
# Graph
# ---------------------------------------------------------------------------

def build_graph():
    """Build the LangGraph StateGraph for modelcraft_agent."""

    # We use a closure to share state between the agent node and tools
    _current_state: dict[str, Any] = {}

    def get_state():
        return _current_state

    tools = make_tools(get_state)
    tool_node = ToolNode(tools)

    llm = ChatOpenAI(
        model=config.LLM_MODEL,
        api_key=config.LLM_API_KEY,
        base_url=config.LLM_BASE_URL if config.LLM_BASE_URL else None,
        temperature=0,
    ).bind_tools(tools)

    async def agent_node(state: AgentState):
        # Update shared state so tools can access authorization + context
        _current_state.update({
            "authorization": state.get("authorization", ""),
            "org_name": state.get("org_name", ""),
            "project_slug": state.get("project_slug", ""),
        })

        system_msg = {
            "role": "system",
            "content": (
                "你是 ModelCraft AI 助手。你可以帮用户查询数据（query_model）"
                "或将自然语言筛选条件转换为 filter JSON（nl2filter）。"
                f"当前项目：{state.get('org_name', '')}/{state.get('project_slug', '')}。"
                "如果用户说"筛选"、"过滤"类需求，先用 nl2filter 生成 filter JSON，"
                "然后告知用户 filter 已生成，前端会自动应用。"
            ),
        }
        messages = [system_msg] + state["messages"]
        response = await llm.ainvoke(messages)
        return {"messages": [response]}

    def should_continue(state: AgentState):
        last = state["messages"][-1]
        if hasattr(last, "tool_calls") and last.tool_calls:
            return "tools"
        return END

    graph = StateGraph(AgentState)
    graph.add_node("agent", agent_node)
    graph.add_node("tools", tool_node)
    graph.set_entry_point("agent")
    graph.add_conditional_edges("agent", should_continue, {"tools": "tools", END: END})
    graph.add_edge("tools", "agent")

    return graph.compile()


# Module-level compiled graph (reused across requests)
modelcraft_graph = build_graph()
```

- [ ] **Step 2: Commit**

```bash
git add modelcraft-agent/agent.py
git commit -m "feat(agent): add LangGraph agent with query_model and nl2filter tools"
```

---

## Task 5: FastAPI 入口（main.py）

**Files:**
- Create: `modelcraft-agent/main.py`

CopilotKit Python SDK 通过 `add_fastapi_endpoint` 挂载标准 `/copilotkit` POST endpoint，接受 CopilotKit 前端 SDK 的流式请求。

- [ ] **Step 1: 创建 main.py**

新建 `modelcraft-agent/main.py`：

```python
"""
FastAPI entry point for modelcraft-agent.

Exposes POST /copilotkit — the CopilotKit runtime endpoint consumed by the
Next.js BFF at /api/copilotkit.
"""
import uvicorn
from fastapi import FastAPI, Request
from copilotkit import CopilotKitRemoteEndpoint, LangGraphAgent
from copilotkit.integrations.fastapi import add_fastapi_endpoint

import config
from agent import modelcraft_graph

app = FastAPI(title="modelcraft-agent", version="0.1.0")


# ---------------------------------------------------------------------------
# Health check
# ---------------------------------------------------------------------------

@app.get("/healthz")
async def healthz():
    return {"status": "ok", "service": "modelcraft-agent"}


# ---------------------------------------------------------------------------
# CopilotKit endpoint
# ---------------------------------------------------------------------------

sdk = CopilotKitRemoteEndpoint(
    agents=[
        LangGraphAgent(
            name="modelcraft_agent",
            description="ModelCraft AI 助手：数据查询 + 自然语言筛选",
            graph=modelcraft_graph,
            initial_state=lambda req, input_data: {
                "authorization": req.headers.get("authorization", ""),
                "org_name": (input_data.get("properties") or {}).get("orgName", ""),
                "project_slug": (input_data.get("properties") or {}).get("projectSlug", ""),
            },
        )
    ]
)

add_fastapi_endpoint(app, sdk, "/copilotkit")


# ---------------------------------------------------------------------------
# Dev server
# ---------------------------------------------------------------------------

if __name__ == "__main__":
    uvicorn.run("main:app", host="0.0.0.0", port=config.PORT, reload=True)
```

- [ ] **Step 2: 本地启动验证**

```bash
cd modelcraft-agent
pip install -r requirements.txt
LLM_PROVIDER=deepseek LLM_API_KEY=test LLM_MODEL=deepseek-chat \
  LLM_BASE_URL=https://api.deepseek.com GATEWAY_URL=http://localhost:8090 \
  python main.py
```

另开终端验证健康检查：
```bash
curl http://localhost:8000/healthz
```

期望：
```json
{"status": "ok", "service": "modelcraft-agent"}
```

- [ ] **Step 3: Commit**

```bash
git add modelcraft-agent/main.py
git commit -m "feat(agent): add FastAPI entry with CopilotKit endpoint"
```

---

## Task 6: Dockerfile

**Files:**
- Create: `modelcraft-agent/Dockerfile`

- [ ] **Step 1: 创建 Dockerfile**

新建 `modelcraft-agent/Dockerfile`：

```dockerfile
FROM python:3.11-slim AS builder

WORKDIR /build
COPY requirements.txt .
RUN pip install --no-cache-dir --user -r requirements.txt

# ---------------------------------------------------------------------------
FROM python:3.11-slim

WORKDIR /app

# Copy installed packages from builder
COPY --from=builder /root/.local /root/.local
ENV PATH=/root/.local/bin:$PATH

COPY . .

EXPOSE 8000

HEALTHCHECK --interval=30s --timeout=5s --start-period=10s --retries=3 \
  CMD python -c "import urllib.request; urllib.request.urlopen('http://localhost:8000/healthz')"

CMD ["python", "-m", "uvicorn", "main:app", "--host", "0.0.0.0", "--port", "8000"]
```

- [ ] **Step 2: Commit**

```bash
git add modelcraft-agent/Dockerfile
git commit -m "feat(agent): add Dockerfile for modelcraft-agent service"
```

---

## Task 7: docker-compose 接入

**Files:**
- Modify: `docker-compose.yml`
- Modify: `.env.example`（根目录）

- [ ] **Step 1: 在 docker-compose.yml 的 gateway service 之后新增 agent service**

在 `docker-compose.yml` 中，`frontend:` service 定义之前，新增如下内容（保持 2 空格缩进与文件一致）：

```yaml
  # ---------------------------------------------------------------------------
  # modelcraft-agent (Python AI Agent)
  # ---------------------------------------------------------------------------
  modelcraft-agent:
    build:
      context: ./modelcraft-agent
      dockerfile: Dockerfile
    image: modelcraft-agent:${IMAGE_TAG:-latest}
    container_name: modelcraft-agent
    ports:
      - "${AGENT_PORT:-8000}:8000"
    environment:
      - LLM_PROVIDER=${LLM_PROVIDER:-deepseek}
      - LLM_MODEL=${LLM_MODEL:-deepseek-chat}
      - LLM_API_KEY=${LLM_API_KEY}
      - LLM_BASE_URL=${LLM_BASE_URL:-https://api.deepseek.com}
      - GATEWAY_URL=http://gateway:8090
      - PORT=8000
      - TZ=Asia/Shanghai
    depends_on:
      gateway:
        condition: service_healthy
    networks:
      - modelcraft-network
    restart: unless-stopped
    healthcheck:
      test: ["CMD", "python", "-c", "import urllib.request; urllib.request.urlopen('http://localhost:8000/healthz')"]
      interval: 30s
      timeout: 5s
      retries: 3
      start_period: 15s
```

同时将 `frontend` service 的 `depends_on` 中加入 agent 依赖（可选，但确保顺序）：

```yaml
  frontend:
    ...
    environment:
      - NODE_ENV=production
      - TZ=Asia/Shanghai
      - AGENT_SERVICE_URL=http://modelcraft-agent:8000   # ← 新增这行
```

- [ ] **Step 2: 更新根目录 .env.example**

在根目录 `.env.example` 尾部追加：

```bash
# modelcraft-agent
LLM_PROVIDER=deepseek
LLM_MODEL=deepseek-chat
LLM_API_KEY=sk-your-deepseek-key
LLM_BASE_URL=https://api.deepseek.com
AGENT_PORT=8000
```

- [ ] **Step 3: Commit**

```bash
git add docker-compose.yml .env.example
git commit -m "feat(deploy): add modelcraft-agent service to docker-compose"
```

---

## Task 8: Next.js BFF Route（/api/copilotkit）

**Files:**
- Create: `modelcraft-front/src/app/api/copilotkit/route.ts`

与现有 BFF route 完全相同的透传模式（参见 `src/app/api/bff/graphql/org/[orgName]/project/[projectSlug]/route.ts`）：原样转发 `Authorization` header 和 `Cookie`，代理到 `AGENT_SERVICE_URL`。

- [ ] **Step 1: 创建 route.ts**

新建 `modelcraft-front/src/app/api/copilotkit/route.ts`：

```typescript
/**
 * BFF proxy for CopilotKit runtime.
 *
 * Forwards the request to the Python modelcraft-agent service.
 * Authorization header and Cookie are forwarded as-is so the Python agent
 * can carry the JWT through to Gateway on every tool call.
 *
 * The CopilotProvider is already configured with:
 *   runtimeUrl="/api/copilotkit"
 *   agent="modelcraft_agent"
 */
import { NextRequest, NextResponse } from 'next/server'

const AGENT_SERVICE_URL = process.env.AGENT_SERVICE_URL ?? 'http://localhost:8000'

async function handler(req: NextRequest): Promise<NextResponse> {
  const upstreamUrl = `${AGENT_SERVICE_URL}/copilotkit`

  const headers = new Headers()
  headers.set('Content-Type', req.headers.get('Content-Type') ?? 'application/json')

  // Forward Authorization so Python agent can carry JWT to Gateway
  const authHeader = req.headers.get('Authorization')
  if (authHeader) headers.set('Authorization', authHeader)

  // Forward Cookie for refresh token support
  const cookieHeader = req.headers.get('cookie')
  if (cookieHeader) headers.set('Cookie', cookieHeader)

  const body = req.method !== 'GET' && req.method !== 'HEAD' ? await req.text() : undefined

  let upstreamRes: Response
  try {
    upstreamRes = await fetch(upstreamUrl, { method: req.method, headers, body })
  } catch {
    return NextResponse.json(
      { error: 'Agent service unreachable' },
      { status: 502 }
    )
  }

  const resBody = await upstreamRes.arrayBuffer()
  const response = new NextResponse(resBody, {
    status: upstreamRes.status,
    statusText: upstreamRes.statusText,
  })

  upstreamRes.headers.forEach((value, key) => {
    if (
      ['content-encoding', 'content-length', 'transfer-encoding', 'connection'].includes(
        key.toLowerCase()
      )
    )
      return
    response.headers.append(key, value)
  })

  return response
}

export const GET = handler
export const POST = handler
```

- [ ] **Step 2: 确认 AGENT_SERVICE_URL 在 Next.js 构建时注入**

检查 `modelcraft-front/next.config.ts`（或 `next.config.js`）是否已暴露服务端环境变量，`AGENT_SERVICE_URL` 是服务端变量（无需 `NEXT_PUBLIC_` 前缀），Next.js 服务端 route handler 默认可读 `process.env`，无需额外配置。

运行验证：
```bash
cd modelcraft-front
grep -r "AGENT_SERVICE_URL" src/ || echo "variable not yet referenced in src (expected)"
```

- [ ] **Step 3: Lint 检查**

```bash
cd modelcraft-front
npx eslint src/app/api/copilotkit/route.ts
```

期望：0 errors

- [ ] **Step 4: Commit**

```bash
git add modelcraft-front/src/app/api/copilotkit/route.ts
git commit -m "feat(front): add /api/copilotkit BFF proxy to Python agent service"
```

---

## Task 9: 前端 Frontend Actions（useCopilotAction）

**Files:**
- Modify: `modelcraft-front/src/web/components/features/end-user-data/EndUserRecordWorkspace.tsx`

在 `EndUserRecordWorkspace` 中注册两个 Frontend Actions，让 Agent 可以直接设置/清空 FilterPanel 状态。这些 hook 必须在 `CopilotProvider` 包裹下调用（`EndUserRecordWorkspace` 已在 project layout 的 `CopilotWrapper` 内）。

- [ ] **Step 1: 读取 EndUserRecordWorkspace.tsx 确认 import 区和 state 位置**

```bash
head -50 modelcraft-front/src/web/components/features/end-user-data/EndUserRecordWorkspace.tsx
```

确认文件顶部已有 `import React` 和其他 import，找到 `setWhereJsonDraft`、`setWhereJsonCommitted`、`handleClearFilter` 的定义行号（来自 filter-panel.md 计划，Task 5 的状态定义）。

- [ ] **Step 2: 在 import 区末尾新增 useCopilotAction import**

在 `EndUserRecordWorkspace.tsx` 的 import 区（在现有最后一个 import 下方）新增：

```typescript
import { useCopilotAction } from '@copilotkit/react-core'
```

- [ ] **Step 3: 在 component 函数体内、return 语句之前注册 Actions**

在 `handleClearFilter` 定义之后、`return (` 之前，新增：

```typescript
// ----- CopilotKit Frontend Actions -----
// These allow the AI agent to directly manipulate the FilterPanel state.
// Must be called inside a component that is wrapped by CopilotProvider.

useCopilotAction({
  name: 'set_filter',
  description:
    '设置 FilterPanel 的 where 筛选条件。接受 ModelCraft filter JSON 字符串，例如: {"AND":[{"name":{"contains":"张"}}]}',
  parameters: [
    {
      name: 'filter_json',
      type: 'string',
      description: 'ModelCraft where JSON 字符串',
      required: true,
    },
  ],
  handler: async ({ filter_json }: { filter_json: string }) => {
    setWhereJsonDraft(filter_json)
    setWhereJsonCommitted(filter_json)
  },
})

useCopilotAction({
  name: 'clear_filter',
  description: '清空 FilterPanel 的所有筛选条件，恢复全量数据展示',
  parameters: [],
  handler: async () => {
    handleClearFilter()
  },
})
// ----- End CopilotKit Frontend Actions -----
```

- [ ] **Step 4: Lint 检查**

```bash
cd modelcraft-front
npx eslint src/web/components/features/end-user-data/EndUserRecordWorkspace.tsx
```

期望：0 errors

- [ ] **Step 5: Commit**

```bash
git add modelcraft-front/src/web/components/features/end-user-data/EndUserRecordWorkspace.tsx
git commit -m "feat(front): register set_filter and clear_filter CopilotKit actions"
```

---

## Task 10: 端对端验收测试

**验收前提：** Filter Panel（`filter-panel.md` 计划）必须已完成，`EndUserRecordWorkspace` 中已有 `whereJsonDraft`、`whereJsonCommitted`、`handleClearFilter`。

- [ ] **Step 1: 启动完整服务栈**

```bash
# 在根目录，确保 .env 中已设置 LLM_API_KEY
docker compose up -d gateway backend modelcraft-agent
cd modelcraft-front && npm run dev
```

- [ ] **Step 2: 验证 Agent 服务健康**

```bash
curl http://localhost:8000/healthz
```

期望：`{"status":"ok","service":"modelcraft-agent"}`

- [ ] **Step 3: 验证 BFF route 可达**

打开浏览器 DevTools → Network，登录后访问任意项目的 EndUser 数据视图，确认 `CopilotSidebar` 侧边栏出现。

在 Network tab 中应能看到 `POST /api/copilotkit` 请求，状态码 200（或 SSE stream）。

- [ ] **Step 4: 验证 NL2Filter + set_filter**

在 CopilotSidebar 输入：`筛选名字包含张的记录`

期望：
1. Agent 回复确认消息
2. FilterPanel 自动展开，where JSON 编辑器中出现类似 `{"AND":[{"name":{"contains":"张"}}]}` 的内容
3. 表格数据刷新为过滤后的结果

- [ ] **Step 5: 验证 query_model**

在 CopilotSidebar 输入：`帮我查一下最新的 5 条记录，只要 id 和名字字段`

期望：
1. Agent 回复包含数据列表（JSON 格式或表格文字）
2. Network tab 可见 Agent → Gateway 的 GraphQL 请求（在 modelcraft-agent 容器日志中可见）

- [ ] **Step 6: 验证 clear_filter**

有激活筛选时，在 CopilotSidebar 输入：`清空筛选条件`

期望：
1. FilterPanel 编辑器清空
2. 表格恢复全量数据
3. 筛选按钮角标消失

- [ ] **Step 7: Final commit（如有遗留改动）**

```bash
git status
# 若有未提交改动
git add -p
git commit -m "feat(agent): complete model-agent service end-to-end"
```

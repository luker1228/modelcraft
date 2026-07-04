# CopilotKit Two-Agent Architecture Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

> Status: historical plan only. Current code keeps `modelcraft_admin_agent` only; `modelcraft_enduser_agent` and `/copilotkit/enduser` are not active.

**Goal:** Split the single `modelcraft_agent` into two independent agents — `modelcraft_admin_agent` (tenant admin) and `modelcraft_enduser_agent` (end user) — each with dedicated tool sets, knowledge bases, and sidebar suggestions.

**Architecture:** Python side splits `agent.py` into `agents/tools.py` (shared tools) + `agents/admin_agent.py` + `agents/enduser_agent.py`, registered separately in `main.py`. Frontend side adds `SharedCopilotActions` (show_toast), `AdminCopilotKnowledge`, `EndUserCopilotKnowledge`, `EndUserCopilotActions`, and wires them into the respective `CopilotProvider`/`EndUserCopilotWrapper` trees.

**Tech Stack:** Python 3.11, LangGraph, LangChain, pytest · Next.js 14, React, @copilotkit/react-core, @copilotkit/react-ui, sonner, Vitest

---

## File Map

### Created
- `modelcraft-agent/agents/__init__.py`
- `modelcraft-agent/agents/shared.py` — AgentState TypedDict + helper functions
- `modelcraft-agent/agents/tools.py` — all `@tool` functions (shared between both agents)
- `modelcraft-agent/agents/admin_agent.py` — admin graph (list_projects + all tools)
- `modelcraft-agent/agents/enduser_agent.py` — enduser graph (query/filter tools only)
- `modelcraft-agent/tests/agents/test_admin_agent.py`
- `modelcraft-agent/tests/agents/test_enduser_agent.py`
- `modelcraft-front/src/web/components/features/copilot/SharedCopilotActions.tsx`
- `modelcraft-front/src/web/components/features/copilot/AdminCopilotKnowledge.tsx`
- `modelcraft-front/src/web/components/features/copilot/EndUserCopilotActions.tsx`
- `modelcraft-front/src/web/components/features/copilot/EndUserCopilotKnowledge.tsx`

### Modified
- `modelcraft-agent/main.py` — register both agents
- `modelcraft-agent/agent.py` — deleted after migration
- `modelcraft-front/src/web/components/features/copilot/CopilotProvider.tsx` — admin agent name, mount SharedCopilotActions + AdminCopilotKnowledge
- `modelcraft-front/src/web/components/features/copilot/ProjectCopilotActions.tsx` — add 4 new nav tools
- `modelcraft-front/src/app/org/[orgName]/layout.tsx` — change agent name
- `modelcraft-front/src/app/org/[orgName]/project/[projectSlug]/layout.tsx` — change agent name
- `modelcraft-front/src/app/end-user/[orgName]/projects/[projectSlug]/data/layout.tsx` — change agent name

---

## Task 1: Create `agents/` package — shared state and helpers

**Files:**
- Create: `modelcraft-agent/agents/__init__.py`
- Create: `modelcraft-agent/agents/shared.py`

- [ ] **Step 1: Create the package init**

```bash
mkdir -p modelcraft-agent/agents
touch modelcraft-agent/agents/__init__.py
```

- [ ] **Step 2: Write `agents/shared.py`**

```python
# modelcraft-agent/agents/shared.py
"""Shared AgentState and logging helpers for ModelCraft agents."""
import time
from typing import Annotated

from typing_extensions import TypedDict
from langgraph.graph.message import add_messages

from client.graphql_client import GraphQLClient
from logging_setup import get_logger


class AgentState(TypedDict):
    messages: Annotated[list, add_messages]
    authorization: str
    org_name: str
    project_slug: str
    layer: str          # "org" | "project" | ""
    current_model: str  # current model name from route
    current_db: str     # current database name from route


def make_client(state: AgentState) -> GraphQLClient:
    """Create a GraphQL client authenticated with the state's token."""
    return GraphQLClient(authorization=state["authorization"])


def log_tool_start(name: str, args_summary: str):
    """Start a tool log entry, return (log, start_time)."""
    log = get_logger()
    log.info("tool.call.start", tool_name=name, args_summary=args_summary[:200])
    return log, time.perf_counter()


def log_tool_end(log, name: str, start: float, success: bool) -> None:
    """Finish a tool log entry."""
    log.info(
        "tool.call.end",
        tool_name=name,
        duration_ms=round((time.perf_counter() - start) * 1000, 2),
        success=success,
    )
```

- [ ] **Step 3: Verify imports resolve**

```bash
cd modelcraft-agent && python -c "from agents.shared import AgentState, make_client, log_tool_start, log_tool_end; print('OK')"
```

Expected: `OK`

- [ ] **Step 4: Commit**

```bash
git add modelcraft-agent/agents/
git commit -m "feat(agent): create agents/ package with shared state and helpers"
```

---

## Task 2: Extract all tools to `agents/tools.py`

**Files:**
- Create: `modelcraft-agent/agents/tools.py`
- Create: `modelcraft-agent/tests/agents/__init__.py`
- Create: `modelcraft-agent/tests/agents/test_tools.py`

- [ ] **Step 1: Write the failing test**

```python
# modelcraft-agent/tests/agents/test_tools.py
"""Smoke tests for agent tool functions."""
import json
from unittest.mock import AsyncMock, MagicMock, patch
import pytest

from agents.tools import list_projects, list_models, get_model_fields, query_model, nl2filter


FAKE_STATE = {
    "messages": [],
    "authorization": "Bearer test-token",
    "org_name": "testorg",
    "project_slug": "testproject",
    "layer": "project",
    "current_model": "",
    "current_db": "",
}


@pytest.mark.asyncio
async def test_list_projects_returns_json():
    mock_result = {"data": {"projects": [{"id": "1", "slug": "demo", "title": "Demo"}]}}
    with patch("agents.tools.make_client") as mock_make_client:
        mock_client = AsyncMock()
        mock_client.list_projects.return_value = mock_result
        mock_make_client.return_value = mock_client

        result = await list_projects.ainvoke({"state": FAKE_STATE})

    data = json.loads(result)
    assert data[0]["slug"] == "demo"


@pytest.mark.asyncio
async def test_list_projects_returns_graphql_error():
    mock_result = {"errors": [{"message": "Unauthorized"}]}
    with patch("agents.tools.make_client") as mock_make_client:
        mock_client = AsyncMock()
        mock_client.list_projects.return_value = mock_result
        mock_make_client.return_value = mock_client

        result = await list_projects.ainvoke({"state": FAKE_STATE})

    assert "GraphQL error" in result


@pytest.mark.asyncio
async def test_nl2filter_returns_valid_json():
    fake_llm_response = MagicMock()
    fake_llm_response.content = '{"name": {"contains": "张"}}'
    with patch("agents.tools._get_llm") as mock_get_llm:
        mock_llm = AsyncMock()
        mock_llm.ainvoke.return_value = fake_llm_response
        mock_get_llm.return_value = mock_llm

        result = await nl2filter.ainvoke({
            "natural_language": "名字包含张",
            "field_names": ["name", "age"],
        })

    parsed = json.loads(result)
    assert "name" in parsed
```

- [ ] **Step 2: Create tests init and run to verify FAIL**

```bash
mkdir -p modelcraft-agent/tests/agents
touch modelcraft-agent/tests/agents/__init__.py
cd modelcraft-agent && python -m pytest tests/agents/test_tools.py -v 2>&1 | head -20
```

Expected: `ImportError: cannot import name 'list_projects' from 'agents.tools'`

- [ ] **Step 3: Write `agents/tools.py`**

```python
# modelcraft-agent/agents/tools.py
"""All @tool functions shared between admin and end-user agents."""
import json
from functools import lru_cache
from typing import Annotated

from langchain_core.tools import tool
from langchain_openai import ChatOpenAI
from langgraph.prebuilt import InjectedState

import config
from agents.shared import AgentState, make_client, log_tool_start, log_tool_end


@lru_cache(maxsize=1)
def _get_llm() -> ChatOpenAI:
    return ChatOpenAI(
        model=config.LLM_MODEL,
        api_key=config.LLM_API_KEY,
        base_url=config.LLM_BASE_URL if config.LLM_BASE_URL else None,
        temperature=0,
    )


@tool
async def list_projects(
    state: Annotated[AgentState, InjectedState()],
) -> str:
    """
    List all projects in the current organization.

    Returns:
        JSON array of projects with id, slug, title, description, status.
    """
    log, start = log_tool_start("list_projects", f"org={state['org_name']}")
    try:
        result = await make_client(state).list_projects(org_name=state["org_name"])
        if "errors" in result and result["errors"]:
            return f"GraphQL error: {result['errors']}"
        projects = result.get("data", {}).get("projects", [])
        log_tool_end(log, "list_projects", start, True)
        return json.dumps(projects, ensure_ascii=False)
    except Exception:
        log.exception("error", tool_name="list_projects")
        log_tool_end(log, "list_projects", start, False)
        raise


@tool
async def list_models(
    database_name: str,
    state: Annotated[AgentState, InjectedState()],
) -> str:
    """
    List all data models in a database of the current project.

    Args:
        database_name: Database name, e.g. "maindb"

    Returns:
        JSON array of models with id, name, title, description, databaseName, displayField.
    """
    log, start = log_tool_start("list_models", f"db={database_name}")
    try:
        result = await make_client(state).list_models(
            org_name=state["org_name"],
            project_slug=state["project_slug"],
            database_name=database_name,
        )
        if "errors" in result and result["errors"]:
            return f"GraphQL error: {result['errors']}"
        data = result.get("data", {}).get("models", {})
        log_tool_end(log, "list_models", start, True)
        return json.dumps(data, ensure_ascii=False)
    except Exception:
        log.exception("error", tool_name="list_models")
        log_tool_end(log, "list_models", start, False)
        raise


@tool
async def get_model_fields(
    model_id: str,
    state: Annotated[AgentState, InjectedState()],
) -> str:
    """
    Get the field definitions of a model by its ID.
    Use this before query_model to know what fields are available.

    Args:
        model_id: Model ID (from list_models)

    Returns:
        JSON array of fields with name, title, schemaType, format, isPrimary, isUnique, etc.
    """
    log, start = log_tool_start("get_model_fields", f"model_id={model_id}")
    try:
        result = await make_client(state).get_model_fields(
            org_name=state["org_name"],
            project_slug=state["project_slug"],
            model_id=model_id,
        )
        if "errors" in result and result["errors"]:
            return f"GraphQL error: {result['errors']}"
        fields = result.get("data", {}).get("fields", [])
        log_tool_end(log, "get_model_fields", start, True)
        return json.dumps(fields, ensure_ascii=False)
    except Exception:
        log.exception("error", tool_name="get_model_fields")
        log_tool_end(log, "get_model_fields", start, False)
        raise


@tool
async def query_model(
    db_name: str,
    model_name: str,
    fields: list[str],
    take: int,
    state: Annotated[AgentState, InjectedState()],
    where: dict | None = None,
    skip: int = 0,
) -> str:
    """
    Query records from a ModelCraft data model.

    Args:
        db_name: Database name, e.g. "maindb"
        model_name: Model name, e.g. "users"
        fields: Field names to return, e.g. ["id", "name", "createdAt"]
        take: Max records to return (1-200)
        where: Optional filter JSON, e.g. {"name": {"contains": "张"}}
        skip: Records to skip for pagination (default 0)

    Returns:
        JSON with items array and totalCount.
    """
    log, start = log_tool_start("query_model", str({"db": db_name, "model": model_name, "take": take}))
    try:
        take = max(1, min(take, 200))
        result = await make_client(state).find_many(
            org_name=state["org_name"],
            project_slug=state["project_slug"],
            db_name=db_name,
            model_name=model_name,
            fields=fields,
            where=where,
            take=take,
            skip=skip,
        )
        if "errors" in result and result["errors"]:
            return f"GraphQL error: {result['errors']}"
        data = result.get("data", {}).get("findMany", {})
        log_tool_end(log, "query_model", start, True)
        return json.dumps(data, ensure_ascii=False)
    except Exception:
        log.exception("error", tool_name="query_model")
        log_tool_end(log, "query_model", start, False)
        raise


@tool
async def nl2filter(
    natural_language: str,
    field_names: list[str],
) -> str:
    """
    Convert a natural language filter description into a ModelCraft where JSON.

    Args:
        natural_language: User's filter intent, e.g. "名字包含张的且年龄大于18"
        field_names: Available field names in the model, e.g. ["id", "name", "age"]

    Returns:
        A JSON string representing the ModelCraft where clause,
        e.g. {"AND": [{"name": {"contains": "张"}}, {"age": {"gt": 18}}]}
    """
    log, start = log_tool_start("nl2filter", natural_language[:100])
    try:
        llm = _get_llm()
        system_prompt = f"""You are a filter JSON generator for ModelCraft.
Convert the user's natural language filter description into a valid ModelCraft where JSON.

Available fields: {field_names}

ModelCraft where JSON rules:
- Top level: {{"AND": [...]}}, {{"OR": [...]}}, or a single field condition
- String operators: contains, startsWith, endsWith, equals, not
- Number operators: equals, not, gt, gte, lt, lte
- Boolean: {{"active": {{"equals": true}}}}
- Combined: {{"AND": [{{"name": {{"contains": "张"}}}}, {{"age": {{"gte": 18}}}}]}}

Return ONLY the raw JSON object, no explanation, no markdown."""

        response = await llm.ainvoke([
            {"role": "system", "content": system_prompt},
            {"role": "user", "content": natural_language},
        ])
        raw = response.content.strip()
        json.loads(raw)  # validate
        log_tool_end(log, "nl2filter", start, True)
        return raw
    except Exception:
        log.exception("error", tool_name="nl2filter")
        log_tool_end(log, "nl2filter", start, False)
        raise
```

- [ ] **Step 4: Run tests — expect PASS**

```bash
cd modelcraft-agent && python -m pytest tests/agents/test_tools.py -v
```

Expected:
```
tests/agents/test_tools.py::test_list_projects_returns_json PASSED
tests/agents/test_tools.py::test_list_projects_returns_graphql_error PASSED
tests/agents/test_tools.py::test_nl2filter_returns_valid_json PASSED
```

- [ ] **Step 5: Commit**

```bash
git add modelcraft-agent/agents/tools.py modelcraft-agent/tests/agents/
git commit -m "feat(agent): extract tools to agents/tools.py"
```

---

## Task 3: Create `agents/admin_agent.py`

**Files:**
- Create: `modelcraft-agent/agents/admin_agent.py`
- Create: `modelcraft-agent/tests/agents/test_admin_agent.py`

- [ ] **Step 1: Write the failing test**

```python
# modelcraft-agent/tests/agents/test_admin_agent.py
"""Tests for the admin agent graph."""
import pytest
from agents.admin_agent import admin_graph, ADMIN_TOOLS


def test_admin_graph_compiles():
    """Graph must build without errors."""
    assert admin_graph is not None


def test_admin_graph_has_agent_and_tools_nodes():
    """Graph must have agent and tools nodes."""
    node_names = set(admin_graph.nodes.keys())
    assert "agent" in node_names
    assert "tools" in node_names


def test_admin_tools_include_list_projects():
    """Admin agent must have access to list_projects."""
    tool_names = {t.name for t in ADMIN_TOOLS}
    assert "list_projects" in tool_names


def test_admin_tools_do_not_include_enduser_only():
    """Admin tool names must be the expected set."""
    tool_names = {t.name for t in ADMIN_TOOLS}
    assert tool_names == {
        "list_projects",
        "list_models",
        "get_model_fields",
        "query_model",
        "nl2filter",
    }
```

- [ ] **Step 2: Run to verify FAIL**

```bash
cd modelcraft-agent && python -m pytest tests/agents/test_admin_agent.py -v 2>&1 | head -10
```

Expected: `ImportError: cannot import name 'admin_graph' from 'agents.admin_agent'`

- [ ] **Step 3: Write `agents/admin_agent.py`**

```python
# modelcraft-agent/agents/admin_agent.py
"""Admin agent — serves tenant administrators."""
from typing import Any

from langchain_openai import ChatOpenAI
from langgraph.checkpoint.memory import MemorySaver
from langgraph.graph import END, StateGraph
from langgraph.prebuilt import ToolNode

import config
from agents.shared import AgentState
from agents.tools import (
    get_model_fields,
    list_models,
    list_projects,
    nl2filter,
    query_model,
)

ADMIN_TOOLS = [
    list_projects,
    list_models,
    get_model_fields,
    query_model,
    nl2filter,
]

_ADMIN_TOOL_NODE = ToolNode(ADMIN_TOOLS)


def _build_admin_graph() -> Any:
    llm = ChatOpenAI(
        model=config.LLM_MODEL,
        api_key=config.LLM_API_KEY,
        base_url=config.LLM_BASE_URL if config.LLM_BASE_URL else None,
        temperature=0,
    ).bind_tools(ADMIN_TOOLS)

    async def agent_node(state: AgentState):
        org = state.get("org_name", "")
        project = state.get("project_slug", "")
        layer = state.get("layer", "")
        current_model = state.get("current_model", "")
        current_db = state.get("current_db", "")

        if layer == "org":
            context = (
                f"当前在 Org 页面（组织：{org}）。\n"
                "可用工具：navigate_to_project、navigate_to_settings、open_create_project、"
                "highlight_project、list_projects、nl2filter、show_toast。\n"
                "注意：不可直接调用 list_models、query_model 等 project 级工具。\n"
                "如需操作项目数据，先调用 navigate_to_project(slug) 跳转到对应项目。"
            )
        elif layer == "project":
            model_ctx = f"，当前模型：{current_model}（数据库：{current_db}）" if current_model else ""
            context = (
                f"当前在 Project 页面（组织：{org}，项目：{project}{model_ctx}）。\n"
                "可用工具：navigate_to_org、navigate_to_model、navigate_to_data、"
                "open_create_model、open_create_record、open_edit_record、highlight_records、"
                "navigate_to_enums、navigate_to_cluster、navigate_to_rbac、navigate_to_end_users、"
                "set_filter、clear_filter、list_models、get_model_fields、query_model、"
                "nl2filter、show_toast。\n"
                "写操作规则：open_create_record 和 open_edit_record 只预填表单，用户点 Save 才真正保存。\n"
                "操作前先用 get_model_fields 确认字段名，避免预填错误字段。"
            )
        else:
            context = (
                f"当前组织：{org}{'，项目：' + project if project else ''}。\n"
                "可用工具取决于当前页面，请先询问用户当前在哪个页面。"
            )

        system_msg = {
            "role": "system",
            "content": (
                "你是 ModelCraft AI 助手（管理员版），帮助租户管理员通过对话完成所有操作。\n\n"
                f"{context}\n\n"
                "通用原则：\n"
                "- 操作数据前先用 list_models 和 get_model_fields 确认模型和字段存在\n"
                "- 删除操作禁止自动执行，必须引导用户在界面手动确认\n"
                "- 如果用户说筛选或过滤，先用 nl2filter 生成 filter JSON，再告知前端已应用\n"
                "- 完成操作后用 show_toast 通知用户结果"
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
    graph.add_node("tools", _ADMIN_TOOL_NODE)
    graph.set_entry_point("agent")
    graph.add_conditional_edges("agent", should_continue, {"tools": "tools", END: END})
    graph.add_edge("tools", "agent")
    return graph.compile(checkpointer=MemorySaver())


class _LazyGraph:
    _instance = None

    def __getattr__(self, name):
        if self._instance is None:
            type(self)._instance = _build_admin_graph()
        return getattr(self._instance, name)


admin_graph = _LazyGraph()
```

- [ ] **Step 4: Run tests — expect PASS**

```bash
cd modelcraft-agent && python -m pytest tests/agents/test_admin_agent.py -v
```

Expected: 4 tests PASS

- [ ] **Step 5: Commit**

```bash
git add modelcraft-agent/agents/admin_agent.py modelcraft-agent/tests/agents/test_admin_agent.py
git commit -m "feat(agent): add admin_agent with list_projects + full tool set"
```

---

## Task 4: Create `agents/enduser_agent.py`

**Files:**
- Create: `modelcraft-agent/agents/enduser_agent.py`
- Create: `modelcraft-agent/tests/agents/test_enduser_agent.py`

- [ ] **Step 1: Write the failing test**

```python
# modelcraft-agent/tests/agents/test_enduser_agent.py
"""Tests for the end-user agent graph."""
import pytest
from agents.enduser_agent import enduser_graph, ENDUSER_TOOLS


def test_enduser_graph_compiles():
    assert enduser_graph is not None


def test_enduser_graph_has_agent_and_tools_nodes():
    node_names = set(enduser_graph.nodes.keys())
    assert "agent" in node_names
    assert "tools" in node_names


def test_enduser_tools_do_not_include_list_projects():
    """End-user agent must NOT have list_projects (admin-only)."""
    tool_names = {t.name for t in ENDUSER_TOOLS}
    assert "list_projects" not in tool_names


def test_enduser_tools_include_query_tools():
    tool_names = {t.name for t in ENDUSER_TOOLS}
    assert tool_names == {
        "list_models",
        "get_model_fields",
        "query_model",
        "nl2filter",
    }
```

- [ ] **Step 2: Run to verify FAIL**

```bash
cd modelcraft-agent && python -m pytest tests/agents/test_enduser_agent.py -v 2>&1 | head -10
```

Expected: `ImportError: cannot import name 'enduser_graph'`

- [ ] **Step 3: Write `agents/enduser_agent.py`**

```python
# modelcraft-agent/agents/enduser_agent.py
"""End-user agent — serves end users querying and filtering data."""
from typing import Any

from langchain_openai import ChatOpenAI
from langgraph.checkpoint.memory import MemorySaver
from langgraph.graph import END, StateGraph
from langgraph.prebuilt import ToolNode

import config
from agents.shared import AgentState
from agents.tools import (
    get_model_fields,
    list_models,
    nl2filter,
    query_model,
)

ENDUSER_TOOLS = [
    list_models,
    get_model_fields,
    query_model,
    nl2filter,
]

_ENDUSER_TOOL_NODE = ToolNode(ENDUSER_TOOLS)


def _build_enduser_graph() -> Any:
    llm = ChatOpenAI(
        model=config.LLM_MODEL,
        api_key=config.LLM_API_KEY,
        base_url=config.LLM_BASE_URL if config.LLM_BASE_URL else None,
        temperature=0,
    ).bind_tools(ENDUSER_TOOLS)

    async def agent_node(state: AgentState):
        org = state.get("org_name", "")
        project = state.get("project_slug", "")

        context = (
            f"当前在数据页面（组织：{org}，项目：{project}）。\n"
            "可用工具：navigate_to_project、navigate_to_workspace、"
            "set_filter、clear_filter、highlight_records、"
            "list_models、get_model_fields、query_model、nl2filter、show_toast。\n"
            "重要原则：只能查询数据，不能修改数据模型或结构。\n"
            "筛选流程：先用 nl2filter 将用户意图转为 filter JSON，再调用前端 set_filter 应用。"
        )

        system_msg = {
            "role": "system",
            "content": (
                "你是 ModelCraft AI 助手（用户版），帮助终端用户探索和查询数据。\n\n"
                f"{context}\n\n"
                "通用原则：\n"
                "- 优先用自然语言解释数据内容，而不是展示原始 JSON\n"
                "- 如果用户不知道有哪些数据，先调用 list_models 介绍\n"
                "- 数据量大时引导用户用自然语言缩小范围\n"
                "- 遇到权限问题时提示用户联系管理员"
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
    graph.add_node("tools", _ENDUSER_TOOL_NODE)
    graph.set_entry_point("agent")
    graph.add_conditional_edges("agent", should_continue, {"tools": "tools", END: END})
    graph.add_edge("tools", "agent")
    return graph.compile(checkpointer=MemorySaver())


class _LazyGraph:
    _instance = None

    def __getattr__(self, name):
        if self._instance is None:
            type(self)._instance = _build_enduser_graph()
        return getattr(self._instance, name)


enduser_graph = _LazyGraph()
```

- [ ] **Step 4: Run tests — expect PASS**

```bash
cd modelcraft-agent && python -m pytest tests/agents/test_enduser_agent.py -v
```

Expected: 4 tests PASS

- [ ] **Step 5: Commit**

```bash
git add modelcraft-agent/agents/enduser_agent.py modelcraft-agent/tests/agents/test_enduser_agent.py
git commit -m "feat(agent): add enduser_agent (query/filter only, no list_projects)"
```

---

## Task 5: Update `main.py` + `route.ts`, delete `agent.py`

**Why two endpoints?**
`RunAgentInput` (the AG-UI body) does NOT contain `agentId`.
CopilotKit runtime routes by mapping agent name → URL, so each agent needs its own endpoint:
- `/copilotkit/admin` → `modelcraft_admin_agent`
- `/copilotkit/enduser` → `modelcraft_enduser_agent`

**Files:**
- Modify: `modelcraft-agent/main.py`
- Modify: `modelcraft-front/src/app/api/copilotkit/route.ts`
- Delete: `modelcraft-agent/agent.py`

- [ ] **Step 1: Rewrite `main.py`**

```python
# modelcraft-agent/main.py
"""
FastAPI entry point for modelcraft-agent.

Two agents on one service, each with its own endpoint:
  POST /copilotkit/admin   → modelcraft_admin_agent  (tenant admins)
  POST /copilotkit/enduser → modelcraft_enduser_agent (end users)

CopilotKit runtime (route.ts) maps agent name → URL; routing is done
there, not here. Each endpoint simply injects Authorization and runs
the appropriate graph.
"""
import uvicorn
from fastapi import FastAPI, Request
from fastapi.responses import StreamingResponse
from ag_ui_langgraph.endpoint import RunAgentInput, EventEncoder
from copilotkit import LangGraphAGUIAgent

import config
from agents.admin_agent import admin_graph
from agents.enduser_agent import enduser_graph
from logging_setup import setup_logging
from middleware import ObservabilityMiddleware

app = FastAPI(title="modelcraft-agent", version="0.2.0")

setup_logging()
app.add_middleware(ObservabilityMiddleware)

_admin_agent = LangGraphAGUIAgent(
    name="modelcraft_admin_agent",
    description="ModelCraft AI 助手（管理员版）：项目管理、建模、数据查询",
    graph=admin_graph,
)

_enduser_agent = LangGraphAGUIAgent(
    name="modelcraft_enduser_agent",
    description="ModelCraft AI 助手（用户版）：数据查询与自然语言筛选",
    graph=enduser_graph,
)


@app.get("/healthz")
async def healthz():
    return {"status": "ok", "service": "modelcraft-agent"}


def _inject_authorization(input_data: RunAgentInput, request: Request) -> RunAgentInput:
    """Extract Authorization header and write it into graph state on every request."""
    authorization = request.headers.get("Authorization", "")
    current_state = dict(input_data.state) if input_data.state else {}
    current_state["authorization"] = authorization
    return input_data.model_copy(update={"state": current_state})


def _make_handler(agent: LangGraphAGUIAgent):
    async def handler(input_data: RunAgentInput, request: Request):
        accept_header = request.headers.get("accept")
        encoder = EventEncoder(accept=accept_header)
        input_data = _inject_authorization(input_data, request)
        request_agent = agent.clone()

        async def event_generator():
            async for event in request_agent.run(input_data):
                yield encoder.encode(event)

        return StreamingResponse(event_generator(), media_type=encoder.get_content_type())
    return handler


app.post("/copilotkit/admin")(_make_handler(_admin_agent))
app.post("/copilotkit/enduser")(_make_handler(_enduser_agent))


@app.get("/copilotkit/health")
def copilotkit_health():
    return {
        "status": "ok",
        "agents": [
            {"name": _admin_agent.name, "endpoint": "/copilotkit/admin"},
            {"name": _enduser_agent.name, "endpoint": "/copilotkit/enduser"},
        ],
    }


if __name__ == "__main__":
    uvicorn.run("main:app", host="0.0.0.0", port=config.PORT, reload=True)
```

- [ ] **Step 2: Update `route.ts` to register both agents**

```typescript
// modelcraft-front/src/app/api/copilotkit/route.ts
/**
 * CopilotKit Runtime endpoint — Next.js App Router
 *
 * Two agents, each pointing to its own Python endpoint:
 *   modelcraft_admin_agent  → AGENT_SERVICE_URL/copilotkit/admin
 *   modelcraft_enduser_agent → AGENT_SERVICE_URL/copilotkit/enduser
 */
import {
  CopilotRuntime,
  ExperimentalEmptyAdapter,
  copilotRuntimeNextJSAppRouterEndpoint,
} from "@copilotkit/runtime"
import { LangGraphHttpAgent } from "@copilotkit/runtime/langgraph"
import { NextRequest } from "next/server"

export const maxDuration = 60

const AGENT_SERVICE_URL = process.env.AGENT_SERVICE_URL ?? "http://localhost:8000"

const serviceAdapter = new ExperimentalEmptyAdapter()

export const POST = async (req: NextRequest) => {
  const authorization = req.headers.get("Authorization") ?? ""
  const authHeaders = authorization ? { Authorization: authorization } : {}

  const runtime = new CopilotRuntime({
    agents: {
      modelcraft_admin_agent: new LangGraphHttpAgent({
        url: `${AGENT_SERVICE_URL}/copilotkit/admin`,
        headers: authHeaders,
      }),
      modelcraft_enduser_agent: new LangGraphHttpAgent({
        url: `${AGENT_SERVICE_URL}/copilotkit/enduser`,
        headers: authHeaders,
      }),
    },
  })

  const { handleRequest } = copilotRuntimeNextJSAppRouterEndpoint({
    runtime,
    serviceAdapter,
    endpoint: "/api/copilotkit",
  })

  return handleRequest(req)
}
```

- [ ] **Step 3: Verify both modules import cleanly**

```bash
cd modelcraft-agent && python -c "from main import app; print('main.py OK')"
cd modelcraft-front && npx tsc --noEmit 2>&1 | grep "route.ts\|error TS" | head -5
```

Expected: `main.py OK` and no TypeScript errors

- [ ] **Step 4: Run all agent tests**

```bash
cd modelcraft-agent && python -m pytest tests/ -v
```

Expected: All PASS

- [ ] **Step 5: Delete `agent.py`**

```bash
git rm modelcraft-agent/agent.py
```

- [ ] **Step 6: Commit**

```bash
git add modelcraft-agent/main.py modelcraft-front/src/app/api/copilotkit/route.ts
git commit -m "feat(agent): two separate endpoints /copilotkit/admin + /copilotkit/enduser, update route.ts"
```

---

## Task 6: Add new nav tools to `ProjectCopilotActions.tsx`

**Files:**
- Modify: `modelcraft-front/src/web/components/features/copilot/ProjectCopilotActions.tsx`

- [ ] **Step 1: Add the 4 new `useCopilotAction` calls inside `ProjectCopilotActions`**

Open `src/web/components/features/copilot/ProjectCopilotActions.tsx` and append these four actions before the closing `return null`:

```tsx
  useCopilotAction({
    name: 'navigate_to_enums',
    description: '跳转到枚举管理页',
    parameters: [
      { name: 'db', type: 'string', description: '数据库名称（可选）', required: false },
    ],
    handler: async ({ db }: { db?: string }) => {
      const query = db ? `?db=${db}` : ''
      router.push(`/org/${orgName}/project/${projectSlug}/enums${query}`)
      return '已跳转到枚举管理页'
    },
  })

  useCopilotAction({
    name: 'navigate_to_cluster',
    description: '跳转到数据库集群配置页',
    parameters: [],
    handler: async () => {
      router.push(`/org/${orgName}/project/${projectSlug}/settings`)
      return '已跳转到集群配置页'
    },
  })

  useCopilotAction({
    name: 'navigate_to_rbac',
    description: '跳转到 RBAC 权限管理页',
    parameters: [
      {
        name: 'section',
        type: 'string',
        description: 'roles | users | bundles | permissions（默认 roles）',
        required: false,
      },
    ],
    handler: async ({ section }: { section?: string }) => {
      const sub = section ?? 'roles'
      router.push(`/org/${orgName}/project/${projectSlug}/rbac/${sub}`)
      return `已跳转到 RBAC ${sub} 页`
    },
  })

  useCopilotAction({
    name: 'navigate_to_end_users',
    description: '跳转到 end-user 管理页',
    parameters: [],
    handler: async () => {
      router.push(`/org/${orgName}/project/${projectSlug}/end-users`)
      return '已跳转到 end-user 管理页'
    },
  })
```

- [ ] **Step 2: Verify no TypeScript errors**

```bash
cd modelcraft-front && npx tsc --noEmit 2>&1 | grep -i "CopilotActions\|error" | head -10
```

Expected: no errors

- [ ] **Step 3: Commit**

```bash
git add modelcraft-front/src/web/components/features/copilot/ProjectCopilotActions.tsx
git commit -m "feat(front): add navigate_to_enums/cluster/rbac/end_users tools to ProjectCopilotActions"
```

---

## Task 7: Create `SharedCopilotActions.tsx` (show_toast)

**Files:**
- Create: `modelcraft-front/src/web/components/features/copilot/SharedCopilotActions.tsx`

- [ ] **Step 1: Create the component**

```tsx
// src/web/components/features/copilot/SharedCopilotActions.tsx
'use client'

import { memo } from 'react'
import { useCopilotAction } from '@copilotkit/react-core'
import { toast } from 'sonner'

/**
 * Frontend tools shared between admin and end-user agents.
 * Mount inside any CopilotKit context tree.
 *
 * Registers:
 *   show_toast — agent sends a one-line notification to the user
 */
export const SharedCopilotActions = memo(function SharedCopilotActions() {
  useCopilotAction({
    name: 'show_toast',
    description: '向用户显示一条临时通知消息（不需要用户在聊天框内查看）',
    parameters: [
      {
        name: 'message',
        type: 'string',
        description: '通知内容',
        required: true,
      },
      {
        name: 'type',
        type: 'string',
        description: 'success | error | info | warning（默认 info）',
        required: false,
      },
    ],
    handler: async ({ message, type }: { message: string; type?: string }) => {
      const fn = (type === 'success' ? toast.success
        : type === 'error' ? toast.error
        : type === 'warning' ? toast.warning
        : toast.info)
      fn(message)
      return 'toast displayed'
    },
  })

  return null
})
```

- [ ] **Step 2: Check TypeScript**

```bash
cd modelcraft-front && npx tsc --noEmit 2>&1 | grep "SharedCopilotActions\|error" | head -10
```

Expected: no errors

- [ ] **Step 3: Commit**

```bash
git add modelcraft-front/src/web/components/features/copilot/SharedCopilotActions.tsx
git commit -m "feat(front): add SharedCopilotActions with show_toast tool"
```

---

## Task 8: Create `EndUserCopilotActions.tsx`

**Files:**
- Create: `modelcraft-front/src/web/components/features/copilot/EndUserCopilotActions.tsx`

- [ ] **Step 1: Create the component**

```tsx
// src/web/components/features/copilot/EndUserCopilotActions.tsx
'use client'

import { memo } from 'react'
import { useCopilotAction } from '@copilotkit/react-core'
import { useRouter } from 'next/navigation'

interface EndUserCopilotActionsProps {
  orgName: string
  projectSlug: string
}

/**
 * Frontend navigation tools for end-user routes.
 * Registers: navigate_to_project, navigate_to_workspace
 */
export const EndUserCopilotActions = memo(function EndUserCopilotActions({
  orgName,
  projectSlug: _projectSlug,
}: EndUserCopilotActionsProps) {
  const router = useRouter()

  useCopilotAction({
    name: 'navigate_to_project',
    description: '切换到指定项目（end-user 路由）',
    parameters: [
      { name: 'slug', type: 'string', description: '项目 slug', required: true },
    ],
    handler: async ({ slug }: { slug: string }) => {
      router.push(`/end-user/${orgName}/projects/${slug}/data`)
      return `已切换到项目 ${slug}`
    },
  })

  useCopilotAction({
    name: 'navigate_to_workspace',
    description: '返回项目选择页',
    parameters: [],
    handler: async () => {
      router.push(`/end-user/${orgName}/select-project`)
      return '已返回项目选择页'
    },
  })

  return null
})
```

- [ ] **Step 2: Check TypeScript**

```bash
cd modelcraft-front && npx tsc --noEmit 2>&1 | grep "EndUserCopilotActions\|error" | head -10
```

Expected: no errors

- [ ] **Step 3: Commit**

```bash
git add modelcraft-front/src/web/components/features/copilot/EndUserCopilotActions.tsx
git commit -m "feat(front): add EndUserCopilotActions with navigate_to_project and navigate_to_workspace"
```

---

## Task 9: Create knowledge and suggestion components

**Files:**
- Create: `modelcraft-front/src/web/components/features/copilot/AdminCopilotKnowledge.tsx`
- Create: `modelcraft-front/src/web/components/features/copilot/EndUserCopilotKnowledge.tsx`

- [ ] **Step 1: Create `AdminCopilotKnowledge.tsx`**

```tsx
// src/web/components/features/copilot/AdminCopilotKnowledge.tsx
'use client'

import { memo } from 'react'
import { useCopilotReadable } from '@copilotkit/react-core'
import { useCopilotChatSuggestions } from '@copilotkit/react-ui'

const ADMIN_ONBOARDING = `
新手引导共 5 步：

Step 1 [创建项目]
  目标：创建第一个项目
  工具：open_create_project(slug, title)
  验证：调用 list_projects，确认项目出现

Step 2 [配置数据库集群]
  目标：连接一个数据库
  工具：navigate_to_cluster，引导用户在页面上手动填写连接信息
  提示：集群配置需要用户提供数据库连接信息，agent 无法代替操作

Step 3 [创建数据模型]
  目标：在项目下创建第一个模型
  工具：navigate_to_project → open_create_model(db, name)
  验证：调用 list_models 确认模型存在

Step 4 [添加字段]
  目标：给模型添加字段
  工具：navigate_to_model(db, model)
  提示：字段编辑在右侧面板，由用户手动操作

Step 5 [查看数据]
  目标：进入数据视图，确认配置完成
  工具：navigate_to_data(db, model)
`.trim()

const ADMIN_TROUBLESHOOTING = `
常见问题排查：

问题：数据库连接失败
  → show_toast("正在带你去检查集群配置", "info")
  → navigate_to_cluster，引导检查 host/port/credentials

问题：找不到模型或字段
  → 调用 list_models(db) 确认模型名是否正确
  → list_models 返回空则模型未创建，建议执行新手引导 Step 3

问题：权限被拒绝
  → navigate_to_rbac(section="users")，检查用户角色分配

问题：字段显示异常
  → navigate_to_model(db, model)，检查字段类型和配置
`.trim()

const ADMIN_SUGGESTIONS = [
  { title: '新手引导：带我完成初始配置', message: '我是新用户，请帮我按步骤完成初始配置' },
  { title: '帮我创建第一个数据模型', message: '帮我创建第一个数据模型' },
  { title: '数据库连不上，帮我排查', message: '数据库连接有问题，请帮我排查' },
  { title: '我有哪些项目？', message: '列出我所有的项目' },
  { title: '解释当前页面的功能', message: '请解释当前页面有哪些功能' },
]

/**
 * Injects admin knowledge base and sidebar suggestions into CopilotKit context.
 * Must be mounted inside a CopilotKit provider tree.
 */
export const AdminCopilotKnowledge = memo(function AdminCopilotKnowledge() {
  useCopilotReadable({
    description: 'ModelCraft 新手引导操作手册（管理员）',
    value: ADMIN_ONBOARDING,
  })

  useCopilotReadable({
    description: 'ModelCraft 常见问题排查手册（管理员）',
    value: ADMIN_TROUBLESHOOTING,
  })

  useCopilotChatSuggestions({
    suggestions: ADMIN_SUGGESTIONS,
    available: 'before-first-message',
  })

  return null
})
```

- [ ] **Step 2: Create `EndUserCopilotKnowledge.tsx`**

```tsx
// src/web/components/features/copilot/EndUserCopilotKnowledge.tsx
'use client'

import { memo } from 'react'
import { useCopilotReadable } from '@copilotkit/react-core'
import { useCopilotChatSuggestions } from '@copilotkit/react-ui'

const ENDUSER_ONBOARDING = `
新手引导共 3 步：

Step 1 [了解当前数据]
  目标：知道项目里有哪些数据
  工具：调用 list_models，用自然语言介绍每个模型的用途

Step 2 [学会筛选]
  目标：用自然语言筛选数据
  步骤：
    1. 询问用户想查什么
    2. 调用 nl2filter(natural_language, field_names) 生成 filter JSON
    3. 调用前端 set_filter(filter_json) 应用筛选
  示例引导语：「你可以说"帮我找金额大于 1000 的订单"，我来帮你筛选」

Step 3 [理解字段含义]
  目标：用户看懂表格里的每一列
  工具：get_model_fields(model_id) → 逐字段用中文解释
`.trim()

const ENDUSER_TROUBLESHOOTING = `
常见问题排查：

问题：看不到数据
  → 先调用前端 clear_filter 排除筛选遮挡
  → 若仍无数据，说明可能没有访问权限，提示联系管理员

问题：不知道怎么筛选
  → 引导用户用自然语言描述需求
  → 执行 nl2filter + set_filter

问题：字段看不懂
  → 调用 get_model_fields(model_id)，逐字段解释含义和示例值

问题：数据量太大，加载慢
  → 引导用户说出筛选条件
  → 用 nl2filter 缩小数据范围后再查看
`.trim()

const ENDUSER_SUGGESTIONS = [
  { title: '新手引导：带我了解这个系统', message: '我是新用户，请带我了解这个系统' },
  { title: '用自然语言帮我筛选数据', message: '帮我筛选数据，我来描述条件' },
  { title: '我看不到想要的数据，帮我排查', message: '我看不到想要的数据，请帮我排查' },
  { title: '这些字段分别是什么意思？', message: '请解释当前表格里各个字段的含义' },
  { title: '帮我统计一下数据', message: '帮我统计一下当前模型的数据情况' },
]

/**
 * Injects end-user knowledge base and sidebar suggestions into CopilotKit context.
 * Must be mounted inside EndUserCopilotWrapper.
 */
export const EndUserCopilotKnowledge = memo(function EndUserCopilotKnowledge() {
  useCopilotReadable({
    description: 'ModelCraft 新手引导操作手册（终端用户）',
    value: ENDUSER_ONBOARDING,
  })

  useCopilotReadable({
    description: 'ModelCraft 常见问题排查手册（终端用户）',
    value: ENDUSER_TROUBLESHOOTING,
  })

  useCopilotChatSuggestions({
    suggestions: ENDUSER_SUGGESTIONS,
    available: 'before-first-message',
  })

  return null
})
```

- [ ] **Step 3: Check TypeScript**

```bash
cd modelcraft-front && npx tsc --noEmit 2>&1 | grep -E "CopilotKnowledge|error TS" | head -10
```

Expected: no errors

- [ ] **Step 4: Commit**

```bash
git add modelcraft-front/src/web/components/features/copilot/AdminCopilotKnowledge.tsx \
        modelcraft-front/src/web/components/features/copilot/EndUserCopilotKnowledge.tsx
git commit -m "feat(front): add AdminCopilotKnowledge and EndUserCopilotKnowledge with suggestions"
```

---

## Task 10: Update `CopilotProvider.tsx` — wire everything together

**Files:**
- Modify: `modelcraft-front/src/web/components/features/copilot/CopilotProvider.tsx`

- [ ] **Step 1: Update `CopilotProvider.tsx`**

Replace the entire file:

```tsx
'use client'

import { memo, useMemo, Suspense } from 'react'
import dynamic from 'next/dynamic'
import type { Project } from '@/types'
import { CopilotAvailableContext } from '@web/components/features/end-user-data/FilterCopilotActions'
import { useAuthStore } from '@shared/stores/auth-store'
import { useEndUserAuthStore } from '@shared/stores/end-user-auth-store'
import { SharedCopilotActions } from './SharedCopilotActions'
import { AdminCopilotKnowledge } from './AdminCopilotKnowledge'
import { EndUserCopilotActions } from './EndUserCopilotActions'
import { EndUserCopilotKnowledge } from './EndUserCopilotKnowledge'

const CopilotKit = dynamic(
  () => import('@copilotkit/react-core').then(mod => mod.CopilotKit),
  { ssr: false }
)

const CopilotSidebar = dynamic(
  () => import('@copilotkit/react-ui').then(mod => mod.CopilotSidebar),
  { ssr: false }
)

interface CopilotProviderProps {
  children: React.ReactNode
  selectedProject: Project | null
  orgName: string
}

/**
 * Inner provider for the admin (tenant) surface.
 * Mounts SharedCopilotActions + AdminCopilotKnowledge inside CopilotKit context.
 */
const CopilotProvider = memo(({ children, selectedProject, orgName }: CopilotProviderProps) => {
  const accessToken = useAuthStore((s) => s.accessToken)

  const copilotContext = useMemo(() => ({
    projectId: selectedProject?.id || 'default',
    projectSlug: selectedProject?.slug || 'Default Project',
    orgName,
  }), [selectedProject?.id, selectedProject?.slug, orgName])

  const headers = useMemo<Record<string, string> | undefined>(
    () => accessToken ? { Authorization: `Bearer ${accessToken}` } : undefined,
    [accessToken]
  )

  const initialMessage = useMemo(() => {
    const projectSlug = selectedProject?.slug || 'Default Project'
    return `你好！我是 ModelCraft AI 助手，当前项目：${projectSlug}。

我可以帮助你：

• 创建和管理数据库集群
• 设计数据模型和字段
• 配置枚举类型
• 管理项目

请问有什么可以帮助你的？`
  }, [selectedProject?.slug])

  return (
    <CopilotKit
      runtimeUrl="/api/copilotkit"
      agent="modelcraft_admin_agent"
      headers={headers}
      properties={copilotContext}
    >
      <SharedCopilotActions />
      <AdminCopilotKnowledge />
      {children}
      <CopilotSidebar
        labels={{
          title: 'ModelCraft AI 助手',
          initial: initialMessage,
        }}
        defaultOpen={false}
        clickOutsideToClose={true}
      />
    </CopilotKit>
  )
})

CopilotProvider.displayName = 'CopilotProvider'

/**
 * Wrapper for admin (tenant) routes — org/* and project/* layouts.
 */
export const CopilotWrapper = memo(({
  children,
  selectedProject,
  orgName,
}: CopilotProviderProps) => {
  return (
    <CopilotAvailableContext.Provider value={true}>
      <Suspense fallback={children}>
        <CopilotProvider selectedProject={selectedProject} orgName={orgName}>
          {children}
        </CopilotProvider>
      </Suspense>
    </CopilotAvailableContext.Provider>
  )
})

CopilotWrapper.displayName = 'CopilotWrapper'

interface EndUserCopilotWrapperProps {
  children: React.ReactNode
  orgName: string
  projectSlug: string
}

/**
 * Wrapper for end-user routes — mounts enduser-specific tools, knowledge, and sidebar.
 */
export const EndUserCopilotWrapper = memo(({
  children,
  orgName,
  projectSlug,
}: EndUserCopilotWrapperProps) => {
  const accessToken = useEndUserAuthStore((s) => s.accessToken)

  const copilotContext = useMemo(() => ({
    orgName,
    projectSlug,
  }), [orgName, projectSlug])

  const headers = useMemo<Record<string, string> | undefined>(
    () => accessToken ? { Authorization: `Bearer ${accessToken}` } : undefined,
    [accessToken]
  )

  const initialMessage = useMemo(() => `你好！我是 ModelCraft AI 助手，当前项目：${projectSlug}。

我可以帮助你：

• 查询和筛选数据
• 分析数据记录

请问有什么可以帮助你的？`, [projectSlug])

  return (
    <CopilotAvailableContext.Provider value={true}>
      <Suspense fallback={children}>
        <CopilotKit
          runtimeUrl="/api/copilotkit"
          agent="modelcraft_enduser_agent"
          headers={headers}
          properties={copilotContext}
        >
          <SharedCopilotActions />
          <EndUserCopilotKnowledge />
          <EndUserCopilotActions orgName={orgName} projectSlug={projectSlug} />
          {children}
          <CopilotSidebar
            labels={{
              title: 'ModelCraft AI 助手',
              initial: initialMessage,
            }}
            defaultOpen={false}
          />
        </CopilotKit>
      </Suspense>
    </CopilotAvailableContext.Provider>
  )
})

EndUserCopilotWrapper.displayName = 'EndUserCopilotWrapper'
```

- [ ] **Step 2: Check TypeScript and diagnostics**

```bash
cd modelcraft-front && npx tsc --noEmit 2>&1 | grep -E "CopilotProvider|error TS" | head -10
```

Expected: no errors

- [ ] **Step 3: Commit**

```bash
git add modelcraft-front/src/web/components/features/copilot/CopilotProvider.tsx
git commit -m "feat(front): wire SharedCopilotActions + knowledge into both providers, use new agent names"
```

---

## Task 11: Final verification

- [ ] **Step 1: Verify no stale `modelcraft_agent` references (old single-agent name)**

```bash
grep -r '"modelcraft_agent"' \
  modelcraft-front/src \
  modelcraft-agent/ \
  --include="*.tsx" --include="*.ts" --include="*.py" \
  --exclude-dir=node_modules
```

Expected: no output (all references now use `modelcraft_admin_agent` or `modelcraft_enduser_agent`)

- [ ] **Step 2: Verify Python endpoints exist**

```bash
cd modelcraft-agent && python -c "
from main import app
routes = [r.path for r in app.routes]
assert '/copilotkit/admin' in routes, 'admin endpoint missing'
assert '/copilotkit/enduser' in routes, 'enduser endpoint missing'
print('Routes OK:', [r for r in routes if 'copilotkit' in r])
"
```

Expected:
```
Routes OK: ['/copilotkit/admin', '/copilotkit/enduser', '/copilotkit/health']
```

- [ ] **Step 3: Lint the frontend**

```bash
cd modelcraft-front && npm run lint 2>&1 | tail -5
```

Expected: `✔ No ESLint warnings or errors`

- [ ] **Step 4: Run all agent tests**

```bash
cd modelcraft-agent && python -m pytest tests/ -v
```

Expected: All tests PASS

- [ ] **Step 5: Final commit**

```bash
git add -A
git commit -m "feat: complete CopilotKit two-agent architecture

- modelcraft_admin_agent  → /copilotkit/admin  (list_projects + full toolset)
- modelcraft_enduser_agent → /copilotkit/enduser (query/filter only)
- SharedCopilotActions: show_toast in both contexts
- AdminCopilotKnowledge / EndUserCopilotKnowledge: onboarding + troubleshooting + suggestions
- route.ts: both agents registered with separate URLs"
```

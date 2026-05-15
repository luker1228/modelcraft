# model-agent 可观测性 Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** 为 `modelcraft-agent` 添加 structlog JSON 结构化日志 + requestId 追踪，覆盖 HTTP 请求、LLM 调用、Tool 执行、GraphQL 下游调用和异常。

**Architecture:** FastAPI 中间件从请求头提取 `X-Request-ID` / `X-Client-Request-Id`，通过 `structlog.contextvars.bind_contextvars()` 注入进程级 ContextVar；后续所有层（agent、graphql_client）通过 `get_logger()` 获取自动携带 requestId 的 logger，手动埋点关键事件。

**Tech Stack:** Python 3.11+, FastAPI, structlog>=24.0, pytest>=8.0, pytest-asyncio>=0.23, respx (httpx mock)

---

## File Map

| 操作 | 路径 | 职责 |
|------|------|------|
| 新建 | `modelcraft-agent/logging_setup.py` | structlog 初始化 + `get_logger()` |
| 新建 | `modelcraft-agent/middleware.py` | `ObservabilityMiddleware`：提取 header、bind contextvars、记录 request.start/end |
| 新建 | `modelcraft-agent/tests/__init__.py` | pytest 包标识 |
| 新建 | `modelcraft-agent/tests/conftest.py` | 共享 fixture：`capture_logs_with_context()` 上下文管理器 |
| 新建 | `modelcraft-agent/tests/test_logging_setup.py` | 测试 `setup_logging` 和 `get_logger` |
| 新建 | `modelcraft-agent/tests/test_middleware.py` | 测试中间件 header 提取、fallback UUID、日志字段 |
| 新建 | `modelcraft-agent/tests/test_graphql_client.py` | 测试 GraphQL 日志埋点（respx mock） |
| 新建 | `modelcraft-agent/pytest.ini` | asyncio_mode = auto |
| 修改 | `modelcraft-agent/requirements.txt` | 添加 structlog, pytest, pytest-asyncio, respx |
| 修改 | `modelcraft-agent/main.py` | 注册 ObservabilityMiddleware，启动时调用 `setup_logging()` |
| 修改 | `modelcraft-agent/agent.py` | LLM call 日志（on_chat_model_start/end）+ tool call 日志 |
| 修改 | `modelcraft-agent/client/graphql_client.py` | graphql.call.start/end + error 日志 |

---

## Task 1: 添加依赖

**Files:**
- Modify: `modelcraft-agent/requirements.txt`

- [ ] **Step 1: 更新 requirements.txt**

  打开 `modelcraft-agent/requirements.txt`，替换为：

  ```text
  fastapi==0.115.12
  uvicorn[standard]==0.34.2
  copilotkit==0.1.89
  langgraph>=0.3.25,<2
  langchain-openai>=0.2.0
  httpx==0.28.1
  python-dotenv==1.1.0
  structlog>=24.0.0
  # dev/test
  pytest>=8.0.0
  pytest-asyncio>=0.23.0
  respx>=0.21.0
  ```

- [ ] **Step 2: 验证安装（在 modelcraft-agent 目录执行）**

  ```bash
  cd modelcraft-agent && pip install -r requirements.txt
  ```

  期望：无报错，输出最后一行类似 `Successfully installed structlog-24.x ...`

- [ ] **Step 3: Commit**

  ```bash
  git add modelcraft-agent/requirements.txt
  git commit -m "chore(agent): add structlog, pytest, respx dependencies"
  ```

---

## Task 2: 创建 logging_setup.py（含测试）

**Files:**
- Create: `modelcraft-agent/logging_setup.py`
- Create: `modelcraft-agent/pytest.ini`
- Create: `modelcraft-agent/tests/__init__.py`
- Create: `modelcraft-agent/tests/conftest.py`
- Create: `modelcraft-agent/tests/test_logging_setup.py`

- [ ] **Step 1: 创建 pytest.ini**

  ```ini
  [pytest]
  asyncio_mode = auto
  testpaths = tests
  ```

- [ ] **Step 2: 写失败测试**

  创建 `modelcraft-agent/tests/__init__.py`（空文件）。

  创建 `modelcraft-agent/tests/conftest.py`：

  ```python
  """Shared test utilities."""
  from contextlib import contextmanager
  from typing import Generator

  import structlog
  from structlog.testing import LogCapture
  from structlog.contextvars import merge_contextvars


  @contextmanager
  def capture_logs_with_context() -> Generator[list, None, None]:
      """Like structlog.testing.capture_logs() but also merges contextvars.

      Use instead of capture_logs() when testing code that calls bind_contextvars().
      """
      cap = LogCapture()
      old_processors = structlog.get_config()["processors"]
      structlog.configure(processors=[merge_contextvars, cap])
      try:
          yield cap.entries
      finally:
          structlog.configure(processors=old_processors)
  ```

  创建 `modelcraft-agent/tests/test_logging_setup.py`：

  ```python
  """Tests for logging_setup module."""
  import structlog
  from structlog.testing import capture_logs

  from logging_setup import setup_logging, get_logger


  def test_get_logger_binds_service_field():
      setup_logging()
      logger = get_logger()
      with capture_logs() as cap:
          logger.info("test.event", key="value")
      assert len(cap) == 1
      assert cap[0]["service"] == "modelcraft-agent"
      assert cap[0]["event"] == "test.event"
      assert cap[0]["key"] == "value"


  def test_setup_logging_is_idempotent():
      """Calling setup_logging() twice should not raise."""
      setup_logging()
      setup_logging()
      logger = get_logger()
      with capture_logs() as cap:
          logger.info("idempotent.check")
      assert cap[0]["event"] == "idempotent.check"
  ```

- [ ] **Step 3: 运行测试，确认失败**

  ```bash
  cd modelcraft-agent && python -m pytest tests/test_logging_setup.py -v
  ```

  期望：`ImportError: No module named 'logging_setup'`

- [ ] **Step 4: 实现 logging_setup.py**

  ```python
  """
  Structlog JSON logging setup for modelcraft-agent.

  Usage:
      from logging_setup import setup_logging, get_logger

      setup_logging()          # once at process startup
      log = get_logger()
      log.info("my.event", key="value")
  """
  import logging

  import structlog


  def setup_logging() -> None:
      """Configure structlog with JSON output. Call once at process startup."""
      structlog.configure(
          processors=[
              structlog.contextvars.merge_contextvars,
              structlog.processors.add_log_level,
              structlog.processors.TimeStamper(fmt="iso", utc=True),
              structlog.processors.StackInfoRenderer(),
              structlog.processors.ExceptionRenderer(),
              structlog.processors.JSONRenderer(),
          ],
          wrapper_class=structlog.make_filtering_bound_logger(logging.INFO),
          logger_factory=structlog.PrintLoggerFactory(),
          cache_logger_on_first_use=False,
      )


  def get_logger() -> structlog.BoundLogger:
      """Return a structlog logger bound with service=modelcraft-agent."""
      return structlog.get_logger().bind(service="modelcraft-agent")
  ```

  > `cache_logger_on_first_use=False` — 允许测试中多次重新配置 structlog。

- [ ] **Step 5: 运行测试，确认通过**

  ```bash
  cd modelcraft-agent && python -m pytest tests/test_logging_setup.py -v
  ```

  期望：
  ```
  tests/test_logging_setup.py::test_get_logger_binds_service_field PASSED
  tests/test_logging_setup.py::test_setup_logging_is_idempotent PASSED
  ```

- [ ] **Step 6: Commit**

  ```bash
  git add modelcraft-agent/logging_setup.py \
          modelcraft-agent/pytest.ini \
          modelcraft-agent/tests/__init__.py \
          modelcraft-agent/tests/conftest.py \
          modelcraft-agent/tests/test_logging_setup.py
  git commit -m "feat(agent): add structlog setup with get_logger()"
  ```

---

## Task 3: 创建 ObservabilityMiddleware（含测试）

**Files:**
- Create: `modelcraft-agent/middleware.py`
- Create: `modelcraft-agent/tests/test_middleware.py`

- [ ] **Step 1: 写失败测试**

  创建 `modelcraft-agent/tests/test_middleware.py`：

  ```python
  """Tests for ObservabilityMiddleware."""
  import pytest
  from fastapi import FastAPI
  from fastapi.testclient import TestClient

  from logging_setup import setup_logging
  from tests.conftest import capture_logs_with_context


  @pytest.fixture(autouse=True)
  def _setup_logging():
      setup_logging()


  @pytest.fixture
  def app():
      """Minimal FastAPI app with ObservabilityMiddleware."""
      from middleware import ObservabilityMiddleware
      _app = FastAPI()
      _app.add_middleware(ObservabilityMiddleware)

      @_app.get("/test")
      def test_endpoint():
          return {"ok": True}

      return _app


  def test_uses_request_id_from_gateway_header(app):
      with TestClient(app) as client:
          with capture_logs_with_context() as cap:
              client.get("/test", headers={"X-Request-ID": "gw-req-001"})

      start_log = next(l for l in cap if l["event"] == "request.start")
      assert start_log["request_id"] == "gw-req-001"


  def test_generates_fallback_request_id_when_header_missing(app):
      with TestClient(app) as client:
          with capture_logs_with_context() as cap:
              client.get("/test")

      start_log = next(l for l in cap if l["event"] == "request.start")
      # fallback UUID should be non-empty
      assert start_log.get("request_id", "")


  def test_captures_optional_client_request_id(app):
      with TestClient(app) as client:
          with capture_logs_with_context() as cap:
              client.get("/test", headers={
                  "X-Request-ID": "gw-req-002",
                  "X-Client-Request-Id": "fe-uuid-abc",
              })

      start_log = next(l for l in cap if l["event"] == "request.start")
      assert start_log["client_request_id"] == "fe-uuid-abc"


  def test_request_end_has_status_code_and_duration(app):
      with TestClient(app) as client:
          with capture_logs_with_context() as cap:
              client.get("/test", headers={"X-Request-ID": "gw-req-003"})

      end_log = next(l for l in cap if l["event"] == "request.end")
      assert end_log["status_code"] == 200
      assert isinstance(end_log["duration_ms"], float)
      assert end_log["duration_ms"] >= 0


  def test_request_end_carries_same_request_id_as_start(app):
      with TestClient(app) as client:
          with capture_logs_with_context() as cap:
              client.get("/test", headers={"X-Request-ID": "gw-req-004"})

      start_log = next(l for l in cap if l["event"] == "request.start")
      end_log = next(l for l in cap if l["event"] == "request.end")
      assert start_log["request_id"] == end_log["request_id"] == "gw-req-004"
  ```

- [ ] **Step 2: 运行测试，确认失败**

  ```bash
  cd modelcraft-agent && python -m pytest tests/test_middleware.py -v
  ```

  期望：`ImportError: No module named 'middleware'`

- [ ] **Step 3: 实现 middleware.py**

  ```python
  """
  ObservabilityMiddleware for modelcraft-agent.

  Extracts X-Request-ID (from gateway) and X-Client-Request-Id (from frontend),
  binds them to structlog contextvars so all downstream log calls carry them
  automatically. Also records request.start and request.end events with timing.
  """
  import time
  import uuid

  import structlog
  from structlog.contextvars import bind_contextvars, clear_contextvars
  from starlette.middleware.base import BaseHTTPMiddleware
  from starlette.requests import Request
  from starlette.responses import Response


  class ObservabilityMiddleware(BaseHTTPMiddleware):
      async def dispatch(self, request: Request, call_next) -> Response:
          # Always start with a clean context to prevent bleed between requests.
          clear_contextvars()

          request_id = request.headers.get("X-Request-ID") or str(uuid.uuid4())
          client_req_id = request.headers.get("X-Client-Request-Id", "")

          bind_contextvars(
              request_id=request_id,
              client_request_id=client_req_id,
          )

          log = structlog.get_logger().bind(service="modelcraft-agent")
          log.info(
              "request.start",
              method=request.method,
              path=request.url.path,
              client_ip=request.client.host if request.client else "",
          )

          start = time.perf_counter()
          response: Response | None = None
          try:
              response = await call_next(request)
          except Exception:
              log.exception("error")
              raise
          finally:
              duration_ms = round((time.perf_counter() - start) * 1000, 2)
              status_code = response.status_code if response is not None else 500
              log.info(
                  "request.end",
                  method=request.method,
                  path=request.url.path,
                  status_code=status_code,
                  duration_ms=duration_ms,
              )

          return response
  ```

- [ ] **Step 4: 运行测试，确认通过**

  ```bash
  cd modelcraft-agent && python -m pytest tests/test_middleware.py -v
  ```

  期望：5 tests PASSED

- [ ] **Step 5: Commit**

  ```bash
  git add modelcraft-agent/middleware.py \
          modelcraft-agent/tests/test_middleware.py
  git commit -m "feat(agent): add ObservabilityMiddleware with requestId tracing"
  ```

---

## Task 4: 更新 main.py — 注册中间件

**Files:**
- Modify: `modelcraft-agent/main.py`

- [ ] **Step 1: 在 main.py 中注册 ObservabilityMiddleware 并调用 setup_logging()**

  打开 `modelcraft-agent/main.py`，做以下两处修改：

  **① 在 import 区末尾添加**（`import config` 后）：

  ```python
  from logging_setup import setup_logging
  from middleware import ObservabilityMiddleware
  ```

  **② 将 `app = FastAPI(...)` 后的代码修改为**：

  ```python
  app = FastAPI(title="modelcraft-agent", version="0.1.0")

  # Observability: must be added before any routes are registered.
  # Middleware is applied in reverse order of add_middleware() calls;
  # ObservabilityMiddleware is outermost so it wraps all requests.
  setup_logging()
  app.add_middleware(ObservabilityMiddleware)
  ```

  完整修改后 `main.py` 头部应如下（只展示变更部分，其余不变）：

  ```python
  import uuid
  import json
  import asyncio
  from typing import Optional, List, Any, AsyncGenerator

  import uvicorn
  from fastapi import FastAPI
  from langchain_core.messages import HumanMessage, AIMessage, ToolMessage
  from copilotkit import CopilotKitRemoteEndpoint, Agent
  from copilotkit.integrations.fastapi import add_fastapi_endpoint
  from copilotkit.action import ActionDict
  from copilotkit.types import Message, MetaEvent

  import config
  from agent import modelcraft_graph
  from logging_setup import setup_logging
  from middleware import ObservabilityMiddleware

  app = FastAPI(title="modelcraft-agent", version="0.1.0")

  setup_logging()
  app.add_middleware(ObservabilityMiddleware)
  ```

- [ ] **Step 2: 验证服务可以正常启动（smoke test）**

  ```bash
  cd modelcraft-agent && python -m uvicorn main:app --port 8001 &
  sleep 2 && curl -s http://localhost:8001/healthz
  kill %1
  ```

  期望：
  ```json
  {"status":"ok","service":"modelcraft-agent"}
  ```
  且终端输出包含 JSON 日志行：
  ```json
  {"service": "modelcraft-agent", "level": "info", "event": "request.start", "method": "GET", "path": "/healthz", ...}
  ```

- [ ] **Step 3: Commit**

  ```bash
  git add modelcraft-agent/main.py
  git commit -m "feat(agent): register ObservabilityMiddleware in FastAPI app"
  ```

---

## Task 5: 为 GraphQLClient 添加日志（含测试）

**Files:**
- Modify: `modelcraft-agent/client/graphql_client.py`
- Create: `modelcraft-agent/tests/test_graphql_client.py`

- [ ] **Step 1: 写失败测试**

  创建 `modelcraft-agent/tests/test_graphql_client.py`：

  ```python
  """Tests for GraphQLClient observability instrumentation."""
  import pytest
  import httpx
  import respx

  from logging_setup import setup_logging
  from tests.conftest import capture_logs_with_context
  from client.graphql_client import GraphQLClient

  MOCK_URL_PREFIX = "http://localhost:8090"


  @pytest.fixture(autouse=True)
  def _setup():
      setup_logging()


  @pytest.mark.asyncio
  @respx.mock
  async def test_find_many_logs_graphql_call_start_and_end():
      respx.post(url__startswith=MOCK_URL_PREFIX).mock(
          return_value=httpx.Response(
              200,
              json={"data": {"findMany": {"items": [{"id": "1"}], "totalCount": 1}}},
          )
      )
      client = GraphQLClient(authorization="Bearer test-token")

      with capture_logs_with_context() as cap:
          await client.find_many(
              org_name="org1",
              project_slug="proj1",
              db_name="maindb",
              model_name="User",
              fields=["id"],
          )

      start_log = next((l for l in cap if l["event"] == "graphql.call.start"), None)
      end_log = next((l for l in cap if l["event"] == "graphql.call.end"), None)

      assert start_log is not None, "graphql.call.start not logged"
      assert start_log["operation"] == "findMany"
      assert "url" in start_log

      assert end_log is not None, "graphql.call.end not logged"
      assert end_log["has_errors"] is False
      assert end_log["status_code"] == 200
      assert isinstance(end_log["duration_ms"], float)
      assert end_log["duration_ms"] >= 0


  @pytest.mark.asyncio
  @respx.mock
  async def test_find_many_logs_has_errors_true_when_graphql_errors_present():
      respx.post(url__startswith=MOCK_URL_PREFIX).mock(
          return_value=httpx.Response(
              200,
              json={"data": None, "errors": [{"message": "Not found"}]},
          )
      )
      client = GraphQLClient(authorization="Bearer test-token")

      with capture_logs_with_context() as cap:
          await client.find_many(
              org_name="org1",
              project_slug="proj1",
              db_name="maindb",
              model_name="User",
              fields=["id"],
          )

      end_log = next(l for l in cap if l["event"] == "graphql.call.end")
      assert end_log["has_errors"] is True


  @pytest.mark.asyncio
  @respx.mock
  async def test_find_many_logs_error_on_http_failure():
      respx.post(url__startswith=MOCK_URL_PREFIX).mock(
          return_value=httpx.Response(500, text="Internal Server Error")
      )
      client = GraphQLClient(authorization="Bearer test-token")

      with capture_logs_with_context() as cap:
          with pytest.raises(httpx.HTTPStatusError):
              await client.find_many(
                  org_name="org1",
                  project_slug="proj1",
                  db_name="maindb",
                  model_name="User",
                  fields=["id"],
              )

      error_log = next((l for l in cap if l["event"] == "error"), None)
      assert error_log is not None, "error event not logged on HTTP 500"
  ```

- [ ] **Step 2: 运行测试，确认失败**

  ```bash
  cd modelcraft-agent && python -m pytest tests/test_graphql_client.py -v
  ```

  期望：3 tests FAILED（找不到 `graphql.call.start` 等日志）

- [ ] **Step 3: 修改 client/graphql_client.py，添加日志**

  在文件头部添加 imports：
  ```python
  import time
  from logging_setup import get_logger
  ```

  将 `find_many` 方法体替换为：

  ```python
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
      validated_fields = fields if fields else ["id"]
      self._validate_fields(validated_fields)
      fields_str = " ".join(validated_fields)

      where_type = f"T{model_name}WhereInput"
      query = f"""
  query FindMany($take: Int, $skip: Int, $where: {where_type}) {{
    findMany(take: $take, skip: $skip, where: $where) {{
      items {{ {fields_str} }}
      totalCount
    }}
  }}
  """

      variables: dict[str, Any] = {"take": take, "skip": skip}
      if where is not None:
          variables["where"] = where

      url = self._build_url(org_name, project_slug, db_name, model_name)
      headers = {
          "Content-Type": "application/json",
          "Authorization": self._authorization,
      }
      payload = {"query": query, "variables": variables}

      log = get_logger()
      log.info("graphql.call.start", url=url, operation="findMany")
      start = time.perf_counter()
      try:
          async with httpx.AsyncClient(timeout=30.0) as client:
              response = await client.post(url, headers=headers, json=payload)
              response.raise_for_status()
              result = response.json()
      except Exception:
          duration_ms = round((time.perf_counter() - start) * 1000, 2)
          log.exception("error", url=url, operation="findMany", duration_ms=duration_ms)
          raise

      duration_ms = round((time.perf_counter() - start) * 1000, 2)
      has_errors = bool(result.get("errors"))
      log.info(
          "graphql.call.end",
          url=url,
          duration_ms=duration_ms,
          has_errors=has_errors,
          status_code=response.status_code,
      )
      return result
  ```

- [ ] **Step 4: 运行测试，确认通过**

  ```bash
  cd modelcraft-agent && python -m pytest tests/test_graphql_client.py -v
  ```

  期望：3 tests PASSED

- [ ] **Step 5: Commit**

  ```bash
  git add modelcraft-agent/client/graphql_client.py \
          modelcraft-agent/tests/test_graphql_client.py
  git commit -m "feat(agent): add graphql.call.* observability to GraphQLClient"
  ```

---

## Task 6: 为 agent.py 添加 LLM + Tool 日志

**Files:**
- Modify: `modelcraft-agent/agent.py`

> **注意**：`agent.py` 的工具函数使用 LangGraph `InjectedState`，无法用普通单元测试隔离调用。此 task 通过 smoke test（Task 4 已验证）以及代码审查保证正确性；LLM 调用在集成环境中可从日志输出验证。

- [ ] **Step 1: 在 agent.py 头部添加 imports**

  在 `import config` 行后添加：

  ```python
  import time
  from logging_setup import get_logger
  from structlog.contextvars import bind_contextvars
  ```

- [ ] **Step 2: 为 query_model tool 添加日志**

  将 `query_model` 函数体替换为：

  ```python
  @tool
  async def query_model(
      db_name: str,
      model_name: str,
      fields: list[str],
      take: int,
      state: Annotated[AgentState, InjectedState()],
      where: dict | None = None,
  ) -> str:
      """
      Query records from a ModelCraft data model.

      Args:
          db_name: Database name, e.g. "maindb"
          model_name: Model name, e.g. "users"
          fields: List of field names to return, e.g. ["id", "name", "createdAt"]
          take: Max records to return (default 20, max 200)
          where: Optional ModelCraft filter JSON, e.g. {"name": {"contains": "张"}}

      Returns:
          JSON string with items array and totalCount.
      """
      log = get_logger()
      args_summary = str({
          "db_name": db_name,
          "model_name": model_name,
          "fields": fields,
          "take": take,
          "where": where,
      })[:200]
      log.info("tool.call.start", tool_name="query_model", args_summary=args_summary)
      start = time.perf_counter()
      try:
          take = max(1, min(take, 200))
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
          duration_ms = round((time.perf_counter() - start) * 1000, 2)
          log.info("tool.call.end", tool_name="query_model", duration_ms=duration_ms, success=True)
          return json.dumps(data, ensure_ascii=False)
      except Exception:
          duration_ms = round((time.perf_counter() - start) * 1000, 2)
          log.exception("error", tool_name="query_model", duration_ms=duration_ms)
          log.info("tool.call.end", tool_name="query_model", duration_ms=duration_ms, success=False)
          raise
  ```

- [ ] **Step 3: 为 nl2filter tool 添加日志**

  将 `nl2filter` 函数体替换为：

  ```python
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
      log = get_logger()
      args_summary = str({"natural_language": natural_language, "field_names": field_names})[:200]
      log.info("tool.call.start", tool_name="nl2filter", args_summary=args_summary)
      start = time.perf_counter()
      try:
          llm = _get_llm()
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
          # Validate it parses as JSON before returning
          json.loads(raw)
          duration_ms = round((time.perf_counter() - start) * 1000, 2)
          log.info("tool.call.end", tool_name="nl2filter", duration_ms=duration_ms, success=True)
          return raw
      except Exception:
          duration_ms = round((time.perf_counter() - start) * 1000, 2)
          log.exception("error", tool_name="nl2filter", duration_ms=duration_ms)
          log.info("tool.call.end", tool_name="nl2filter", duration_ms=duration_ms, success=False)
          raise
  ```

- [ ] **Step 4: 为 _stream_graph 添加 thread_id binding 和 LLM 日志**

  将 `_stream_graph` 函数替换为：

  ```python
  async def _stream_graph(thread_id: str, state: dict, lc_messages: list) -> AsyncGenerator[str, None]:
      """Run the LangGraph and yield newline-delimited JSON events."""
      # Bind thread_id so all downstream logs (tools, graphql) carry it.
      bind_contextvars(thread_id=thread_id)

      log = get_logger()
      initial = {
          **state,
          "messages": lc_messages,
      }
      config_run = {"configurable": {"thread_id": thread_id}}

      # Track per-run LLM start times for accurate duration measurement.
      llm_start_times: dict[str, float] = {}

      async for event in modelcraft_graph.astream_events(initial, config=config_run, version="v2"):
          kind = event.get("event", "")
          run_id = event.get("run_id", "")

          if kind == "on_chat_model_start":
              llm_start_times[run_id] = time.perf_counter()
              log.info("llm.call.start", model=config.LLM_MODEL)

          elif kind == "on_chat_model_stream":
              chunk = event.get("data", {}).get("chunk", {})
              content = getattr(chunk, "content", "") or ""
              if content:
                  yield json.dumps({"type": "text", "content": content}) + "\n"

          elif kind == "on_chat_model_end":
              start_t = llm_start_times.pop(run_id, None)
              duration_ms = round(
                  (time.perf_counter() - (start_t if start_t is not None else time.perf_counter())) * 1000, 2
              )
              output = event.get("data", {}).get("output")
              usage = getattr(output, "usage_metadata", None) or {}
              log.info(
                  "llm.call.end",
                  model=config.LLM_MODEL,
                  duration_ms=duration_ms,
                  input_tokens=usage.get("input_tokens", 0),
                  output_tokens=usage.get("output_tokens", 0),
              )

          elif kind == "on_chain_end" and event.get("name") == "LangGraph":
              output = event.get("data", {}).get("output", {})
              messages_out = output.get("messages", [])
              last = messages_out[-1] if messages_out else None
              if last and hasattr(last, "content") and last.content:
                  yield json.dumps({"type": "state_snapshot", "state": {}}) + "\n"
  ```

- [ ] **Step 5: 运行全套测试，确认无回归**

  ```bash
  cd modelcraft-agent && python -m pytest -v
  ```

  期望：所有 tests PASSED（test_logging_setup、test_middleware、test_graphql_client）

- [ ] **Step 6: Commit**

  ```bash
  git add modelcraft-agent/agent.py
  git commit -m "feat(agent): add LLM and tool observability logging to agent"
  ```

---

## Task 7: 全量验证（端到端 smoke test）

**Files:** 无代码变更，仅验证

- [ ] **Step 1: 启动服务，验证日志格式**

  ```bash
  cd modelcraft-agent && python -m uvicorn main:app --port 8001 2>&1 &
  sleep 2

  # 发送带 requestId 的请求
  curl -s -X GET http://localhost:8001/healthz \
    -H "X-Request-ID: smoke-test-001" \
    -H "X-Client-Request-Id: fe-client-abc"

  kill %1
  ```

  期望：日志输出中包含如下结构（每行一个 JSON）：
  ```json
  {"service": "modelcraft-agent", "level": "info", "event": "request.start", "request_id": "smoke-test-001", "client_request_id": "fe-client-abc", "method": "GET", "path": "/healthz", "timestamp": "..."}
  {"service": "modelcraft-agent", "level": "info", "event": "request.end", "request_id": "smoke-test-001", "client_request_id": "fe-client-abc", "status_code": 200, "duration_ms": ..., "timestamp": "..."}
  ```

- [ ] **Step 2: 验证无 X-Request-ID 时自动生成 fallback UUID**

  ```bash
  cd modelcraft-agent && python -m uvicorn main:app --port 8001 2>&1 &
  sleep 2

  curl -s http://localhost:8001/healthz

  kill %1
  ```

  期望：`request.start` 日志中 `request_id` 字段为有效 UUID（非空，格式为 `xxxxxxxx-xxxx-xxxx-xxxx-xxxxxxxxxxxx`）

- [ ] **Step 3: Final commit — 完整功能提交**

  ```bash
  git add -A
  git commit -m "feat(agent): complete observability — structured JSON logging with requestId tracing"
  ```

---

## Self-Review

**Spec coverage check:**

| Spec 要求 | 覆盖 Task |
|-----------|----------|
| HTTP 请求生命周期（开始/结束/耗时/状态码）| Task 3 (middleware) + Task 4 (main.py) |
| LLM 调用（model、token、耗时）| Task 6 (agent.py _stream_graph) |
| Tool 调用（入参摘要、耗时、success）| Task 6 (query_model, nl2filter) |
| GraphQL 下游调用（URL、耗时、errors、status）| Task 5 (graphql_client) |
| 异常（exc_type、message、traceback）| Task 3 middleware + Task 5 + Task 6 |
| X-Request-ID 从 gateway 读取 | Task 3 |
| X-Client-Request-Id 从前端读取（可选）| Task 3 |
| structlog JSON 格式 | Task 2 (logging_setup) |
| structlog bind_contextvars 跨层传播 | Task 3 (bind) + Task 6 (thread_id bind) |

**Placeholder scan:** ✅ 无 TBD / TODO

**Type consistency check:**
- `get_logger()` 在 Task 2 定义，Task 5 和 Task 6 一致调用
- `capture_logs_with_context()` 在 Task 2 conftest 定义，Task 3 和 Task 5 测试一致使用
- `bind_contextvars` 仅在 `middleware.py` 和 `_stream_graph` 中调用，两处 import 路径一致
- `ObservabilityMiddleware` 在 Task 3 定义，Task 4 中 `from middleware import ObservabilityMiddleware` 路径一致

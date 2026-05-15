# model-agent 可观测性设计

**日期**: 2026-05-15  
**状态**: 待实现  
**关联服务**: `modelcraft-agent/`（Python FastAPI + LangGraph）

---

## 背景

`modelcraft-agent` 是基于 FastAPI + LangGraph 的 Python 服务，当前完全无日志，无法追踪请求链路、LLM 调用、工具执行和下游 GraphQL 请求，出现问题时无从排查。

---

## 目标

为 `modelcraft-agent` 添加结构化 JSON 日志 + requestId 追踪，覆盖：
1. 每个 HTTP 请求的生命周期（开始/结束/耗时/状态码）
2. LLM 调用（model、token 用量、耗时）
3. Tool 调用（`query_model` / `nl2filter`，入参摘要、耗时、成功与否）
4. GraphQL 下游调用（URL、耗时、HTTP 状态、是否有 errors）
5. 所有异常（exc_type、message、traceback）

---

## 方案选型

**采用方案一：FastAPI 中间件集中注入 + structlog bind_contextvars**

- FastAPI `@app.middleware("http")` 负责提取请求头、绑定 context、记录请求开始/结束
- `structlog.contextvars.bind_contextvars()` 将 requestId 注入进程级 ContextVar，所有层自动继承
- 各层手动埋点（约 15 行）
- 不引入 OpenTelemetry / LangChain Callback（当前基础设施不具备，over-engineering）

---

## 架构

### 新增文件

```
modelcraft-agent/
├── logging_setup.py       # structlog 初始化 + get_logger()
└── middleware.py          # ObservabilityMiddleware
```

### 修改文件

| 文件 | 变更内容 |
|------|----------|
| `main.py` | 注册 ObservabilityMiddleware；启动时调用 setup_logging() |
| `agent.py` | _stream_graph 加 LLM 调用日志；tool 函数加 tool.call.* 日志 |
| `client/graphql_client.py` | find_many 加 graphql.call.* 日志 |
| `requirements.txt` | 新增 `structlog` |

---

## requestId 传播

| 字段 | 来源 | 必填 |
|------|------|------|
| `request_id` | `X-Request-ID` 请求头（gateway 注入） | 否（缺失时 fallback 生成 UUID） |
| `client_request_id` | `X-Client-Request-Id` 请求头（前端可选） | 否 |

两个字段在中间件中通过 `bind_contextvars` 注入，后续所有日志自动携带。

---

## JSON 日志格式

### 固定字段（每条日志）

```json
{
  "timestamp": "2026-05-15T10:23:45.123Z",
  "level": "info",
  "event": "<event_name>",
  "request_id": "01HXZ...",
  "client_request_id": "fe-uuid-...",
  "thread_id": "thread-abc",
  "service": "modelcraft-agent"
}
```

### 事件字段规范

| event | 附加字段 |
|-------|----------|
| `request.start` | `method`, `path`, `client_ip` |
| `request.end` | `method`, `path`, `status_code`, `duration_ms` |
| `llm.call.start` | `model` |
| `llm.call.end` | `model`, `duration_ms`, `input_tokens`, `output_tokens` |
| `tool.call.start` | `tool_name`, `args_summary`（截断至 200 chars） |
| `tool.call.end` | `tool_name`, `duration_ms`, `success` |
| `graphql.call.start` | `url`, `operation` |
| `graphql.call.end` | `url`, `duration_ms`, `has_errors`, `status_code` |
| `error` | `exc_type`, `exc_message`, `traceback` |

> `args_summary` 对入参做 `str()[:200]` 截断，避免打出完整 JWT 或大数据集。

---

## 实现细节

### logging_setup.py

```python
import structlog

def setup_logging() -> None:
    structlog.configure(
        processors=[
            structlog.contextvars.merge_contextvars,
            structlog.processors.add_log_level,
            structlog.processors.TimeStamper(fmt="iso", utc=True),
            structlog.processors.JSONRenderer(),
        ],
        wrapper_class=structlog.make_filtering_bound_logger(logging.INFO),
        logger_factory=structlog.PrintLoggerFactory(),
    )

def get_logger():
    return structlog.get_logger().bind(service="modelcraft-agent")
```

### middleware.py

```python
import time, uuid
import structlog
from structlog.contextvars import bind_contextvars, clear_contextvars
from starlette.middleware.base import BaseHTTPMiddleware

class ObservabilityMiddleware(BaseHTTPMiddleware):
    async def dispatch(self, request, call_next):
        clear_contextvars()
        request_id = request.headers.get("X-Request-ID") or str(uuid.uuid4())
        client_req_id = request.headers.get("X-Client-Request-Id", "")
        bind_contextvars(request_id=request_id, client_request_id=client_req_id)

        log = structlog.get_logger()
        log.info("request.start", method=request.method, path=request.url.path,
                 client_ip=request.client.host if request.client else "")
        start = time.perf_counter()
        try:
            response = await call_next(request)
        except Exception:
            log.exception("error")
            raise
        finally:
            duration_ms = round((time.perf_counter() - start) * 1000, 2)
            log.info("request.end", method=request.method, path=request.url.path,
                     status_code=response.status_code, duration_ms=duration_ms)
        return response
```

### agent.py 埋点

- `_stream_graph` 入口：`log.info("llm.call.start", model=config.LLM_MODEL, thread_id=thread_id)`
- `on_chain_end` 事件处：记录 `llm.call.end`（含 duration_ms、token 用量从 event metadata 读取）
- `query_model` / `nl2filter` tool：入口记 `tool.call.start`，出口记 `tool.call.end`，异常记 `error`

### graphql_client.py 埋点

- `find_many` 开始：`log.info("graphql.call.start", url=url, operation="findMany")`
- 请求返回后：`log.info("graphql.call.end", url=url, duration_ms=..., has_errors=..., status_code=...)`
- `raise_for_status()` 异常：`log.exception("error")`

---

## 依赖变更

`requirements.txt` 新增：
```
structlog>=24.0.0
```

---

## 不在范围内

- OpenTelemetry / Jaeger 分布式 trace（基础设施未就绪）
- LangChain CallbackHandler（无法拿到 HTTP requestId）
- 日志采集 / 告警配置（ELK / Loki / Grafana）
- 请求体/响应体记录（隐私风险）

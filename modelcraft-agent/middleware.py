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

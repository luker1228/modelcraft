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

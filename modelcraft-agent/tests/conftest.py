"""Shared test utilities."""
from contextlib import contextmanager
from typing import Generator

import pytest
import structlog
from structlog.testing import LogCapture
from structlog.contextvars import clear_contextvars, merge_contextvars


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


@pytest.fixture(autouse=True)
def _reset_structlog():
    """Reset structlog to default state and clear contextvars between tests to prevent state leakage."""
    yield
    clear_contextvars()
    structlog.reset_defaults()
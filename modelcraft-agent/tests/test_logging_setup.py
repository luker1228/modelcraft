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

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
        # Disable caching so tests can reconfigure structlog's processor chain
        # via capture_logs() / capture_logs_with_context() without stale state.
        cache_logger_on_first_use=False,
    )


def get_logger() -> structlog.BoundLogger:
    """Return a structlog logger bound with service=modelcraft-agent."""
    return structlog.get_logger().bind(service="modelcraft-agent")

"""Privacy-safe usage/trace export hook and service logging.

The default export carries identifiers, configured/actual model identity,
token counts, latency and outcome only. It never carries prompts,
completions, retrieved content, envelope input or secrets. Content capture
requires constructing the hook with ``capture_content=True`` explicitly; there
is no implicit or environment-driven way to turn it on.

Vendor observability is never the usage ledger: core records durable usage by
joining the ``usage`` stream event to the envelope it issued.

Every line the process writes through ``logging`` is one JSON object with the
envelope ``time``, ``level``, ``service``, ``event`` and ``logger``:

* the service's own loggers (``intelligence``, ``intelligence.*``) use the
  record's message template as ``event`` (arguments are never interpolated)
  plus their explicit ``fields``;
* uvicorn records are mapped by template to fixed ``uvicorn.*`` events, with
  arguments dropped;
* any other library becomes ``library.log`` and warnings become
  ``python.warning``; their messages are never written.

Field keys that look sensitive are redacted, configured secret values are
scrubbed from every string, strings are capped at 512 characters, and
exceptions are reduced to their class name.
"""

from __future__ import annotations

import json
import logging
import math
import re
import sys
import warnings
from collections.abc import Callable, Iterable, Mapping, Sequence
from dataclasses import asdict
from datetime import UTC, datetime
from typing import Any

from providers import Message, UsageRecord

LOGGER_NAME = "intelligence"
EXPORT_LOGGER_NAME = "intelligence.observability"
SERVICE_NAME = "intelligence"

MAX_STRING_CHARS = 512
MAX_DEPTH = 16
REDACTED = "[REDACTED]"
SENSITIVE_KEY = re.compile(
    r"(?i)(password|passwd|secret|token|api_?key|authorization|cookie|dsn|credential)"
)
ERROR_TYPE = re.compile(r"[A-Za-z][A-Za-z0-9_.]{0,63}")
LOGGER_NAME_PATTERN = re.compile(r"[A-Za-z0-9_.-]{1,64}")
ENVELOPE_KEYS = frozenset({"time", "level", "service", "event", "logger"})

STARTUP_FAILED_EVENT = "intelligence.startup.failed"
LIBRARY_EVENT = "library.log"
WARNING_EVENT = "python.warning"
WARNINGS_LOGGER = "py.warnings"
NON_TEMPLATE_EVENT = "intelligence.message"

# uvicorn 0.54.0 message templates (``record.msg``), verified against the
# installed package. Arguments (hosts, ports, process IDs, values) are dropped.
UVICORN_EVENTS: dict[str, str] = {
    "Started server process [%d]": "uvicorn.process_started",
    "Waiting for application startup.": "uvicorn.startup_waiting",
    "Application startup complete.": "uvicorn.startup_complete",
    "Uvicorn running on %s://%s:%d (Press CTRL+C to quit)": "uvicorn.listening",
    "Uvicorn running on %s://[%s]:%d (Press CTRL+C to quit)": "uvicorn.listening",
    "Uvicorn running on socket %s (Press CTRL+C to quit)": "uvicorn.listening",
    "Uvicorn running on unix socket %s (Press CTRL+C to quit)": "uvicorn.listening",
    "Shutting down": "uvicorn.shutting_down",
    "Waiting for application shutdown.": "uvicorn.shutdown_waiting",
    "Waiting for connections to close. (CTRL+C to force quit)": "uvicorn.shutdown_waiting",
    "Waiting for background tasks to complete. (CTRL+C to force quit)": (
        "uvicorn.shutdown_waiting"
    ),
    "Application shutdown complete.": "uvicorn.shutdown_complete",
    "Finished server process [%d]": "uvicorn.process_finished",
    "Invalid HTTP request received.": "uvicorn.invalid_request",
    "Unsupported upgrade request.": "uvicorn.invalid_request",
    "Exception in ASGI application\n": "uvicorn.asgi_exception",
    "Exception in 'lifespan' protocol\n": "uvicorn.asgi_exception",
    "ASGI callable returned without starting response.": "uvicorn.server_error",
    "ASGI callable returned without completing response.": "uvicorn.server_error",
    "ASGI callable should return None, but returned '%s'.": "uvicorn.server_error",
    "Application startup failed. Exiting.": "uvicorn.server_error",
    "Application shutdown failed. Exiting.": "uvicorn.server_error",
    "Cancel %s running task(s), timeout graceful shutdown exceeded": "uvicorn.server_error",
}
# ``config.load_app`` formats the exception into the message itself.
UVICORN_PREFIX_EVENTS: tuple[tuple[str, str], ...] = (
    ("Error loading ASGI app.", "uvicorn.app_load_failed"),
)
UVICORN_EXCEPTION_EVENT = "uvicorn.server_error"  # e.g. ``logger.error(OSError(...))``
UVICORN_OTHER_EVENT = "uvicorn.other"

Sink = Callable[[Mapping[str, Any]], None]


def log_sink(record: Mapping[str, Any]) -> None:
    """Default sink: one structured log line per execution."""
    logging.getLogger(EXPORT_LOGGER_NAME).info("execution", extra={"fields": dict(record)})


class ObservabilityHook:
    """Builds and exports one trace record per execution."""

    def __init__(self, sink: Sink = log_sink, *, capture_content: bool = False) -> None:
        self._sink = sink
        self.capture_content = capture_content

    def record(
        self,
        *,
        request_id: str,
        trace_id: str,
        capability: str,
        outcome: str,
        usage: UsageRecord | None = None,
        messages: Sequence[Message] = (),
        completion: str | None = None,
    ) -> dict[str, Any]:
        """Export a record; return it (tests inspect what was exported)."""
        record: dict[str, Any] = {
            "request_id": request_id,
            "trace_id": trace_id,
            "capability": capability,
            "outcome": outcome,
        }
        if usage is not None:
            record["usage"] = asdict(usage)
        if self.capture_content:
            record["content"] = {
                "messages": [{"role": m.role, "content": m.content} for m in messages],
                "completion": completion,
            }
        self._sink(record)
        return record


def error_type_name(value: BaseException | type[BaseException] | None) -> str:
    """The class name of an exception (taken from its class, never its data)."""
    cls = value if isinstance(value, type) else type(value)
    name = getattr(cls, "__name__", "")
    return name if isinstance(name, str) and ERROR_TYPE.fullmatch(name) else "Exception"


def _level_name(levelno: int) -> str:
    if levelno >= logging.ERROR:
        return "error"
    if levelno >= logging.WARNING:
        return "warn"
    if levelno >= logging.INFO:
        return "info"
    return "debug"


def _is_own_logger(name: str) -> bool:
    return name == LOGGER_NAME or name.startswith(LOGGER_NAME + ".")


def _is_uvicorn_logger(name: str) -> bool:
    return name == "uvicorn" or name.startswith("uvicorn.")


def uvicorn_event(msg: object) -> str:
    """The fixed event for a uvicorn record's template (``record.msg``)."""
    if isinstance(msg, BaseException):
        return UVICORN_EXCEPTION_EVENT
    if isinstance(msg, str):
        event = UVICORN_EVENTS.get(msg)
        if event is not None:
            return event
        for prefix, prefixed_event in UVICORN_PREFIX_EVENTS:
            if msg.startswith(prefix):
                return prefixed_event
    return UVICORN_OTHER_EVENT


class JsonFormatter(logging.Formatter):
    """One bounded JSON object per record (see the module docstring)."""

    def __init__(self, secrets: Iterable[str] = ()) -> None:
        super().__init__()
        self._secrets: tuple[str, ...] = ()
        self.set_secrets(secrets)

    @property
    def secrets(self) -> tuple[str, ...]:
        return self._secrets

    def set_secrets(self, secrets: Iterable[str]) -> None:
        # Longest first, so overlapping values are removed completely.
        unique = {s for s in secrets if isinstance(s, str) and s}
        self._secrets = tuple(sorted(unique, key=len, reverse=True))

    def add_secrets(self, secrets: Iterable[str]) -> None:
        self.set_secrets((*self._secrets, *secrets))

    # -- value sanitation ---------------------------------------------------

    def _text(self, value: str) -> str:
        for secret in self._secrets:
            if secret in value:
                value = value.replace(secret, REDACTED)
        return value[:MAX_STRING_CHARS]

    def _value(self, value: Any, depth: int = 0) -> Any:
        if value is None or isinstance(value, bool | int):
            return value
        if isinstance(value, float):
            return value if math.isfinite(value) else self._text(str(value))
        if isinstance(value, str):
            return self._text(value)
        if depth >= MAX_DEPTH:
            return REDACTED
        if isinstance(value, Mapping):
            return self._mapping(value, depth + 1)
        if isinstance(value, list | tuple | set | frozenset):
            return [self._value(item, depth + 1) for item in value]
        return self._text(str(value))

    def _mapping(self, fields: Mapping[Any, Any], depth: int = 0) -> dict[str, Any]:
        out: dict[str, Any] = {}
        for raw_key, value in fields.items():
            key = self._text(str(raw_key))
            out[key] = REDACTED if SENSITIVE_KEY.search(key) else self._value(value, depth)
        return out

    # -- record mapping -------------------------------------------------------

    def _envelope(self, record: logging.LogRecord, event: str) -> dict[str, Any]:
        created = datetime.fromtimestamp(record.created, UTC)
        name = record.name if LOGGER_NAME_PATTERN.fullmatch(record.name) else "unknown"
        return {
            "time": created.strftime("%Y-%m-%dT%H:%M:%S.") + f"{created.microsecond // 1000:03d}Z",
            "level": _level_name(record.levelno),
            "service": SERVICE_NAME,
            "event": self._text(event),
            "logger": name,
        }

    @staticmethod
    def _exc_type(record: logging.LogRecord) -> str | None:
        if record.exc_info and record.exc_info[0] is not None:
            return error_type_name(record.exc_info[0])
        return None

    def format(self, record: logging.LogRecord) -> str:
        name = record.name
        if _is_own_logger(name):
            event = record.msg if isinstance(record.msg, str) else NON_TEMPLATE_EVENT
            payload = self._envelope(record, event)
            fields = getattr(record, "fields", None)
            if isinstance(fields, Mapping):
                for key, value in self._mapping(fields).items():
                    if key not in ENVELOPE_KEYS:
                        payload[key] = value
            exc_type = self._exc_type(record)
            if exc_type is not None and "error_type" not in payload:
                payload["error_type"] = exc_type
        elif _is_uvicorn_logger(name):
            payload = self._envelope(record, uvicorn_event(record.msg))
            exc_type = self._exc_type(record)
            if exc_type is None and isinstance(record.msg, BaseException):
                exc_type = error_type_name(record.msg)
            if exc_type is not None:
                payload["error_type"] = exc_type
        elif name == WARNINGS_LOGGER:
            payload = self._envelope(record, WARNING_EVENT)
            category = getattr(record, "warning_category", None)
            valid = isinstance(category, str) and ERROR_TYPE.fullmatch(category)
            payload["category"] = category if valid else "Warning"
        else:
            payload = self._envelope(record, LIBRARY_EVENT)
            exc_type = self._exc_type(record)
            if exc_type is not None:
                payload["error_type"] = exc_type
        return json.dumps(payload, ensure_ascii=True, separators=(",", ":"))


class _StdoutHandler(logging.StreamHandler):
    """Writes to whatever ``sys.stdout`` is at emit time (survives capture swaps)."""

    _intelligence = True

    def __init__(self) -> None:
        super().__init__(sys.stdout)

    @property  # type: ignore[override]
    def stream(self) -> Any:
        return sys.stdout

    @stream.setter
    def stream(self, value: Any) -> None:  # the stream is always the current stdout
        pass

    def handleError(self, record: logging.LogRecord) -> None:
        """Drop a record that cannot be formatted or written: never fall back to raw text (§5.4)."""
        return


def _structured_showwarning(
    message: Warning | str,
    category: type[Warning],
    filename: str,
    lineno: int,
    file: Any = None,
    line: str | None = None,
) -> None:
    """``warnings.showwarning`` replacement: category only, never the text."""
    if file is not None:
        warnings._showwarning_orig(message, category, filename, lineno, file, line)  # type: ignore[attr-defined]
        return
    logging.getLogger(WARNINGS_LOGGER).warning(
        WARNING_EVENT, extra={"warning_category": error_type_name(category)}
    )


def _json_handler() -> _StdoutHandler | None:
    for handler in logging.getLogger().handlers:
        if getattr(handler, "_intelligence", False) and isinstance(handler, _StdoutHandler):
            return handler
    return None


def logging_configured() -> bool:
    """True once ``configure_logging`` has installed the JSON handler."""
    return _json_handler() is not None


def configure_logging(level: int = logging.INFO, secrets: Iterable[str] = ()) -> None:
    """Route every log record to stdout as one JSON line. Idempotent.

    Installs a single JSON handler on the root logger (root at WARNING, the
    service's loggers at ``level``), takes over uvicorn's loggers (the
    ``uvicorn`` handlers are replaced, ``uvicorn.access`` is disabled) and
    captures warnings. ``secrets`` are added to the values scrubbed from every
    string; calling again with further secrets updates the existing handler.
    """
    root = logging.getLogger()
    handler = _json_handler()
    if handler is None:
        handler = _StdoutHandler()
        handler.setFormatter(JsonFormatter())
        root.addHandler(handler)
    formatter = handler.formatter
    if not isinstance(formatter, JsonFormatter):
        formatter = JsonFormatter()
        handler.setFormatter(formatter)
    formatter.add_secrets(secrets)
    root.setLevel(logging.WARNING)

    service = logging.getLogger(LOGGER_NAME)
    service.setLevel(level)
    for old in list(service.handlers):  # records propagate to the root handler
        if getattr(old, "_intelligence", False):
            service.removeHandler(old)

    framework = logging.getLogger("uvicorn")
    framework.handlers = [handler]
    framework.propagate = False
    if framework.level == logging.NOTSET:
        framework.setLevel(logging.INFO)
    logging.getLogger("uvicorn.error").handlers = []
    access = logging.getLogger("uvicorn.access")
    access.handlers = []
    access.propagate = False
    access.disabled = True

    logging.captureWarnings(True)
    warnings.showwarning = _structured_showwarning

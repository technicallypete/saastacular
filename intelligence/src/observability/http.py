"""Diagnostic context and one completion record per health request.

``CompletionLogMiddleware`` is a pure ASGI middleware scoped to exactly
``/healthz`` and ``/readyz``. It derives the request's diagnostic context from
the raw request headers (counting duplicated fields), adds the effective
``X-Request-Id`` to the response and writes exactly one
``http.request.completed`` record when the request ends: completed,
cancelled or failed. It never logs headers, query strings, bodies or
exception text, and keeps all state in local variables of the call.

Context rules:

* ``X-Request-Id`` is adopted only when exactly one field is present, it fully
  matches ``^[A-Za-z0-9._-]{8,64}$`` and it contains none of the service's
  configured secrets; otherwise 32 lowercase hex characters are generated.
* ``traceparent`` is adopted only when exactly one field of at most 512 bytes
  fully matches ``00-<32 hex>-<16 hex>-<2 hex>`` (lowercase) with nonzero
  trace and parent IDs; otherwise a new trace is started. Only the sampled
  flag is kept. ``tracestate`` and ``baggage`` are never read.
* Every request gets a fresh nonzero span ID; the inbound parent ID is
  recorded as ``parent_span_id`` only when ``traceparent`` was valid.
"""

from __future__ import annotations

import asyncio
import logging
import re
import secrets as _secrets
import time
from collections.abc import Awaitable, Callable, Iterable, MutableMapping
from dataclasses import dataclass
from typing import Any

from observability.export import LOGGER_NAME, error_type_name

HTTP_LOGGER_NAME = f"{LOGGER_NAME}.http"
COMPLETION_EVENT = "http.request.completed"
SCOPED_ROUTES = frozenset({"/healthz", "/readyz"})
METHODS = frozenset({"GET", "HEAD", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"})

REQUEST_ID_HEADER = b"x-request-id"
TRACEPARENT_HEADER = b"traceparent"
REQUEST_ID_PATTERN = re.compile(r"[A-Za-z0-9._-]{8,64}")
TRACEPARENT_PATTERN = re.compile(r"00-([0-9a-f]{32})-([0-9a-f]{16})-([0-9a-f]{2})")
MAX_TRACEPARENT_BYTES = 512

Scope = MutableMapping[str, Any]
Message = MutableMapping[str, Any]
Receive = Callable[[], Awaitable[Message]]
Send = Callable[[Message], Awaitable[None]]
ASGIApp = Callable[[Scope, Receive, Send], Awaitable[None]]


@dataclass(frozen=True, slots=True)
class DiagnosticContext:
    request_id: str
    request_id_source: str
    trace_id: str
    trace_source: str
    span_id: str
    parent_span_id: str | None
    trace_flags: str


def _random_hex(nbytes: int) -> str:
    """Random lowercase hex that is never all zeros."""
    while True:
        value = _secrets.token_hex(nbytes)
        if value.strip("0"):
            return value


def _field_values(headers: Iterable[tuple[bytes, bytes]], name: bytes) -> list[bytes]:
    return [value for key, value in headers if bytes(key).lower() == name]


def _inbound_request_id(values: list[bytes], secrets: tuple[str, ...]) -> str | None:
    if len(values) != 1:
        return None
    try:
        candidate = bytes(values[0]).decode("ascii")
    except UnicodeDecodeError:
        return None
    if REQUEST_ID_PATTERN.fullmatch(candidate) is None:
        return None
    if any(secret in candidate for secret in secrets):
        return None
    return candidate


def _inbound_traceparent(values: list[bytes]) -> tuple[str, str, str] | None:
    if len(values) != 1 or len(values[0]) > MAX_TRACEPARENT_BYTES:
        return None
    try:
        candidate = bytes(values[0]).decode("ascii")
    except UnicodeDecodeError:
        return None
    match = TRACEPARENT_PATTERN.fullmatch(candidate)
    if match is None:
        return None
    trace_id, parent_id, flags = match.groups()
    if not trace_id.strip("0") or not parent_id.strip("0"):
        return None
    return trace_id, parent_id, "01" if int(flags, 16) & 0x01 else "00"


def derive_context(
    headers: Iterable[tuple[bytes, bytes]], secrets: Iterable[str] = ()
) -> DiagnosticContext:
    """The diagnostic context for one request, from its raw header list."""
    raw = list(headers)
    secret_values = tuple(s for s in secrets if s)
    request_id = _inbound_request_id(_field_values(raw, REQUEST_ID_HEADER), secret_values)
    traceparent = _inbound_traceparent(_field_values(raw, TRACEPARENT_HEADER))
    if traceparent is None:
        trace_id, parent_span_id, flags, trace_source = _random_hex(16), None, "00", "generated"
    else:
        trace_id, parent_span_id, flags = traceparent
        trace_source = "inbound"
    return DiagnosticContext(
        request_id=request_id or _random_hex(16),
        request_id_source="generated" if request_id is None else "inbound",
        trace_id=trace_id,
        trace_source=trace_source,
        span_id=_random_hex(8),
        parent_span_id=parent_span_id,
        trace_flags=flags,
    )


class CompletionLogMiddleware:
    """Pure ASGI middleware: context, ``X-Request-Id`` and one completion record."""

    def __init__(self, app: ASGIApp, secrets: Iterable[str] = ()) -> None:
        self.app = app
        self._secrets = tuple(s for s in secrets if s)
        self._logger = logging.getLogger(HTTP_LOGGER_NAME)

    async def __call__(self, scope: Scope, receive: Receive, send: Send) -> None:
        if scope.get("type") != "http" or scope.get("path") not in SCOPED_ROUTES:
            await self.app(scope, receive, send)
            return

        started = time.monotonic()
        context = derive_context(scope.get("headers") or (), self._secrets)
        response_id = (REQUEST_ID_HEADER, context.request_id.encode("ascii"))
        status = 0
        outcome = "completed"
        error_type: str | None = None

        async def send_with_context(message: Message) -> None:
            nonlocal status
            if message.get("type") == "http.response.start" and status == 0:
                status = int(message.get("status", 0))
                headers = [
                    (key, value)
                    for key, value in message.get("headers") or ()
                    if bytes(key).lower() != REQUEST_ID_HEADER
                ]
                message = {**message, "headers": [*headers, response_id]}
            await send(message)

        try:
            await self.app(scope, receive, send_with_context)
        except asyncio.CancelledError:
            outcome = "cancelled"
            raise
        except Exception as exc:
            outcome = "failed"
            error_type = error_type_name(exc)
            # FastAPI's outermost ServerErrorMiddleware answers an exception raised before the
            # response started with a 500, so that is the final status (§5.2).
            status = status or 500
            raise
        except BaseException as exc:
            outcome = "failed"
            error_type = error_type_name(exc)
            raise
        finally:
            self._log(scope, context, status, started, outcome, error_type)

    def _log(
        self,
        scope: Scope,
        context: DiagnosticContext,
        status: int,
        started: float,
        outcome: str,
        error_type: str | None,
    ) -> None:
        method = scope.get("method")
        fields: dict[str, Any] = {
            "method": method if method in METHODS else "OTHER",
            "route": scope["path"],
            "status": status,
            "duration_ms": max(0, int((time.monotonic() - started) * 1000)),
            "request_id": context.request_id,
            "request_id_source": context.request_id_source,
            "trace_id": context.trace_id,
            "trace_source": context.trace_source,
            "span_id": context.span_id,
        }
        if context.parent_span_id is not None:
            fields["parent_span_id"] = context.parent_span_id
        fields["trace_flags"] = context.trace_flags
        fields["outcome"] = outcome
        if error_type is not None:
            fields["error_type"] = error_type
        if outcome == "failed":
            level = logging.ERROR
        elif status >= 500:
            level = logging.WARNING
        else:
            level = logging.INFO
        self._logger.log(level, COMPLETION_EVENT, extra={"fields": fields})

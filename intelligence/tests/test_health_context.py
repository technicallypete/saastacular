"""I-HEALTH-CONTEXT: health headers, diagnostic context and completion records."""

from __future__ import annotations

import asyncio
import logging
import re
from typing import Any

import httpx
import pytest

from conftest import SERVICE_TOKEN, base_env
from main import create_app
from observability.http import (
    COMPLETION_EVENT,
    HTTP_LOGGER_NAME,
    CompletionLogMiddleware,
    derive_context,
)

HEX32 = re.compile(r"[0-9a-f]{32}")
HEX16 = re.compile(r"[0-9a-f]{16}")
TRACE = "4bf92f3577b34da6a3ce929d0e0e4736"
PARENT = "00f067aa0ba902b7"
VALID_TP = f"00-{TRACE}-{PARENT}-01"
OPENROUTER_VALUE = "synthetic-openrouter-value-for-tests-0001"
COMPLETION_KEYS = {
    "method",
    "route",
    "status",
    "duration_ms",
    "request_id",
    "request_id_source",
    "trace_id",
    "trace_source",
    "span_id",
    "parent_span_id",
    "trace_flags",
    "outcome",
    "error_type",
}


def _app(**overrides: str):
    return create_app(base_env(OPENROUTER_API_KEY=OPENROUTER_VALUE, **overrides))


async def _get(app, path: str, headers: Any = None) -> httpx.Response:
    transport = httpx.ASGITransport(app=app)
    async with httpx.AsyncClient(transport=transport, base_url="http://intelligence") as client:
        return await client.get(path, headers=headers)


def get(app, path: str, headers: Any = None) -> httpx.Response:
    return asyncio.run(_get(app, path, headers))


def completion_records(caplog) -> list[dict[str, Any]]:
    return [
        dict(r.fields)
        for r in caplog.records
        if r.name == HTTP_LOGGER_NAME and r.getMessage() == COMPLETION_EVENT
    ]


@pytest.fixture
def records(caplog):
    caplog.set_level(logging.INFO, logger=HTTP_LOGGER_NAME)
    return caplog


# ---------------------------------------------------------------- health routes


@pytest.mark.parametrize(
    ("path", "body"), [("/healthz", b'{"status":"ok"}'), ("/readyz", b'{"status":"ready"}')]
)
def test_i_health_bodies_unchanged_with_no_store_and_request_id(path, body, records) -> None:
    response = get(_app(), path)
    assert response.status_code == 200
    assert response.content == body  # byte-identical to the baseline literals
    assert response.headers["content-type"] == "application/json"
    assert response.headers["cache-control"] == "no-store"
    assert HEX32.fullmatch(response.headers["x-request-id"])
    assert len(response.headers.get_list("x-request-id")) == 1
    [record] = completion_records(records)
    assert record["route"] == path
    assert record["status"] == 200
    assert record["request_id"] == response.headers["x-request-id"]


def test_i_health_ready_without_any_configured_model(records) -> None:
    app = create_app(base_env())
    assert app.state.settings.resolver.configured == ()
    response = get(app, "/readyz")
    assert (response.status_code, response.content) == (200, b'{"status":"ready"}')


def test_i_health_valid_inbound_context_is_adopted(records) -> None:
    response = get(_app(), "/healthz", {"X-Request-Id": "req-abcdef12", "traceparent": VALID_TP})
    assert response.headers["x-request-id"] == "req-abcdef12"
    [record] = completion_records(records)
    assert set(record) <= COMPLETION_KEYS
    assert record["request_id"] == "req-abcdef12"
    assert record["request_id_source"] == "inbound"
    assert (record["trace_id"], record["trace_source"]) == (TRACE, "inbound")
    assert record["parent_span_id"] == PARENT
    assert HEX16.fullmatch(record["span_id"]) and record["span_id"] != PARENT
    assert record["trace_flags"] == "01"
    assert record["method"] == "GET"
    assert record["outcome"] == "completed"
    assert "error_type" not in record
    assert isinstance(record["duration_ms"], int) and record["duration_ms"] >= 0


# ---------------------------------------------------------------- context table

SECRET_CASES = {"token-exact": SERVICE_TOKEN, "openrouter-embedded": "x" + OPENROUTER_VALUE}

REQUEST_ID_ROWS = [
    ("valid", [(b"x-request-id", b"req-abcdef12")], "req-abcdef12"),
    ("valid-64", [(b"x-request-id", b"a" * 64)], "a" * 64),
    ("absent", [], None),
    ("malformed", [(b"x-request-id", b"req abcdef12")], None),
    ("too-short", [(b"x-request-id", b"req-123")], None),
    ("oversized-65", [(b"x-request-id", b"a" * 65)], None),
    ("oversized-8k", [(b"x-request-id", b"a" * 8192)], None),
    ("trailing-newline", [(b"x-request-id", b"req-abcdef12\n")], None),
    ("duplicated", [(b"x-request-id", b"req-abcdef12"), (b"x-request-id", b"req-abcdef12")], None),
    ("comma-joined", [(b"x-request-id", b"req-abcdef12, req-abcdef13")], None),
    ("non-ascii", [(b"x-request-id", "req-abcdéf12".encode())], None),
    ("secret-token", [(b"x-request-id", SERVICE_TOKEN.encode())], None),
    ("secret-embedded", [(b"x-request-id", SECRET_CASES["openrouter-embedded"].encode())], None),
]


@pytest.mark.parametrize(
    ("headers", "expected"),
    [(h, e) for _, h, e in REQUEST_ID_ROWS],
    ids=[name for name, _, _ in REQUEST_ID_ROWS],
)
def test_i_context_request_id_table(headers, expected) -> None:
    context = derive_context(headers, (SERVICE_TOKEN, OPENROUTER_VALUE))
    if expected is None:
        assert HEX32.fullmatch(context.request_id)
        assert context.request_id_source == "generated"
    else:
        assert (context.request_id, context.request_id_source) == (expected, "inbound")


TP = "traceparent"
TRACEPARENT_ROWS = [
    ("valid-sampled", [VALID_TP], (TRACE, PARENT, "01")),
    ("valid-unsampled", [f"00-{TRACE}-{PARENT}-00"], (TRACE, PARENT, "00")),
    ("unknown-flag-bits-dropped", [f"00-{TRACE}-{PARENT}-fe"], (TRACE, PARENT, "00")),
    ("sampled-plus-unknown-bits", [f"00-{TRACE}-{PARENT}-ff"], (TRACE, PARENT, "01")),
    ("absent", [], None),
    ("malformed", ["not-a-traceparent"], None),
    ("all-zero-trace", [f"00-{'0' * 32}-{PARENT}-01"], None),
    ("all-zero-parent", [f"00-{TRACE}-{'0' * 16}-01"], None),
    ("uppercase", [VALID_TP.upper()], None),
    ("version-01", [f"01-{TRACE}-{PARENT}-01"], None),
    ("version-ff", [f"ff-{TRACE}-{PARENT}-01"], None),
    ("extra-fields", [f"{VALID_TP}-extra"], None),
    ("internal-whitespace", [f"00-{TRACE} -{PARENT}-01"], None),
    ("trailing-newline", [VALID_TP + "\n"], None),
    ("short-trace", [f"00-{TRACE[:31]}-{PARENT}-01"], None),
    ("duplicated", [VALID_TP, VALID_TP], None),
    ("comma-joined", [f"{VALID_TP},{VALID_TP}"], None),
    ("oversized-513", ["0" * 513], None),
    ("oversized-8k", ["a" * 8192], None),
]


@pytest.mark.parametrize(
    ("values", "expected"),
    [(v, e) for _, v, e in TRACEPARENT_ROWS],
    ids=[name for name, _, _ in TRACEPARENT_ROWS],
)
def test_i_context_traceparent_table(values, expected) -> None:
    headers = [(TP.encode(), v.encode()) for v in values]
    headers.append((b"tracestate", b"vendor=ignored"))
    headers.append((b"baggage", b"key=ignored"))
    context = derive_context(headers)
    assert HEX16.fullmatch(context.span_id) and context.span_id.strip("0")
    if expected is None:
        assert HEX32.fullmatch(context.trace_id) and context.trace_id.strip("0")
        assert context.trace_source == "generated"
        assert context.parent_span_id is None
        assert context.trace_flags == "00"
    else:
        trace, parent, flags = expected
        assert (context.trace_id, context.parent_span_id, context.trace_flags) == (
            trace,
            parent,
            flags,
        )
        assert context.trace_source == "inbound"
        assert context.span_id != parent


def test_i_context_header_names_are_case_insensitive() -> None:
    context = derive_context(
        [(b"X-Request-ID", b"req-abcdef12"), (b"TraceParent", VALID_TP.encode())]
    )
    assert context.request_id == "req-abcdef12"
    assert context.trace_id == TRACE


# ----------------------------------------------------- invalid values via HTTP


@pytest.mark.parametrize(
    "headers",
    [
        [("X-Request-Id", "req-abcdef12"), ("X-Request-Id", "req-abcdef13")],
        [("X-Request-Id", "req bad value")],
        [("X-Request-Id", SERVICE_TOKEN)],
        [("X-Request-Id", OPENROUTER_VALUE)],
        [("traceparent", VALID_TP), ("traceparent", VALID_TP)],
        [("traceparent", VALID_TP.upper())],
    ],
    ids=["dup-id", "bad-id", "token-id", "openrouter-id", "dup-tp", "upper-tp"],
)
def test_i_health_invalid_values_never_reflected(headers, records, capsys) -> None:
    response = get(_app(), "/healthz", headers)
    assert response.status_code == 200
    generated = response.headers["x-request-id"]
    assert HEX32.fullmatch(generated)
    [record] = completion_records(records)
    assert record["request_id"] == generated
    assert record["request_id_source"] == "generated" or record["trace_source"] == "generated"
    output = capsys.readouterr().out
    assert COMPLETION_EVENT in output  # the record was written, without the values
    for _, value in headers:
        assert value not in output
        assert value not in generated
    if any(name == "traceparent" for name, _ in headers):
        assert TRACE not in output and PARENT not in output
    for secret in (SERVICE_TOKEN, OPENROUTER_VALUE):
        assert secret not in output


# ------------------------------------------------------------- scope and counts


def test_i_health_non_health_paths_produce_no_completion_record(records) -> None:
    app = _app()
    assert get(app, "/internal/v1/execute").status_code == 401
    assert get(app, "/nope").status_code == 404
    assert get(app, "/healthz/").status_code in (307, 404)
    assert completion_records(records) == []


def test_i_health_head_and_other_methods_are_bounded(records) -> None:
    app = _app()

    async def run() -> list[httpx.Response]:
        transport = httpx.ASGITransport(app=app)
        async with httpx.AsyncClient(transport=transport, base_url="http://i") as client:
            return [
                await client.head("/healthz"),
                await client.request("PROPFIND", "/readyz"),
            ]

    head, other = asyncio.run(run())
    records_ = completion_records(records)
    assert [r["method"] for r in records_] == ["HEAD", "OTHER"]
    assert [r["status"] for r in records_] == [head.status_code, other.status_code]


def test_i_health_100_concurrent_requests_are_isolated(records) -> None:
    app = _app()

    async def run() -> list[httpx.Response]:
        transport = httpx.ASGITransport(app=app)
        async with httpx.AsyncClient(transport=transport, base_url="http://i") as client:
            return await asyncio.gather(
                *(
                    client.get(
                        "/healthz" if n % 2 else "/readyz",
                        headers={
                            "X-Request-Id": f"req-conc-{n:04d}",
                            "traceparent": f"00-{n + 1:032x}-{n + 1:016x}-01",
                        },
                    )
                    for n in range(100)
                )
            )

    responses = asyncio.run(run())
    assert [r.headers["x-request-id"] for r in responses] == [
        f"req-conc-{n:04d}" for n in range(100)
    ]
    by_id = {r["request_id"]: r for r in completion_records(records)}
    assert len(by_id) == 100
    for n in range(100):
        record = by_id[f"req-conc-{n:04d}"]
        assert record["trace_id"] == f"{n + 1:032x}"
        assert record["parent_span_id"] == f"{n + 1:016x}"
        assert record["route"] == ("/healthz" if n % 2 else "/readyz")
    assert len({r["span_id"] for r in by_id.values()}) == 100


# ------------------------------------------------- failure and cancellation


def _scope(path: str = "/healthz", headers: list[tuple[bytes, bytes]] | None = None):
    return {"type": "http", "method": "GET", "path": path, "headers": headers or []}


async def _receive() -> dict[str, Any]:
    return {"type": "http.request", "body": b"", "more_body": False}


class SyntheticFailure(RuntimeError):
    pass


def test_i_health_exception_is_logged_once_and_reraised(records) -> None:
    sent: list[dict[str, Any]] = []

    async def failing_app(scope, receive, send):
        raise SyntheticFailure("failure text CANARY-exception-7d1e")

    async def send(message):
        sent.append(message)

    middleware = CompletionLogMiddleware(failing_app)
    with pytest.raises(SyntheticFailure):
        asyncio.run(middleware(_scope(), _receive, send))
    [record] = completion_records(records)
    assert (record["outcome"], record["error_type"], record["status"]) == (
        "failed",
        "SyntheticFailure",
        500,
    )
    [log] = [r for r in records.records if r.name == HTTP_LOGGER_NAME]
    assert log.levelno == logging.ERROR
    assert "CANARY" not in repr(record)


def test_i_health_exception_after_response_start_keeps_first_status(records) -> None:
    async def app(scope, receive, send):
        await send({"type": "http.response.start", "status": 503, "headers": []})
        await send({"type": "http.response.start", "status": 200, "headers": []})
        raise SyntheticFailure

    sent: list[dict[str, Any]] = []

    async def send(message):
        sent.append(message)

    with pytest.raises(SyntheticFailure):
        asyncio.run(CompletionLogMiddleware(app)(_scope(), _receive, send))
    [record] = completion_records(records)
    assert record["status"] == 503
    assert record["outcome"] == "failed"
    # The second start is forwarded unchanged; only the first gains the ID.
    assert [b"x-request-id" in dict(m["headers"]) for m in sent] == [True, False]


def test_i_health_cancellation_is_logged_once_and_reraised(records) -> None:
    started = asyncio.Event()

    async def slow_app(scope, receive, send):
        started.set()
        await asyncio.sleep(30)

    async def send(message):
        pass

    async def run() -> None:
        task = asyncio.ensure_future(CompletionLogMiddleware(slow_app)(_scope(), _receive, send))
        await started.wait()
        task.cancel()
        with pytest.raises(asyncio.CancelledError):
            await task

    asyncio.run(run())
    [record] = completion_records(records)
    assert (record["outcome"], record["status"]) == ("cancelled", 0)
    assert "error_type" not in record


def test_i_health_server_error_status_is_warn(records) -> None:
    async def app(scope, receive, send):
        await send(
            {"type": "http.response.start", "status": 503, "headers": [(b"x-request-id", b"x")]}
        )
        await send({"type": "http.response.body", "body": b"{}"})

    sent: list[dict[str, Any]] = []

    async def send(message):
        sent.append(message)

    asyncio.run(CompletionLogMiddleware(app)(_scope("/readyz"), _receive, send))
    [log] = [r for r in records.records if r.name == HTTP_LOGGER_NAME]
    assert log.levelno == logging.WARNING
    ids = [v for k, v in sent[0]["headers"] if k == b"x-request-id"]
    assert len(ids) == 1 and ids[0] != b"x"


def test_i_health_state_does_not_leak_between_requests(records) -> None:
    app = _app()
    get(app, "/healthz", {"X-Request-Id": "req-first-0001", "traceparent": VALID_TP})
    second = get(app, "/healthz")
    first_record, second_record = completion_records(records)
    assert second.headers["x-request-id"] != "req-first-0001"
    assert second_record["trace_id"] != TRACE
    assert "parent_span_id" not in second_record
    assert second_record["trace_flags"] == "00"
    assert first_record["span_id"] != second_record["span_id"]

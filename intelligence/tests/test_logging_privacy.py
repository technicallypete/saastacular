"""I-LOGGING-PRIVACY: bounded JSON envelope, redaction and no content in logs."""

from __future__ import annotations

import json
import logging
import secrets
import sys
from typing import Any

import pytest
from fastapi.testclient import TestClient

from conftest import StubTransport, auth_headers, base_env, make_envelope
from main import create_app, get_gateway, get_observability
from observability import ObservabilityHook, configure_logging
from observability.export import (
    LOGGER_NAME,
    MAX_STRING_CHARS,
    REDACTED,
    JsonFormatter,
    _StdoutHandler,
)
from providers import OpenRouterGateway

ENVELOPE_KEYS = ("time", "level", "service", "event", "logger")


def canary(label: str) -> str:
    return f"CANARY-{label}-{secrets.token_hex(8)}"


def record(name: str, msg: Any, *, level: int = logging.INFO, args: Any = (), **attrs: Any):
    rec = logging.LogRecord(name, level, __file__, 1, msg, args, None)
    for key, value in attrs.items():
        setattr(rec, key, value)
    return rec


def lines(text: str) -> list[dict[str, Any]]:
    return [json.loads(line) for line in text.splitlines() if line.strip()]


# --------------------------------------------------------------- envelope


def test_i_logging_envelope_keys_and_time_format() -> None:
    out = json.loads(JsonFormatter().format(record(LOGGER_NAME, "configured")))
    assert list(out)[:5] == list(ENVELOPE_KEYS)
    assert out["service"] == "intelligence"
    assert out["event"] == "configured"
    assert out["logger"] == LOGGER_NAME
    assert out["level"] == "info"
    assert len(out["time"]) == len("2026-01-01T00:00:00.000Z") and out["time"].endswith("Z")


@pytest.mark.parametrize(
    ("level", "name"),
    [
        (logging.DEBUG, "debug"),
        (logging.INFO, "info"),
        (logging.WARNING, "warn"),
        (logging.ERROR, "error"),
        (logging.CRITICAL, "error"),
    ],
)
def test_i_logging_level_names(level: int, name: str) -> None:
    out = json.loads(JsonFormatter().format(record(LOGGER_NAME, "configured", level=level)))
    assert out["level"] == name


def test_i_logging_event_is_the_template_never_the_formatted_message() -> None:
    value = canary("arg")
    out = json.loads(JsonFormatter().format(record(LOGGER_NAME, "configured %s", args=(value,))))
    assert out["event"] == "configured %s"
    assert value not in json.dumps(out)


def test_i_logging_fields_cannot_override_the_envelope() -> None:
    fields = {"event": "forged", "service": "core", "logger": "x", "time": "t", "level": "debug"}
    out = json.loads(JsonFormatter().format(record(LOGGER_NAME, "configured", fields=fields)))
    assert (out["event"], out["service"], out["logger"]) == (
        "configured",
        "intelligence",
        LOGGER_NAME,
    )
    assert out["level"] == "info"


# -------------------------------------------------------------- redaction


def test_i_logging_secret_values_scrubbed_everywhere() -> None:
    secret = canary("secret")
    formatter = JsonFormatter([secret])
    rec = record(
        f"{LOGGER_NAME}.observability",
        f"template {secret}",
        fields={
            "plain": f"prefix {secret} suffix",
            "nested": {"deeper": [f"{secret}", {"leaf": secret}]},
            f"key-{secret}": 1,
        },
    )
    text = formatter.format(rec)
    assert secret not in text
    out = json.loads(text)
    assert out["event"] == f"template {REDACTED}"
    assert out["plain"] == f"prefix {REDACTED} suffix"
    assert out["nested"]["deeper"] == [REDACTED, {"leaf": REDACTED}]


@pytest.mark.parametrize(
    "key",
    [
        "password",
        "passwd",
        "client_secret",
        "access_token",
        "api_key",
        "apikey",
        "Authorization",
        "cookie",
        "database_dsn",
        "credential",
    ],
)
def test_i_logging_sensitive_keys_redacted_recursively(key: str) -> None:
    value = canary("value")
    fields = {key: value, "outer": {"inner": [{key: {"deep": value}}]}}
    text = JsonFormatter().format(record(LOGGER_NAME, "configured", fields=fields))
    assert value not in text
    out = json.loads(text)
    assert out[key] == REDACTED
    assert out["outer"]["inner"][0][key] == REDACTED


def test_i_logging_exception_reduced_to_type() -> None:
    value = canary("exception")
    try:
        raise ValueError(f"message {value}")
    except ValueError:
        exc_info = sys.exc_info()
    rec = logging.LogRecord(LOGGER_NAME, logging.ERROR, __file__, 1, "configured", (), exc_info)
    text = JsonFormatter().format(rec)
    assert value not in text and "Traceback" not in text
    assert json.loads(text)["error_type"] == "ValueError"


def test_i_logging_strings_capped() -> None:
    out = json.loads(
        JsonFormatter().format(record(LOGGER_NAME, "x" * 2000, fields={"long": "y" * 2000}))
    )
    assert len(out["event"]) == MAX_STRING_CHARS
    assert len(out["long"]) == MAX_STRING_CHARS


def test_i_logging_secret_scrubbed_before_cap() -> None:
    secret = "s" * 40 + canary("cap")
    text = JsonFormatter([secret]).format(
        record(LOGGER_NAME, "configured", fields={"v": "a" * 500 + secret})
    )
    assert secret[:12] not in text


def test_i_logging_secrets_are_updatable(capsys) -> None:
    first, second = canary("first"), canary("second")
    configure_logging(secrets=[first])
    configure_logging(secrets=[second])
    log = logging.getLogger(LOGGER_NAME)
    log.warning("configured", extra={"fields": {"a": first, "b": second}})
    out = capsys.readouterr().out
    assert first not in out and second not in out
    assert out.count('"event":"configured"') == 1


def test_i_logging_positive_control_unredacting_formatter_leaks() -> None:
    """The assertions above are not vacuous: a plain formatter does leak."""
    value = canary("control")
    rec = record(LOGGER_NAME, "configured %s", args=(value,), fields={"password": value})
    assert value in logging.Formatter("%(message)s").format(rec)

    class Unredacting(JsonFormatter):
        def _text(self, value: str) -> str:
            return value

        def _mapping(self, fields, depth=0):
            return dict(fields)

    assert value in Unredacting([value]).format(
        record(LOGGER_NAME, "x", fields={"password": value})
    )


# ------------------------------------------------- end-to-end through the app


def _client(env: dict[str, str], hook: ObservabilityHook | None = None, transport=None):
    app = create_app(env)
    if transport is not None:
        gateway = OpenRouterGateway(transport)
        app.dependency_overrides[get_gateway] = lambda: gateway
    if hook is not None:
        app.dependency_overrides[get_observability] = lambda: hook
    return TestClient(app)


def test_i_logging_health_request_canaries_never_logged(capsys) -> None:
    values = {name: canary(name) for name in ("auth", "cookie", "apikey", "baggage", "state")}
    values.update({name: canary(name) for name in ("x", "ua", "referer", "query", "q", "body")})
    client = _client(base_env())
    capsys.readouterr()
    response = client.request(
        "GET",
        f"/healthz?token={values['query']}&q={values['q']}",
        content=values["body"].encode(),
        headers={
            "Authorization": f"Bearer {values['auth']}",
            "Cookie": f"session={values['cookie']}",
            "X-Api-Key": values["apikey"],
            "baggage": f"k={values['baggage']}",
            "tracestate": f"v={values['state']}",
            "X-Canary": values["x"],
            "User-Agent": values["ua"],
            "Referer": f"http://example.invalid/{values['referer']}",
            "X-Request-Id": f"req-{values['x']}-with-a-space in-it",
        },
    )
    assert response.status_code == 200
    output = capsys.readouterr().out
    records = lines(output)
    completed = [r for r in records if r["event"] == "http.request.completed"]
    assert len(completed) == 1  # diagnostics are not suppressed
    assert completed[0]["route"] == "/healthz" and completed[0]["status"] == 200
    for value in values.values():
        assert value not in output
        assert value not in response.headers.get("x-request-id", "")


def test_i_logging_prompt_and_completion_never_logged(capsys) -> None:
    prompt, completion, ref = canary("prompt"), canary("completion"), canary("ref")
    authz = canary("authz")
    transport = StubTransport(
        {
            "model": "anthropic/test-model",
            "choices": [{"message": {"content": completion}}],
            "usage": {"prompt_tokens": 3, "completion_tokens": 2},
        }
    )
    client = _client(base_env(AI_MODEL_CHAT="anthropic:test-model"), transport=transport)
    envelope = make_envelope(
        authorization_ref=authz,
        context_refs=[{"kind": "document", "ref": ref}],
        input={"messages": [{"role": "user", "content": prompt}]},
    )
    response = client.post("/internal/v1/execute", json=envelope, headers=auth_headers())
    assert response.status_code == 200 and completion in response.text
    output = capsys.readouterr().out
    assert any(r["event"] == "execution" for r in lines(output))
    for value in (prompt, completion, ref, authz):
        assert value not in output


def test_i_logging_sampled_trace_and_forged_header_do_not_enable_capture(capsys) -> None:
    prompt = canary("prompt")
    exported: list[dict[str, Any]] = []
    hook = ObservabilityHook(sink=exported.append)
    client = _client(base_env(AI_MODEL_CHAT="anthropic:test-model"), hook, StubTransport())
    envelope = make_envelope(input={"messages": [{"role": "user", "content": prompt}]})
    headers = {
        **auth_headers(),
        "traceparent": "00-4bf92f3577b34da6a3ce929d0e0e4736-00f067aa0ba902b7-01",
        "X-Capture-Content": "true",
        "X-Debug": "1",
    }
    assert client.post("/internal/v1/execute", json=envelope, headers=headers).status_code == 200
    assert len(exported) == 1 and "content" not in exported[0]
    assert hook.capture_content is False
    assert prompt not in capsys.readouterr().out


def test_i_logging_uvicorn_loggers_taken_over(capsys) -> None:
    configure_logging()
    access = logging.getLogger("uvicorn.access")
    assert access.disabled and access.handlers == [] and access.propagate is False
    framework = logging.getLogger("uvicorn")
    assert framework.propagate is False
    assert [getattr(h, "_intelligence", False) for h in framework.handlers] == [True]
    value = canary("asgi")
    try:
        raise RuntimeError(value)
    except RuntimeError as exc:
        logging.getLogger("uvicorn.error").error("Exception in ASGI application\n", exc_info=exc)
    access.info('%s - "%s %s HTTP/%s" %d', "client", "GET", f"/x?{value}", "1.1", 200)
    output = capsys.readouterr().out
    assert value not in output
    [line] = lines(output)
    assert (line["event"], line["error_type"], line["level"]) == (
        "uvicorn.asgi_exception",
        "RuntimeError",
        "error",
    )


def test_i_logging_error_never_falls_back_to_text(capsys) -> None:
    class FailingFormatter(logging.Formatter):
        def format(self, record: logging.LogRecord) -> str:
            raise ValueError(canary("format"))

    handler = _StdoutHandler()
    handler.setFormatter(FailingFormatter())
    value = canary("argument")
    handler.handle(
        logging.LogRecord("intelligence", logging.ERROR, __file__, 1, "message %s", (value,), None)
    )
    output = capsys.readouterr()
    assert (output.out, output.err) == ("", "")

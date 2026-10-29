"""I-ERROR-OUTPUT: startup failures, framework, library and warning output are bounded JSON."""

from __future__ import annotations

import json
import logging
import os
import queue
import secrets
import signal
import subprocess
import sys
import threading
import time
import warnings
from pathlib import Path
from typing import Any

import pytest
import uvicorn

import main
from conftest import SERVICE_TOKEN
from observability import configure_logging
from observability.export import JsonFormatter

PROJECT = Path(__file__).resolve().parents[1]
UVICORN_SOURCE = Path(uvicorn.__file__).parent

UVICORN_EVENTS = [
    "uvicorn.process_started",
    "uvicorn.startup_waiting",
    "uvicorn.startup_complete",
    "uvicorn.listening",
    "uvicorn.shutting_down",
    "uvicorn.shutdown_waiting",
    "uvicorn.shutdown_complete",
    "uvicorn.process_finished",
    "uvicorn.app_load_failed",
    "uvicorn.invalid_request",
    "uvicorn.asgi_exception",
    "uvicorn.server_error",
    "uvicorn.other",
]
CATALOGUE = {
    "configured",
    "execution",
    "INTELLIGENCE_SERVICE_TOKEN is not set; /internal/* refuses all requests",
    "http.request.completed",
    "intelligence.startup.failed",
    "library.log",
    "python.warning",
    *UVICORN_EVENTS,
}

# Written independently of the implementation's table: uvicorn 0.54.0 template -> event.
TEMPLATES = [
    ("Started server process [%d]", "uvicorn.process_started", (7,)),
    ("Waiting for application startup.", "uvicorn.startup_waiting", ()),
    ("Application startup complete.", "uvicorn.startup_complete", ()),
    (
        "Uvicorn running on %s://%s:%d (Press CTRL+C to quit)",
        "uvicorn.listening",
        ("http", "{canary}.invalid", 8000),
    ),
    (
        "Uvicorn running on %s://[%s]:%d (Press CTRL+C to quit)",
        "uvicorn.listening",
        ("http", "{canary}", 8000),
    ),
    ("Uvicorn running on socket %s (Press CTRL+C to quit)", "uvicorn.listening", ("{canary}",)),
    (
        "Uvicorn running on unix socket %s (Press CTRL+C to quit)",
        "uvicorn.listening",
        ("/{canary}",),
    ),
    ("Shutting down", "uvicorn.shutting_down", ()),
    ("Waiting for application shutdown.", "uvicorn.shutdown_waiting", ()),
    ("Waiting for connections to close. (CTRL+C to force quit)", "uvicorn.shutdown_waiting", ()),
    (
        "Waiting for background tasks to complete. (CTRL+C to force quit)",
        "uvicorn.shutdown_waiting",
        (),
    ),
    ("Application shutdown complete.", "uvicorn.shutdown_complete", ()),
    ("Finished server process [%d]", "uvicorn.process_finished", (7,)),
    ("Invalid HTTP request received.", "uvicorn.invalid_request", ()),
    ("Unsupported upgrade request.", "uvicorn.invalid_request", ()),
    ("Exception in ASGI application\n", "uvicorn.asgi_exception", ()),
    ("Exception in 'lifespan' protocol\n", "uvicorn.asgi_exception", ()),
    ("ASGI callable returned without starting response.", "uvicorn.server_error", ()),
    ("ASGI callable returned without completing response.", "uvicorn.server_error", ()),
    ("ASGI callable should return None, but returned '%s'.", "uvicorn.server_error", ("{canary}",)),
    ("Application startup failed. Exiting.", "uvicorn.server_error", ()),
    ("Application shutdown failed. Exiting.", "uvicorn.server_error", ()),
    (
        "Cancel %s running task(s), timeout graceful shutdown exceeded",
        "uvicorn.server_error",
        (3,),
    ),
]


def canary(label: str) -> str:
    return f"CANARY{label}{secrets.token_hex(8)}"


def lines(text: str) -> list[dict[str, Any]]:
    return [json.loads(line) for line in text.splitlines() if line.strip()]


def fmt(name: str, msg: Any, args: Any = (), level: int = logging.INFO, exc_info: Any = None):
    rec = logging.LogRecord(name, level, __file__, 1, msg, args, exc_info)
    return json.loads(JsonFormatter().format(rec))


# --------------------------------------------------------- startup failure


def test_i_error_startup_failure_is_one_bounded_record(monkeypatch, capsys) -> None:
    name = "AI_MODEL_CANARY" + secrets.token_hex(8).upper()
    monkeypatch.delenv("APP_ENV", raising=False)
    monkeypatch.setenv(name, "openai:x")
    monkeypatch.setattr(main, "_app", None)
    capsys.readouterr()
    with pytest.raises(SystemExit) as raised:
        _ = main.app
    assert raised.value.code == 3
    assert raised.value.__suppress_context__ and raised.value.__cause__ is None
    out = capsys.readouterr()
    assert name not in out.out + out.err
    assert "Traceback" not in out.out + out.err
    [record] = lines(out.out)
    assert record["event"] == "intelligence.startup.failed"
    assert (record["level"], record["service"], record["error_type"]) == (
        "error",
        "intelligence",
        "ConfigError",
    )
    assert set(record) == {"time", "level", "service", "event", "logger", "error_type"}
    assert main._app is None


# ------------------------------------------------------- uvicorn templates


@pytest.mark.parametrize(
    ("template", "event", "args"), TEMPLATES, ids=[t[1] + ":" + t[0][:24] for t in TEMPLATES]
)
def test_i_error_uvicorn_template_mapping(template: str, event: str, args: tuple) -> None:
    value = canary("arg")
    filled = tuple(a.format(canary=value) if isinstance(a, str) else a for a in args)
    out = fmt("uvicorn.error", template, filled)
    assert out["event"] == event
    assert value not in json.dumps(out)
    assert set(out) <= {"time", "level", "service", "event", "logger", "error_type"}


def test_i_error_uvicorn_templates_exist_in_installed_package() -> None:
    source = "\n".join(p.read_text(encoding="utf-8") for p in UVICORN_SOURCE.rglob("*.py"))
    assert uvicorn.__version__ == "0.54.0"
    for template, _, _ in TEMPLATES:
        if template.startswith("Uvicorn running on %s:"):
            continue  # built from addr_format, checked below
        assert template.replace("\n", "\\n") in source, template
    assert 'addr_format = "%s://%s:%d"' in source
    assert 'addr_format = "%s://[%s]:%d"' in source
    assert 'f"Uvicorn running on {addr_format} (Press CTRL+C to quit)"' in source
    assert 'logger.error("Error loading ASGI app. %s" % exc)' in source


def test_i_error_uvicorn_unknown_template_is_other() -> None:
    value = canary("other")
    out = fmt("uvicorn.error", f"some future message {value}", level=logging.WARNING)
    assert (out["event"], out["level"]) == ("uvicorn.other", "warn")
    assert value not in json.dumps(out)


def test_i_error_uvicorn_app_load_failure_drops_formatted_text() -> None:
    value = canary("module")
    out = fmt("uvicorn.error", f'Error loading ASGI app. Could not import module "{value}".')
    assert out["event"] == "uvicorn.app_load_failed"
    assert value not in json.dumps(out)


def test_i_error_uvicorn_bind_error_object_keeps_type_only() -> None:
    value = canary("host")
    exc = OSError(99, f"error while attempting to bind on address ('{value}', 8000)")
    out = fmt("uvicorn.error", exc, level=logging.ERROR)
    assert (out["event"], out["error_type"], out["level"]) == (
        "uvicorn.server_error",
        "OSError",
        "error",
    )
    assert value not in json.dumps(out)


def test_i_error_uvicorn_exception_info_keeps_type_only() -> None:
    value = canary("exc")
    try:
        raise KeyError(value)
    except KeyError:
        exc_info = sys.exc_info()
    out = fmt("uvicorn.error", "Exception in ASGI application\n", level=40, exc_info=exc_info)
    assert (out["event"], out["error_type"]) == ("uvicorn.asgi_exception", "KeyError")
    assert value not in json.dumps(out)


# ------------------------------------------------ libraries and warnings


def test_i_error_third_party_warning_becomes_library_log(capsys) -> None:
    configure_logging()
    value = canary("lib")
    capsys.readouterr()
    logging.getLogger("thirdparty.client").warning("connection to %s failed", value)
    logging.getLogger("thirdparty.client").info("ignored below WARNING %s", value)
    out = capsys.readouterr().out
    assert value not in out
    [record] = lines(out)
    assert (record["event"], record["logger"], record["level"]) == (
        "library.log",
        "thirdparty.client",
        "warn",
    )
    assert set(record) == {"time", "level", "service", "event", "logger"}


def test_i_error_library_logger_name_is_validated() -> None:
    value = canary("name")
    out = fmt(f"bad name {value}", "message", level=logging.WARNING)
    assert out["event"] == "library.log" and out["logger"] == "unknown"
    assert value not in json.dumps(out)


def test_i_error_warning_becomes_python_warning(capsys) -> None:
    value = canary("warn")
    with warnings.catch_warnings():
        warnings.simplefilter("always")
        configure_logging()
        capsys.readouterr()
        warnings.warn(value, UserWarning, stacklevel=1)
        warnings.warn(value, DeprecationWarning, stacklevel=1)
    captured = capsys.readouterr()
    assert value not in captured.out + captured.err
    records = lines(captured.out)
    assert [(r["event"], r["category"], r["level"]) for r in records] == [
        ("python.warning", "UserWarning", "warn"),
        ("python.warning", "DeprecationWarning", "warn"),
    ]


# ------------------------------------------------------- handler takeover


def test_i_error_exactly_one_handler_and_no_duplicate_lines(capsys) -> None:
    for _ in range(3):
        configure_logging(secrets=(canary("dup"),))
    root = logging.getLogger()
    ours = [h for h in root.handlers if getattr(h, "_intelligence", False)]
    assert len(ours) == 1
    assert not any(
        getattr(h, "_intelligence", False) for h in logging.getLogger("intelligence").handlers
    )
    assert logging.getLogger("uvicorn").handlers == ours
    assert logging.getLogger("uvicorn.error").handlers == []
    assert root.level == logging.WARNING
    assert logging.getLogger("intelligence").level == logging.INFO
    capsys.readouterr()
    logging.getLogger("intelligence.http").info("http.request.completed")
    logging.getLogger("uvicorn.error").info("Application startup complete.")
    out = capsys.readouterr().out
    assert [r["event"] for r in lines(out)] == [
        "http.request.completed",
        "uvicorn.startup_complete",
    ]


def test_i_error_uvicorn_access_logger_disabled(capsys) -> None:
    configure_logging()
    access = logging.getLogger("uvicorn.access")
    assert access.disabled is True
    assert access.handlers == []
    assert access.propagate is False
    capsys.readouterr()
    access.info('%s - "%s %s HTTP/%s" %d', "127.0.0.1:1", "GET", "/healthz?q=1", "1.1", 200)
    assert capsys.readouterr().out == ""


# --------------------------------------------- the real uvicorn process


def _uvicorn_env(**extra: str) -> dict[str, str]:
    env = {
        "PATH": os.environ.get("PATH", "/usr/bin:/bin"),
        "PYTHONPATH": str(PROJECT / "src"),
        "PYTHONUNBUFFERED": "1",
        "PYTHONDONTWRITEBYTECODE": "1",
    }
    env.update(extra)
    return env


UVICORN_CMD = [sys.executable, "-m", "uvicorn", "main:app", "--host", "127.0.0.1", "--port", "0"]


def _assert_catalogued(output: str) -> list[dict[str, Any]]:
    records = []
    for raw in output.splitlines():
        record = json.loads(raw)  # every line is JSON; a raw text line fails here
        assert record["service"] == "intelligence"
        assert record["event"] in CATALOGUE, record["event"]
        records.append(record)
    return records


def test_i_error_real_uvicorn_startup_failure_exits_3_with_json_only() -> None:
    name = "AI_MODEL_CANARY" + secrets.token_hex(8).upper()
    result = subprocess.run(
        UVICORN_CMD,
        cwd=PROJECT,
        env=_uvicorn_env(**{name: "openai:x"}),
        capture_output=True,
        text=True,
        timeout=60,
        check=False,
    )
    output = result.stdout + result.stderr
    assert result.returncode == 3, output
    assert name not in output and "Traceback" not in output
    records = _assert_catalogued(output)
    assert [(r["event"], r.get("error_type")) for r in records] == [
        ("intelligence.startup.failed", "ConfigError")
    ]


def test_i_error_real_uvicorn_lifecycle_is_catalogued_json() -> None:
    process = subprocess.Popen(
        UVICORN_CMD,
        cwd=PROJECT,
        env=_uvicorn_env(APP_ENV="test", INTELLIGENCE_SERVICE_TOKEN=SERVICE_TOKEN),
        stdout=subprocess.PIPE,
        stderr=subprocess.STDOUT,
        text=True,
    )
    collected: queue.Queue[str] = queue.Queue()

    def pump() -> None:
        assert process.stdout is not None
        for raw in process.stdout:
            collected.put(raw)

    reader = threading.Thread(target=pump, daemon=True)
    reader.start()
    output: list[str] = []
    try:
        deadline = time.monotonic() + 60
        while not any("uvicorn.startup_complete" in raw for raw in output):
            assert time.monotonic() < deadline, "".join(output)
            try:
                output.append(collected.get(timeout=1))
            except queue.Empty:
                assert process.poll() is None, "".join(output)
        process.send_signal(signal.SIGTERM)
        # After a graceful shutdown uvicorn re-raises the captured signal, so the
        # process ends by SIGTERM; the shutdown records below prove it was graceful.
        assert process.wait(timeout=30) in (0, -signal.SIGTERM)
    finally:
        if process.poll() is None:
            process.kill()
            process.wait(timeout=10)
    reader.join(timeout=10)
    while not collected.empty():
        output.append(collected.get())
    text = "".join(output)
    assert "Uvicorn running" not in text and "127.0.0.1" not in text
    events = [r["event"] for r in _assert_catalogued(text)]
    assert events[:2] == ["configured", "uvicorn.process_started"]
    for expected in (
        "uvicorn.startup_waiting",
        "uvicorn.startup_complete",
        "uvicorn.listening",
        "uvicorn.shutting_down",
        "uvicorn.shutdown_waiting",
        "uvicorn.shutdown_complete",
        "uvicorn.process_finished",
    ):
        assert events.count(expected) == 1, events
    listening = next(r for r in lines(text) if r["event"] == "uvicorn.listening")
    assert set(listening) == {"time", "level", "service", "event", "logger"}

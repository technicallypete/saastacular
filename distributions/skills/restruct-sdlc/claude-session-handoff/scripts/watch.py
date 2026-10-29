#!/usr/bin/env python3
"""Coordinator -> supervisor milestone observer for one repository; standard library only.

Polls the verified, pinned Claude coordinator's own session transcript and queues concise milestone notices to one
explicitly bound, existing Codex supervisor thread with `codex queue`. No model inference, approvals, session
creation, settings changes, raw transcript forwarding or implicit retargeting. Notices carry no authority.

  watch.py --repo PATH bind (--supervisor UUID | --supervisor-file FILE) [--replace]
                             [--size-threshold BYTES | --no-size-advisory]
  watch.py --repo PATH start [--resume] [--interval SECONDS] | stop | status | unbind
  watch.py --repo PATH smoke --label TEXT      # one labelled benign delivery to the bound thread
  watch.py --repo PATH poll-once               # one foreground poll (diagnostics)

State is private (0700/0600) under the handoff helper's repository state directory, in `observer/`, outside Git.
Uncertain identity, incompatible capabilities, state or queue outcome suspends delivery until an explicit `bind --replace`.
Each poll also reads the coordinator transcript's size from file metadata; at or above the bound threshold (default
100 MiB) one advisory is queued per coordinator UUID and threshold. Nothing is pruned and no session is changed.
"""
import argparse
import contextlib
import fcntl
import hashlib
import json
import os
from pathlib import Path
import re
import signal
import subprocess
import sys
import time
import uuid

import handoff
from handoff import Blocked, require

SCHEMA = "claude-session-handoff/observer/v1"
BASELINE_CODEX_CLI = {"0.160.0"}
BASELINE_CODEX_DAEMON = {"0.162.1", "0.163.0"}
DEFAULT_INTERVAL = 30
DEFAULT_SIZE_THRESHOLD = 104857600  # 100 MiB
MIN_SIZE_THRESHOLD = 1024
MIN_INTERVAL, MAX_INTERVAL = 5, 300
QUEUE_TIMEOUT = 30
MAX_READ = 1048576           # one read; also the per-row parsing limit
MAX_SCAN = 16 * MAX_READ     # bytes examined per poll (bounded progress through oversized rows)
PREFIX_BYTES = 4096          # kept from an oversized row to tell whether it was coordinator output
ASSISTANT_PREFIX = re.compile(r'"type"\s*:\s*"assistant"')
SIDECHAIN_PREFIX = re.compile(r'"isSidechain"\s*:\s*true')
MAX_KEYS = 2000
NOTICE_MARK = "[observer-notice:"
ECHO_MARKS = (NOTICE_MARK, "[handoff-id:")
STATUS_MARKER = re.compile(
    r"COORDINATOR_STATUS: (working|decision|review-ready|blocked|question|failed|completed)"
    r" \| ([A-Za-z0-9][A-Za-z0-9_.-]{0,63}) \| ([^\r\n|]+)")
# A mixed-task working summary can itself report an actionable handoff. Inspect only that bounded summary,
# not historical body prose; keep the existing source/echo and native-question checks ahead of this fallback.
WORKING_HANDOFFS = (
    ("question", re.compile(r"\b(?:(?:decision|approval)\s+(?:is\s+)?(?:pending|needed|required)|"
                            r"(?:awaiting|waiting (?:on|for)|needs?|requires?)\s+"
                            r"(?:(?:the )?(?:owner|codex|supervisor)(?:['’]s)?\s+)?"
                            r"(?:decision|approval))\b", re.I)),
    ("review-ready", re.compile(r"\b(?:review[- ]ready|(?:ready|submitted) for\s+"
                                r"(?:(?:codex|supervisor)(?:['’]s)?\s+)?(?:re[- ]?)?review|"
                                r"(?:awaiting|waiting (?:on|for))\s+"
                                r"(?:(?:codex|supervisor)(?:['’]s)?\s+)?(?:re[- ]?)?review)\b", re.I)),
)
CATEGORIES = (
    ("failed", re.compile(r"\b(?:failed|failing|failures?)\b", re.I)),
    ("question", re.compile(r"\b(?:question for (?:the )?(?:owner|codex|supervisor)|"
                            r"(?:decision|approval)\s+(?:is\s+)?(?:needed|required))\b", re.I)),
    ("blocked", re.compile(r"\b(?:blocked|blocker|escalat(?:e|es|ed|ing|ion))\b", re.I)),
    ("review-ready", re.compile(r"\b(?:review[- ]ready|ready for (?:(?:codex|supervisor)(?:['’]s)? )?"
                                r"(?:review|mechanical (?:verification|confirmation))|submitted for "
                                r"(?:codex |supervisor )?(?:implementation |plan )?review)\b", re.I)),
    ("completed", re.compile(r"\b(?:handback|handed back|work (?:is )?complete|completed (?:the )?"
                             r"(?:task|goal|attempt))\b", re.I)),
)
STOP = False


def run_codex(arguments, timeout):
    """The only Codex CLI boundary (tests replace it)."""
    return subprocess.run(["codex", *arguments], capture_output=True, text=True, timeout=timeout)


def coordinator_status(root):
    """Reuse the handoff helper's pin/discovery verification (tests replace it)."""
    registry = Path.home() / ".claude" / "sessions"
    projects = Path.home() / ".claude" / "projects"
    report = handoff.status_report(root, registry, projects)
    if report.get("status") == "pinned-live-verified":
        row = handoff.live_registration(registry, report["pin"]["session"], root)
        report["registration_status"] = row.get("status")
    return report


def projects_dir():
    return Path.home() / ".claude" / "projects"


def fingerprint(value):
    return hashlib.sha256(value.encode()).hexdigest()[:12]


def observer_dir(root):
    return handoff.state_subdir(root, "observer")


def load(directory, name, default=None):
    data = handoff.read_private(directory / name)
    if data is None:
        return default
    value = json.loads(data)
    require(isinstance(value, dict) and value.get("schema") == SCHEMA, "state-unsafe")
    return value


def save(directory, name, value):
    handoff.write_private_atomic(directory, name, json.dumps({**value, "schema": SCHEMA}, sort_keys=True).encode())


def empty_state():
    return {"offsets": {}, "keys": [], "deliveries": [], "suspended": None, "waiting": False, "smoke": None,
            "size_advised": [], "transcript": None}


def size_check(state, binding, path):
    """Metadata-only size check; (key, size, threshold) once per coordinator UUID and threshold, else None."""
    threshold = binding.get("size_threshold", DEFAULT_SIZE_THRESHOLD)
    size = path.stat().st_size
    key = f"size:{binding['coordinator']}:{threshold}"
    advised = state.setdefault("size_advised", [])
    state["transcript"] = {"bytes": size, "threshold": threshold, "coordinator": binding["coordinator"],
                           "advised": key in advised, "time": handoff.utc_now().isoformat()}
    if threshold is None or size < threshold or key in advised:
        return None
    return key, size, threshold


def suspend(directory, state, reason):
    if not state.get("suspended"):
        state["suspended"] = {"reason": reason, "time": handoff.utc_now().isoformat()}
    save(directory, "state.json", state)
    return {"status": "suspended", "reason": state["suspended"]["reason"], "queue_calls": 0}


def codex_versions():
    cli = run_codex(["--version"], 15)
    match = re.search(r"(\d+\.\d+\.\d+)", cli.stdout or "")
    daemon = run_codex(["app-server", "daemon", "version"], 15)
    try:
        server = json.loads(daemon.stdout).get("appServerVersion")
        if not isinstance(server, str) or not re.fullmatch(r"\d+\.\d+\.\d+", server):
            server = None
    except (ValueError, AttributeError):
        server = None
    return (match.group(1) if cli.returncode == 0 and match else None), (server if daemon.returncode == 0 else None)


def queue_compatibility():
    """Read-only surface checks, not a guarantee about future implementations. No test message is sent."""
    cli, daemon = codex_versions()
    help_result = run_codex(["queue", "--help"], 15)
    help_text = help_result.stdout or ""
    supported = (cli is not None and daemon is not None and help_result.returncode == 0
                 and "Usage: codex queue" in help_text
                 and all(re.search(r"--" + flag + r"[ =]+<[^>\n]+>", help_text)
                         for flag in ("thread", "message")))
    warnings = []
    if cli not in BASELINE_CODEX_CLI:
        warnings.append("unbaselined-codex-cli")
    if daemon not in BASELINE_CODEX_DAEMON:
        warnings.append("unbaselined-codex-daemon")
    return supported, cli, daemon, warnings


def verified_identity(root, binding):
    """(ok, reason, report): the pin must be live-verified and equal the bound coordinator UUID."""
    report = coordinator_status(root)
    status = report.get("status")
    if status != "pinned-live-verified":
        return False, f"coordinator-{status or 'unknown'}", report
    if report.get("pin", {}).get("session") != binding["coordinator"]:
        return False, "coordinator-changed-rebind-required", report
    return True, None, report


def transcript_path(session):
    paths = handoff.transcripts(projects_dir(), session)
    require(len(paths) == 1, "coordinator-transcript-ambiguous" if paths else "coordinator-transcript-missing")
    return paths[0]


def new_rows(path, offsets, limits):
    """Complete JSONL rows after the stored offset, with bounded memory and bounded work per poll.

    A truncated or replaced file restarts at 0; a short partial tail waits. A row longer than MAX_READ is scanned
    forward in MAX_READ chunks (resuming across polls while it grows) until its newline; it is never parsed, its
    fragments are never parsed as records, and it is appended to `limits` (start, bytes, prefix) for reporting.
    """
    info = path.stat()
    key = str(path)
    stored = offsets.get(key)
    if stored is None:
        offsets[key] = {"inode": info.st_ino, "offset": info.st_size}
        return []
    if stored["inode"] != info.st_ino or info.st_size < stored["offset"]:
        stored = {"inode": info.st_ino, "offset": 0}  # replaced, or truncated below the committed offset
    elif info.st_size < (stored.get("oversized") or {}).get("scanned", 0):
        # Truncated inside a pending oversized row: the bytes before its committed start are intact, so drop the
        # stale scan state and re-read from that start (no history replay, no stall at the old scan position).
        stored = {"inode": info.st_ino, "offset": stored["offset"]}
    pos, over, examined, rows = stored["offset"], stored.get("oversized"), 0, []
    with path.open("rb") as stream:
        while examined < MAX_SCAN:
            if over:
                stream.seek(over["scanned"])
                data = stream.read(MAX_READ)
                examined += len(data)
                newline = data.find(b"\n")
                if newline < 0:
                    over["scanned"] += len(data)
                    if len(data) < MAX_READ:
                        break  # still growing: resume from `scanned` next poll
                    continue
                end = over["scanned"] + newline + 1
                if not over.get("discard_partial"):
                    limits.append({"inode": info.st_ino, "start": over["start"], "bytes": end - over["start"],
                                   "prefix": over["prefix"]})
                pos, over = end, None
                continue
            stream.seek(pos)
            data = stream.read(MAX_READ)
            examined += len(data)
            last = data.rfind(b"\n")
            if last < 0:
                if len(data) < MAX_READ:
                    break  # short partial tail: wait for its newline
                over = {"start": pos, "scanned": pos + len(data),
                        "prefix": data[:PREFIX_BYTES].decode("utf-8", errors="replace")}
                continue
            for raw in data[:last + 1].split(b"\n")[:-1]:  # JSONL rows end at LF only; U+2028/U+2029/U+0085
                line = raw.decode("utf-8", errors="replace")   # inside JSON strings are content, not delimiters
                with contextlib.suppress(ValueError):
                    row = json.loads(line)
                    if isinstance(row, dict):
                        rows.append(row)
            pos += last + 1
            if len(data) < MAX_READ:
                break
    offsets[key] = {"inode": info.st_ino, "offset": pos, **({"oversized": over} if over else {})}
    return rows


def active_alert(line, pattern):
    """Ignore negated/resolved alerts and conditional failure rules; retain separate active issues."""
    line = re.sub(r"[*`_]", "", line)
    for match in pattern.finditer(line):
        before, after = line[:match.start()], line[match.end():]
        negated = re.search(r"\b(?:no|zero|0|without|not|never|no longer|expected)\s+"
                            r"(?:(?:known|remaining|current|new|active|currently|still|unexpected)\s+)*$", before, re.I)
        resolved = re.match(r"(?:\s+(?:gate|pipeline|job|test|check|build|validation))?"
                            r"(?:\s+(?:is|was|now|already|has been|had been))*\s+"
                            r"(?:resolved|cleared|fixed|removed|gone|passed|passing|recovered)\b", after, re.I)
        conditional = (match.group().lower() in ("failure", "failures")
                       and re.search(r"\b(?:any|on(?: any)?)\s+$", before, re.I)
                       and re.match(r"\s+(?:we\s+)?(?:(?:will|would)\s+)?(?:stops?|rolls? back|returns?|triggers?)\b",
                                    after, re.I))
        if not negated and not resolved and not conditional:
            return True
    return False


def explicit_status(text):
    """Parse the full final status within the row limit; truncate only after classification."""
    lines = text.rstrip().split("\n")
    marker = STATUS_MARKER.fullmatch(lines[-1])
    if marker is None or not marker[3].strip():
        return None
    fence = None
    for line in lines[:-1]:
        match = re.match(r"^\s*(`{3,}|~{3,})", line)
        if match:
            delimiter = match[1]
            if fence is None:
                fence = delimiter
            elif delimiter[0] == fence[0] and len(delimiter) >= len(fence):
                fence = None
    if fence is not None:
        return None
    return marker[1], f"{marker[2]}: {marker[3].strip()}"


def milestone(row):
    """A new coordinator-authored milestone, or None. Incoming handoffs, echoes and receipts never qualify."""
    if row.get("type") != "assistant" or row.get("isSidechain"):
        return None
    content = row.get("message", {}).get("content", [])
    if isinstance(content, str):
        content = [{"type": "text", "text": content}]
    texts = [c.get("text", "") for c in content if isinstance(c, dict) and c.get("type") == "text"]
    joined = "\n".join(texts)
    if any(mark in joined for mark in ECHO_MARKS):
        return None
    if any(isinstance(c, dict) and c.get("type") == "tool_use" and c.get("name") == "AskUserQuestion"
           for c in content):
        return "question", "the coordinator asked a native question (inspect it in the coordinator session)"
    status = explicit_status(joined)
    if status is not None:
        if status[0] != "working":
            return status[0], handoff.safe_excerpt(status[1], 240)
        for category, pattern in WORKING_HANDOFFS:
            if active_alert(status[1], pattern):
                return category, handoff.safe_excerpt(status[1], 240)
        return None
    for category, pattern in CATEGORIES:
        for line in joined.split("\n"):
            if re.match(r"^\s*(?:>\s*)?COORDINATOR_STATUS:", line):
                continue  # malformed/nonterminal/quoted markers are not prose alerts
            if pattern.search(line) and (category not in ("failed", "blocked", "question")
                                         or active_alert(line, pattern)):
                return category, handoff.safe_excerpt(line.strip(), 240)
    return None


def row_key(row):
    if isinstance(row.get("uuid"), str):
        return "row:" + row["uuid"]
    return "row:" + hashlib.sha256(json.dumps(row, sort_keys=True).encode()).hexdigest()


def remember(state, key):
    state["keys"] = (state["keys"] + [key])[-MAX_KEYS:]


def recover_pending(state):
    """An intent persisted before `codex queue` but never finalized is uncertain: never retried, never acknowledged."""
    found = False
    for record in state.get("deliveries", []):
        if record.get("outcome") == "pending-intent":
            record["outcome"], found = "uncertain-interrupted", True
    if (state.get("smoke") or {}).get("outcome") == "pending-intent":
        state["smoke"]["outcome"], found = "uncertain-interrupted", True
    return found


def recover_known_unsent(state):
    """Explicit rebind may release legacy version-blocked markers, never a key with an attempted/uncertain send."""
    records = state.get("deliveries", [])
    unsent = {"not-sent-unverified-codex-version", "not-sent-incompatible-codex-capabilities"}
    attempted = {r.get("key") for r in records if r.get("outcome") not in unsent or r.get("notice_id")}
    release = {r.get("key") for r in records
               if r.get("outcome") in unsent and isinstance(r.get("key"), str)} - attempted
    state["keys"] = [key for key in state["keys"] if key not in release]
    state["size_advised"] = [key for key in state.get("size_advised", []) if key not in release]
    transcript = state.get("transcript")
    if transcript and f"size:{transcript.get('coordinator')}:{transcript.get('threshold')}" in release:
        transcript["advised"] = False
    waiting = [r for r in records if r.get("kind") == "waiting"]
    if waiting and waiting[-1].get("key") in release:
        state["waiting"] = False
    smoke = state.get("smoke")
    if smoke and f"smoke:{smoke.get('label')}" in release:
        state["smoke"] = None


def deliver(directory, state, binding, text, key, kind):
    """Queue one notice. The intent and its dedup key are persisted first (no queue if that fails); uncertainty is
    recorded and suspends delivery; stored intents and attempted sends are never retried."""
    try:
        supported, cli, daemon, warnings = queue_compatibility()
    except (OSError, subprocess.SubprocessError):
        supported, cli, daemon, warnings = False, None, None, []
    if not supported:
        state["deliveries"].append({"key": key, "kind": kind, "outcome": "not-sent-incompatible-codex-capabilities",
                                    "cli": cli, "daemon": daemon})
        return suspend(directory, state, "incompatible-codex-capabilities")
    notice_id = str(uuid.uuid4())
    message = f"{NOTICE_MARK} {notice_id}] {text}"
    remember(state, key)
    # Consume markers only after capability verification, in the same durable write as the queue intent.
    if kind == "size-advisory":
        state["size_advised"] = (state.get("size_advised", []) + [key])[-200:]
        state["transcript"]["advised"] = True
    elif kind == "waiting":
        state["waiting"] = True
    elif kind == "smoke":
        state["smoke"] = {"label": key.removeprefix("smoke:"), "outcome": "pending-intent",
                          "time": handoff.utc_now().isoformat()}
    record = {"key": key, "kind": kind, "notice_id": notice_id, "time": handoff.utc_now().isoformat(),
              "message_sha256": hashlib.sha256(message.encode()).hexdigest(), "cli": cli, "daemon": daemon,
              "compatibility_warnings": warnings,
              "outcome": "pending-intent"}
    state["deliveries"] = (state["deliveries"] + [record])[-200:]
    save(directory, "state.json", state)  # durable intent before the external queue; raises -> nothing queued
    try:
        result = run_codex(["queue", "--thread", binding["supervisor"], "--message", message], QUEUE_TIMEOUT)
        outcome = "queued-not-acknowledged" if result.returncode == 0 else f"queue-uncertain-exit-{result.returncode}"
    except (OSError, subprocess.SubprocessError):
        outcome = "queue-uncertain-error"
    record["outcome"] = outcome
    if outcome != "queued-not-acknowledged":
        return suspend(directory, state, "queue-uncertain")
    save(directory, "state.json", state)
    return {"status": "queued-not-acknowledged", "notice_id": notice_id, "queue_calls": 1,
            "compatibility_warnings": warnings}


def poll_once(root):
    directory = observer_dir(root)
    binding = load(directory, "binding.json")
    if binding is None:
        return {"status": "unbound", "queue_calls": 0}
    state = load(directory, "state.json", empty_state())
    if recover_pending(state):
        return suspend(directory, state, "interrupted-delivery-uncertain")
    if state.get("suspended"):
        return {"status": "suspended", "reason": state["suspended"]["reason"], "queue_calls": 0}
    ok, reason, report = verified_identity(root, binding)
    if not ok:
        return suspend(directory, state, reason)
    offsets, limits = json.loads(json.dumps(state["offsets"])), []  # committed only after every delivery
    try:
        path = transcript_path(binding["coordinator"])
        rows = new_rows(path, offsets, limits)
        advisory = size_check(state, binding, path)
    except Blocked as error:
        return suspend(directory, state, str(error))
    events = []
    for limit in limits:
        assistant = bool(ASSISTANT_PREFIX.search(limit["prefix"])) and not SIDECHAIN_PREFIX.search(limit["prefix"])
        state["parse_limits"] = (state.get("parse_limits", []) + [
            {"start": limit["start"], "bytes": limit["bytes"], "coordinator_output": assistant,
             "time": handoff.utc_now().isoformat()}])[-50:]
        key = f"oversized:{limit['inode']}:{limit['start']}"
        if assistant and key not in state["keys"]:
            events.append((key, "parse-limit", f"a coordinator output row of {limit['bytes']} bytes exceeded the "
                           f"{MAX_READ}-byte parsing limit and could not be assessed for milestones; inspect the "
                           "coordinator session", ""))
    for row in rows:
        found = milestone(row)
        key = row_key(row)
        if found and key not in state["keys"]:
            events.append((key, *found, row.get("timestamp", "")))
    waiting = report.get("registration_status") == "waiting"
    if waiting and not state.get("waiting"):
        events.append((f"waiting:{handoff.utc_now().isoformat()}", "waiting",
                       "the coordinator reports waiting; the native prompt was not inspected (inspect with claude "
                       "attach before answering)", ""))
    if not waiting:
        state["waiting"] = False
    calls = 0
    if advisory:
        key, size, threshold = advisory
        text = (f"Claude coordinator transcript size advisory: {size} bytes ({size / 1048576:.1f} MiB) is at or above "
                f"the {threshold}-byte threshold. Recommendation: assess a safe session handoff or archive at a safe "
                "point under the recovery rules. Nothing was pruned, no session was changed, and this notice carries "
                "no approval or task authority.")
        result = deliver(directory, state, binding, text, key, "size-advisory")
        calls += result.get("queue_calls", 0)
        if result["status"] == "suspended":
            return {**result, "queue_calls": calls}
    for key, category, excerpt, stamp in events:
        text = (f"Claude coordinator milestone ({category}{', ' + stamp if stamp else ''}): {excerpt}. This notice "
                "carries no approval or task authority; verify the repository records.")
        result = deliver(directory, state, binding, text, key, category)
        calls += result.get("queue_calls", 0)
        if result["status"] == "suspended":
            return {**result, "queue_calls": calls}
    state["offsets"] = offsets
    save(directory, "state.json", state)
    return {"status": "polled", "events": len(events), "queue_calls": calls}


# ---------------------------------------------------------------- lifecycle

def child_argv(root, interval):
    return [sys.executable, "-B", str(Path(__file__).resolve()), "--repo", str(root), "run", "--interval",
            str(interval)]


def cmdline(pid):
    try:
        return Path(f"/proc/{pid}/cmdline").read_bytes().split(b"\0")[:-1]
    except OSError:
        return None


def verified_process(record):
    """The recorded PID still runs the recorded command with the recorded start time (no PID-reuse kill)."""
    if not record:
        return False
    pid = record.get("pid")
    if not isinstance(pid, int) or pid <= 1:
        return False
    try:
        start = handoff.process_start(pid)
    except (OSError, ValueError, IndexError):
        return False
    argv = cmdline(pid)
    return start == record.get("start") and argv == [a.encode() for a in record.get("argv", [])]


def cmd_start(root, interval, resume=False):
    require(MIN_INTERVAL <= interval <= MAX_INTERVAL, "interval-out-of-range")
    directory = observer_dir(root)
    binding = load(directory, "binding.json")
    require(binding is not None, "unbound-bind-first")
    record = load(directory, "process.json")
    if verified_process(record):
        return {"status": "already-running", "pid": record["pid"]}
    ok, reason, _ = verified_identity(root, binding)
    if not ok:
        return {"status": "inactive", "reason": reason, "next": "verify the session pair and rebind explicitly"}
    if not resume:
        state = load(directory, "state.json", empty_state())
        if recover_pending(state):
            return suspend(directory, state, "interrupted-delivery-uncertain")
        require(not state.get("suspended"), "suspended-rebind-first")
        path = transcript_path(binding["coordinator"])
        with path.open("rb") as stream:
            info = os.fstat(stream.fileno())
            partial = False
            if info.st_size:
                stream.seek(info.st_size - 1)
                partial = stream.read(1) != b"\n"
        cursor = {"inode": info.st_ino, "offset": info.st_size}
        if partial:
            cursor["oversized"] = {"start": info.st_size, "scanned": info.st_size,
                                   "prefix": "", "discard_partial": True}
        state["offsets"][str(path)] = cursor
        state["waiting"] = False
        save(directory, "state.json", state)
    argv = child_argv(root, interval)
    log = os.open(directory / "observer.log", os.O_WRONLY | os.O_CREAT | os.O_APPEND | os.O_NOFOLLOW, 0o600)
    try:
        child = subprocess.Popen(argv, stdin=subprocess.DEVNULL, stdout=log, stderr=log, start_new_session=True,
                                 close_fds=True)
    finally:
        os.close(log)
    time.sleep(0.3)
    if child.poll() is not None:
        return {"status": "blocked", "reason": f"observer-exited-{child.returncode}"}
    save(directory, "process.json", {"pid": child.pid, "start": handoff.process_start(child.pid), "argv": argv,
                                     "interval": interval, "started": handoff.utc_now().isoformat()})
    return {"status": "started", "pid": child.pid, "interval": interval}


def cmd_stop(root):
    directory = observer_dir(root)
    record = load(directory, "process.json")
    if not record:
        return {"status": "not-running"}
    if not verified_process(record):
        return {"status": "stale-record-not-signalled", "pid": record.get("pid"),
                "next": "the recorded process is gone or its identity changed; no signal was sent"}
    os.kill(record["pid"], signal.SIGTERM)
    deadline = time.monotonic() + 10
    while time.monotonic() < deadline and verified_process(record):
        time.sleep(0.1)
    if verified_process(record):
        return {"status": "blocked", "reason": "observer-did-not-stop", "pid": record["pid"]}
    os.unlink(directory / "process.json")
    return {"status": "stopped", "pid": record["pid"]}


def cmd_status(root):
    directory = observer_dir(root)
    binding = load(directory, "binding.json")
    state = load(directory, "state.json", empty_state())
    record = load(directory, "process.json")
    running = verified_process(record)
    return {"status": "running" if running else ("stale-record" if record else "stopped"),
            "pid": record.get("pid") if running else None,
            "bound": binding is not None,
            "supervisor_fingerprint": fingerprint(binding["supervisor"]) if binding else None,
            "coordinator": binding.get("coordinator") if binding else None,
            "suspended": state.get("suspended"), "deliveries": len(state.get("deliveries", [])),
            "last_delivery": (state.get("deliveries") or [None])[-1], "dedup_keys": len(state.get("keys", [])),
            "size_threshold": binding.get("size_threshold", DEFAULT_SIZE_THRESHOLD) if binding else None,
            "transcript": state.get("transcript"), "size_advised": len(state.get("size_advised", [])),
            "parse_limit": {"row_bytes": MAX_READ, "oversized_rows": len(state.get("parse_limits", [])),
                            "last": (state.get("parse_limits") or [None])[-1]},
            "pending_intents": sum(1 for d in state.get("deliveries", []) if d.get("outcome") == "pending-intent")}


def read_supervisor(args):
    if args.supervisor:
        return args.supervisor
    raw = handoff.read_private(Path(args.supervisor_file))
    require(raw is not None, "supervisor-file-missing")
    text = raw.decode().strip()
    with contextlib.suppress(ValueError):
        value = json.loads(text)
        if isinstance(value, dict):
            text = value.get("supervisor_thread", "")
    return text


def cmd_bind(root, args):
    supervisor = read_supervisor(args)
    require(handoff.is_canonical_uuid(supervisor), "invalid-supervisor-uuid")
    directory = observer_dir(root)
    current = load(directory, "binding.json")
    require(current is None or args.replace, "already-bound-use-replace")
    report = coordinator_status(root)
    require(report.get("status") == "pinned-live-verified", "coordinator-not-pinned-live-verified")
    if args.no_size_advisory:
        threshold = None
    elif args.size_threshold is None:  # omitted: default; an explicit value (including 0) is validated as given
        threshold = DEFAULT_SIZE_THRESHOLD
    else:
        threshold = args.size_threshold
    require(threshold is None or threshold >= MIN_SIZE_THRESHOLD, "size-threshold-too-small")
    binding = {"supervisor": supervisor, "coordinator": report["pin"]["session"], "size_threshold": threshold,
               "bound": handoff.utc_now().isoformat(),
               "replaced": {"supervisor_fingerprint": fingerprint(current["supervisor"]),
                            "coordinator": current["coordinator"]} if current else None}
    save(directory, "binding.json", binding)
    state = load(directory, "state.json", empty_state())
    recover_known_unsent(state)
    recover_pending(state)  # still uncertain and deduplicated: rebinding never retries them
    state["suspended"] = None
    save(directory, "state.json", state)
    return {"status": "bound", "supervisor_fingerprint": fingerprint(supervisor), "coordinator": binding["coordinator"]}


def cmd_unbind(root):
    directory = observer_dir(root)
    require(not verified_process(load(directory, "process.json")), "stop-the-observer-first")
    with contextlib.suppress(FileNotFoundError):
        os.unlink(directory / "binding.json")
    return {"status": "unbound"}


def cmd_smoke(root, label):
    require(label and len(label) <= 80 and re.fullmatch(r"[A-Za-z0-9 ._:-]+", label), "invalid-label")
    directory = observer_dir(root)
    binding = load(directory, "binding.json")
    require(binding is not None, "unbound-bind-first")
    state = load(directory, "state.json", empty_state())
    if recover_pending(state):
        save(directory, "state.json", state)
        raise Blocked("interrupted-delivery-uncertain-rebind-first")
    require(not state.get("smoke"), "smoke-already-delivered")
    require(not state.get("suspended"), "suspended-rebind-first")
    ok, reason, _ = verified_identity(root, binding)
    require(ok, reason or "identity-unverified")
    text = (f"OBSERVER SMOKE TEST ({label}): labelled, benign delivery check only. No reply, task action or approval "
            "is requested, and this notice carries no authority.")
    result = deliver(directory, state, binding, text, f"smoke:{label}", "smoke")  # persists the smoke intent first
    if state.get("smoke"):
        state["smoke"]["outcome"] = result["status"]
    save(directory, "state.json", state)
    last = state["deliveries"][-1]
    return {**result, "cli": last.get("cli"), "daemon": last.get("daemon"),
            "supervisor_fingerprint": fingerprint(binding["supervisor"])}


def cmd_run(root, interval):
    """Foreground loop for `start`; one per repository (exclusive lock)."""
    directory = observer_dir(root)
    lock = os.open(directory / "run.lock", os.O_RDWR | os.O_CREAT | os.O_NOFOLLOW, 0o600)
    try:
        fcntl.flock(lock, fcntl.LOCK_EX | fcntl.LOCK_NB)
    except BlockingIOError:
        return {"status": "already-running"}

    def stop(*_):
        global STOP
        STOP = True
    signal.signal(signal.SIGTERM, stop)
    signal.signal(signal.SIGINT, stop)
    while not STOP:
        try:
            print(json.dumps({"time": handoff.utc_now().isoformat(), **poll_once(root)}), flush=True)
        except Blocked as error:
            print(json.dumps({"status": "blocked", "reason": str(error)}), flush=True)
        except (OSError, ValueError, KeyError, subprocess.SubprocessError) as error:
            print(json.dumps({"status": "blocked", "reason": type(error).__name__}), flush=True)
        deadline = time.monotonic() + interval
        while not STOP and time.monotonic() < deadline:
            time.sleep(0.2)
    return {"status": "stopped"}


def parser():
    p = argparse.ArgumentParser(description="Coordinator milestone observer")
    p.add_argument("--repo", required=True)
    sub = p.add_subparsers(dest="command", required=True)
    bind = sub.add_parser("bind")
    target = bind.add_mutually_exclusive_group(required=True)
    target.add_argument("--supervisor")
    target.add_argument("--supervisor-file")
    bind.add_argument("--replace", action="store_true")
    size = bind.add_mutually_exclusive_group()
    size.add_argument("--size-threshold", type=int)
    size.add_argument("--no-size-advisory", action="store_true")
    for name in ("start", "run"):
        sp = sub.add_parser(name)
        sp.add_argument("--interval", type=int, default=DEFAULT_INTERVAL)
        if name == "start":
            sp.add_argument("--resume", action="store_true",
                            help="catch up from the saved cursor instead of starting at the current transcript tail")
    smoke = sub.add_parser("smoke")
    smoke.add_argument("--label", required=True)
    for name in ("stop", "status", "unbind", "poll-once"):
        sub.add_parser(name)
    return p


def main(argv=None):
    args = parser().parse_args(argv)
    root = handoff.repo_root(Path(args.repo))
    if args.command == "bind":
        return cmd_bind(root, args)
    if args.command == "start":
        return cmd_start(root, args.interval, args.resume)
    if args.command == "run":
        require(MIN_INTERVAL <= args.interval <= MAX_INTERVAL, "interval-out-of-range")
        return cmd_run(root, args.interval)
    if args.command == "stop":
        return cmd_stop(root)
    if args.command == "unbind":
        return cmd_unbind(root)
    if args.command == "smoke":
        return cmd_smoke(root, args.label)
    if args.command == "poll-once":
        return poll_once(root)
    return cmd_status(root)


if __name__ == "__main__":
    try:
        result = main()
        print(json.dumps(result, separators=(",", ":")))
        sys.exit(1 if result.get("status") in ("blocked", "suspended") else 0)
    except Blocked as error:
        print(json.dumps({"status": "blocked", "reason": str(error)}, separators=(",", ":")))
        sys.exit(1)
    except (OSError, ValueError, KeyError, TypeError, subprocess.SubprocessError):
        print(json.dumps({"status": "blocked", "reason": "local-input-process-or-cli-error"}))
        sys.exit(1)
